package analysis

import (
	"context"
	"fmt"
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/sirupsen/logrus"
	"github.com/tdahar/eth-cl-live-metrics/pkg/analysis/additional_structs"
	"github.com/tdahar/eth-cl-live-metrics/pkg/client_api"
	"github.com/tdahar/eth-cl-live-metrics/pkg/postgresql"
)

var (
	moduleName = "Analysis"
	log        = logrus.WithField(
		"module", moduleName)
)

type ClientLiveData struct {
	ctx              context.Context
	Eth2Provider     client_api.APIClient                                       // connection to the beacon node
	AttHistory       map[phase0.Slot]map[phase0.CommitteeIndex]bitfield.Bitlist // 32 slots of attestation per slot and committeeIndex
	BlockRootHistory map[phase0.Slot]phase0.Root                                // 64 slots of roots
	log              *logrus.Entry                                              // each analyzer has its own logger
	ProcessNewHead   chan struct{}
	DBClient         *postgresql.PostgresDBService
	EpochData        additional_structs.EpochStructs
	CurrentHeadSlot  uint64
}

func NewBlockAnalyzer(ctx context.Context, label string, cliEndpoint string, timeout time.Duration, dbClient *postgresql.PostgresDBService) (*ClientLiveData, error) {
	client, err := client_api.NewAPIClient(ctx, label, cliEndpoint, timeout)
	if err != nil {
		log.Errorf("could not create eth2 client: %s", err)
		return &ClientLiveData{}, err
	}
	return &ClientLiveData{
		ctx:              ctx,
		Eth2Provider:     *client,
		DBClient:         dbClient,
		AttHistory:       make(map[phase0.Slot]map[phase0.CommitteeIndex]bitfield.Bitlist),
		BlockRootHistory: make(map[phase0.Slot]phase0.Root),
		log:              log.WithField("label", label),
		EpochData:        additional_structs.NewEpochData(client.Api),
		CurrentHeadSlot:  0,
	}, nil
}

// Asks for a block proposal to the client and stores score in the database
func (b *ClientLiveData) ProposeNewBlock(slot phase0.Slot) error {
	log := b.log.WithField("task", "generate-block")
	log.Debugf("processing new block: %d\n", slot)

	randaoReveal := phase0.BLSSignature{}
	graffiti := []byte("")
	snapshot := time.Now()
	block, err := b.Eth2Provider.Api.BeaconBlockProposal(b.ctx, slot, randaoReveal, graffiti) // ask for block proposal
	blockTime := time.Since(snapshot).Seconds()                                               // time to generate block

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
	metrics, err := b.BellatrixBlockMetrics(block.Bellatrix, blockTime)
	if err != nil {
		return fmt.Errorf("error analyzing block from %s: %s", b.Eth2Provider.Label, err)
	}
	log.Infof("Block Generation Time: %f", blockTime)
	log.Infof("Metrics: %+v", metrics)

	// Store in DB
	params := make([]interface{}, 0)
	params = append(params, metrics.Slot)
	params = append(params, metrics.Label)
	params = append(params, metrics.Score)
	params = append(params, metrics.Duration)
	params = append(params, metrics.CorrectSource)
	params = append(params, metrics.CorrectTarget)
	params = append(params, metrics.CorrectHead)
	params = append(params, metrics.Sync1Bits)
	params = append(params, metrics.AttNum)
	params = append(params, metrics.NewVotes)

	writeTask := postgresql.WriteTask{
		QueryString: postgresql.InsertNewScore,
		Params:      params,
	}

	b.DBClient.WriteChan <- writeTask

	// We block the update attestations as new head could impact attestations of the proposed block
	b.ProcessNewHead <- struct{}{} // Allow the new head to update attestations
	return nil
}
