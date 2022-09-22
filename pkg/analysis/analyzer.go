package analysis

import (
	"context"
	"fmt"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/sirupsen/logrus"
	"github.com/tdahar/block-scorer/pkg/client_api"
)

var (
	moduleName = "Analysis"
	log        = logrus.WithField(
		"module", moduleName)
)

type BlockAnalyzer struct {
	ctx              context.Context
	Eth2Provider     client_api.APIClient                                       // connection to the beacon node
	AttHistory       map[phase0.Slot]map[phase0.CommitteeIndex]bitfield.Bitlist // 32 slots of attestation per slot and committeeIndex
	BlockRootHistory map[phase0.Slot]phase0.Root                                // 64 slots of roots
	log              *logrus.Entry                                              // each analyzer has its own logger
}

func NewBlockAnalyzer(ctx context.Context, label string, cliEndpoint string, timeout time.Duration) (*BlockAnalyzer, error) {
	client, err := client_api.NewAPIClient(ctx, label, cliEndpoint, timeout)
	if err != nil {
		log.Errorf("could not create eth2 client: %s", err)
		return &BlockAnalyzer{}, err
	}
	return &BlockAnalyzer{
		ctx:              ctx,
		Eth2Provider:     *client,
		AttHistory:       make(map[phase0.Slot]map[phase0.CommitteeIndex]bitfield.Bitlist),
		BlockRootHistory: make(map[phase0.Slot]phase0.Root),
		log:              log.WithField("label", label),
	}, nil
}

func (b *BlockAnalyzer) ProcessNewBlock(slot phase0.Slot) error {

	randaoReveal := phase0.BLSSignature{}
	graffiti := []byte("")
	snapshot := time.Now()
	block, err := b.Eth2Provider.Api.BeaconBlockProposal(b.ctx, slot, randaoReveal, graffiti)
	blockTime := time.Since(snapshot).Seconds() // time to ask for block
	if err != nil {
		return fmt.Errorf("error requesting block from %s: %s", b.Eth2Provider.Label, err)

	}
	for i := range b.AttHistory {
		if i+32 < slot { // attestations can only reference 32 slots back
			delete(b.AttHistory, i) // remove old entries from the map
		}
	}

	for i := range b.BlockRootHistory {
		if i+64 < slot { // attestations can only reference 32 slots back
			delete(b.BlockRootHistory, i) // remove old entries from the map
		}
	}

	// for now we just have Bellatrix
	metrics, err := b.BellatrixBlockMetrics(block.Bellatrix)
	if err != nil {
		return fmt.Errorf("error analyzing block from %s: %s", b.Eth2Provider.Label, err)
	}
	b.log.Infof("Block Duration: %f", blockTime)
	b.log.Infof("Metrics: %+v", metrics)

	return nil
}

type BeaconBlockMetrics struct {
	AttScore       float64
	SyncScore      float64
	AttestationNum int64
	NewVotes       int64
	CorrectSource  int64
	CorrectTarget  int64
	CorrectHead    int64
}
