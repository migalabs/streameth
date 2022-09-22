package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/sirupsen/logrus"
	"github.com/tdahar/block-scorer/pkg/analysis"
	"github.com/tdahar/block-scorer/pkg/chain_stats"
)

var (
	modName = "Main App"
	log     = logrus.WithField(
		"module", modName,
	)
)

type AppService struct {
	ctx       context.Context
	Analyzers []*analysis.BlockAnalyzer
	initTime  time.Time
	ChainTime chain_stats.ChainTime
	HeadSlot  phase0.Slot
}

func NewAppService(ctx context.Context, bnEndpoints []string) (*AppService, error) {

	analyzers := make([]*analysis.BlockAnalyzer, 0) // one analyzer per beacon node

	for i := range bnEndpoints {
		// parse each beacon node endpoint
		if !strings.Contains(bnEndpoints[i], "/") {
			log.Errorf("incorrect format for endpoint: %s", bnEndpoints[i])
		}
		label := strings.Split(bnEndpoints[i], "/")[0]
		endpoint := strings.Split(bnEndpoints[i], "/")[1]
		newAnalyzer, err := analysis.NewBlockAnalyzer(ctx, label, endpoint, time.Second*5)
		if err != nil {
			log.Errorf("could not create client for endpoint: %s ", endpoint, err)
			continue
		}
		analyzers = append(analyzers, newAnalyzer)
	}
	// get genesis time to calculate each slot time
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
		Analyzers: analyzers,
		initTime:  time.Now(),
		HeadSlot:  headHeader.Header.Message.Slot, // start 64 slots behind to create attestation history
		ChainTime: chain_stats.ChainTime{
			GenesisTime: genesis,
		},
	}, nil
}

// Main routine
func (s *AppService) Run() {
	log = log.WithField("routine", "main")

	// tick every slot start (12 seconds)
	ticker := time.After(time.Until(s.ChainTime.SlotTime(phase0.Slot(s.HeadSlot + 1))))

	for {

		select {
		case <-s.ctx.Done():

			return

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
				go analyzer.ProcessNewBlock(s.HeadSlot)
			}
		default:
		}
	}
}
