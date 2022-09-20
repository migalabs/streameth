package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/sirupsen/logrus"
	"github.com/tdahar/block-scorer/pkg/chain_stats"
	"github.com/tdahar/block-scorer/pkg/client_api"
)

var (
	modName = "App"
	log     = logrus.WithField(
		"module", modName,
	)
)

type AppService struct {
	ctx       context.Context
	Clients   []*client_api.APIClient
	initTime  time.Time
	ChainTime chain_stats.ChainTime
	HeadSlot  phase0.Slot
}

func NewAppService(ctx context.Context, bnEndpoints []string) (*AppService, error) {

	clients := make([]*client_api.APIClient, 0)

	for i := range bnEndpoints {
		if !strings.Contains(bnEndpoints[i], "/") {
			log.Errorf("incorrect format for endpoint: %s", bnEndpoints[i])
		}
		label := strings.Split(bnEndpoints[i], "/")[0]
		endpoint := strings.Split(bnEndpoints[i], "/")[1]
		newClient, err := client_api.NewAPIClient(ctx, label, endpoint, time.Second*5)
		if err != nil {
			log.Errorf("could not create client for endpoint: %s ", endpoint, err)
			continue
		}
		clients = append(clients, newClient)
	}

	genesis, err := clients[0].Api.GenesisTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not obtain genesis time: %s", err)
	}
	headHeader, err := clients[0].Api.BeaconBlockHeader(ctx, "head")
	if err != nil {
		return nil, fmt.Errorf("could not obtain head block header: %s", err)
	}
	return &AppService{
		ctx:      ctx,
		Clients:  clients,
		initTime: time.Now(),
		HeadSlot: headHeader.Header.Message.Slot,
		ChainTime: chain_stats.ChainTime{
			GenesisTime: genesis,
		},
	}, nil
}

func (s *AppService) Run() {

	ticker := time.After(time.Until(s.ChainTime.SlotTime(phase0.Slot(s.HeadSlot + 1))))

	for {

		select {
		case <-s.ctx.Done():

			return

		case <-ticker:

			s.HeadSlot++
			log.Infof("Entered a new slot!: %d, time: %s", s.HeadSlot, time.Now())
			ticker = time.After(time.Until(s.ChainTime.SlotTime(phase0.Slot(s.HeadSlot + 1))))
			// a new slot has begun, therefore execute all needed actions
			log.Infof("Next Duration: %d", time.Until(s.ChainTime.SlotTime(phase0.Slot(s.HeadSlot+1))))
		default:
		}
	}
}
