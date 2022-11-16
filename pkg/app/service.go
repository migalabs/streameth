package app

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/sirupsen/logrus"
	"github.com/tdahar/eth-cl-live-metrics/pkg/analysis"
	"github.com/tdahar/eth-cl-live-metrics/pkg/chain_stats"
	"github.com/tdahar/eth-cl-live-metrics/pkg/postgresql"
)

var (
	modName = "Main App"
	log     = logrus.WithField(
		"module", modName,
	)
	attestationMetric = "attestations"
	proposalMetric    = "proposals"
	reorgMetric       = "reorgs"
)

type AppService struct {
	ctx         context.Context
	cancel      context.CancelFunc
	Analyzers   []*analysis.ClientLiveData
	initTime    time.Time
	ChainTime   chain_stats.ChainTime
	HeadSlot    phase0.Slot
	Metrics     []string
	finishTasks int32
	DBClient    *postgresql.PostgresDBService
}

func NewAppService(pCtx context.Context,
	bnEndpoints []string,
	dbEndpooint string,
	dbWorkers int,
	metrics []string) (*AppService, error) {

	ctx, cancel := context.WithCancel(pCtx)
	batchLen := len(bnEndpoints)
	for _, item := range metrics {
		if item == attestationMetric {
			batchLen = 100
		}
	}

	dbClient, err := postgresql.ConnectToDB(ctx, dbEndpooint, dbWorkers, batchLen)

	if err != nil {
		log.Panicf("could not connect to database: %s", err)
	}

	analyzers := make([]*analysis.ClientLiveData, 0) // one analyzer per beacon node

	for i := range bnEndpoints {
		// parse each beacon node endpoint
		if !strings.Contains(bnEndpoints[i], "/") {
			log.Errorf("incorrect format for endpoint: %s", bnEndpoints[i])
		}
		// label := strings.Split(bnEndpoints[i], "/")[0]
		endpoint := strings.Split(bnEndpoints[i], "/")[1]
		// newAnalyzer, err := analysis.NewBlockAnalyzer(ctx, label, endpoint, time.Second*5)
		newAnalyzer, err := analysis.NewBlockAnalyzer(ctx, bnEndpoints[i], endpoint, time.Second*5, dbClient)

		if err != nil {
			log.Errorf("could not create client for endpoint: %s ", endpoint, err)
			continue
		}
		analyzers = append(analyzers, newAnalyzer)
	}
	// get genesis time to calculate each slot time
	// Keep in mind first endpoint will be used as master
	genesis, err := analyzers[0].Eth2Provider.Api.GenesisTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not obtain genesis time: %s", err)
	}
	// check the current chain head
	headHeader, err := analyzers[0].Eth2Provider.Api.BeaconBlockHeader(ctx, "head")
	if err != nil {
		return nil, fmt.Errorf("could not obtain head block header: %s", err)
	}
	return &AppService{
		ctx:       ctx,
		cancel:    cancel,
		Analyzers: analyzers,
		initTime:  time.Now(),
		HeadSlot:  headHeader.Header.Message.Slot,
		ChainTime: chain_stats.ChainTime{
			GenesisTime: genesis,
		},
		Metrics:  metrics,
		DBClient: dbClient,
	}, nil
}

// Main routine: build block history and block proposals every 12 seconds
func (s *AppService) Run() {
	defer s.cancel()
	var wg sync.WaitGroup
	for _, item := range s.Metrics {
		if item == attestationMetric {
			log.Infof("initiating attestation events monitoring")
			wg.Add(1)
			s.RunAttestations()
		}

		if item == reorgMetric {
			wg.Add(1)
			s.RunReOrgs()
		}

		if item == proposalMetric {
			log.Infof("initiating block proposal monitoring")
			wg.Add(1)
			go s.RunMainRoutine(&wg)
		}
	}

	wg.Wait()

}

// Main routine: build block history and block proposals every 12 seconds
func (s *AppService) RunAttestations() {

	// Subscribe to events from each client
	for _, item := range s.Analyzers {
		err := item.Eth2Provider.Api.Events(s.ctx, []string{"attestation"}, item.HandleAttestationEvent) // every new head
		if err != nil {
			log.Panicf("failed to subscribe to head events: %s, label: %s", err, item.Eth2Provider.Label)
		}

	}
}

// Main routine: build block history and block proposals every 12 seconds
func (s *AppService) RunReOrgs() {

	// Subscribe to events from each client
	for _, item := range s.Analyzers {
		err := item.Eth2Provider.Api.Events(s.ctx, []string{"chain_reorg"}, item.HandleAttestationEvent) // every new head
		if err != nil {
			log.Panicf("failed to subscribe to reorg events: %s, label: %s", err, item.Eth2Provider.Label)
		}

	}
}

// Main routine: build block history and block proposals every 12 seconds
func (s *AppService) RunMainRoutine(wg *sync.WaitGroup) {
	defer wg.Done()
	log = log.WithField("routine", "main")
	historyBuilt := false

	for !historyBuilt {
		historyBuilt = true
		for _, item := range s.Analyzers {
			ok := item.BuildHistory()
			if !ok {
				historyBuilt = false
			}
		}
	}

	// Subscribe to events from each client
	for _, item := range s.Analyzers {
		err := item.Eth2Provider.Api.Events(s.ctx, []string{"head"}, item.HandleHeadEvent) // every new head
		if err != nil {
			log.Panicf("failed to subscribe to head events: %s, label: %s", err, item.Eth2Provider.Label)
		}

	}

	// tick every slot start (12 seconds)
	ticker := time.After(time.Until(s.ChainTime.SlotTime(phase0.Slot(s.HeadSlot + 1))))
loop:
	for {

		if s.finishTasks > 0 {
			log.Infof("closing main routine")
			s.DBClient.DoneTasks() // all the analyzers have the same db client
			break loop
		}
		select {
		case <-s.ctx.Done():
			log.Infof("closing main routine")
			s.DBClient.DoneTasks() // all the analyzers have the same db client
			break loop

		case <-ticker:
			// we entered a new slot time
			s.HeadSlot++
			log.Infof("Entered a new slot!: %d, time: %s", s.HeadSlot, time.Now())
			// reset ticker to next slot
			ticker = time.After(time.Until(s.ChainTime.SlotTime(phase0.Slot(s.HeadSlot + 1))))
			// a new slot has begun, therefore execute all needed actions
			log.Tracef("Time until next slot tick: %s", time.Until(s.ChainTime.SlotTime(phase0.Slot(s.HeadSlot+1))).String())
			for _, analyzer := range s.Analyzers {
				// for each beacon node, get a new block and analyze it
				go analyzer.ProposeNewBlock(s.HeadSlot)
			}

		default:
		}
	}
	log.Infof("finished")
}

func (s *AppService) Close() {
	log.Info("Sudden closed detected, closing Live Metrics")
	atomic.AddInt32(&s.finishTasks, int32(1))
	s.DBClient.WgDBWriter.Wait()
	s.cancel()
}
