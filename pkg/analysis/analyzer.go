package analysis

import (
	"context"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/migalabs/streameth/pkg/analysis/additional_structs"
	"github.com/migalabs/streameth/pkg/client_api"
	"github.com/migalabs/streameth/pkg/postgresql"
	"github.com/migalabs/streameth/pkg/utils"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/sirupsen/logrus"
)

var (
	moduleName = "Analysis"
	log        = logrus.WithField(
		"module", moduleName)
)

// TODO: make attributes private where possible
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
	Monitoring       MonitoringMetrics
	client           string
	label            string
	blocksDir        string
}

func NewBlockAnalyzer(
	ctx context.Context,
	clientName string,
	label string,
	cliEndpoint string,
	timeout time.Duration,
	dbClient *postgresql.PostgresDBService,
	blocksBaseDir string) (*ClientLiveData, error) {
	client, err := client_api.NewAPIClient(ctx, label, cliEndpoint, timeout)
	if err != nil {
		log.Errorf("could not create eth2 client: %s", err)
		return &ClientLiveData{}, err
	}

	if !utils.CheckValidClientName(clientName) {
		log.Errorf("could not identify eth2 client, try one of: Prysm,Lighthouse,Teku,Nimbus,Lodestar,Grandine")
		return &ClientLiveData{}, nil
	}

	analyzer := &ClientLiveData{
		ctx:              ctx,
		Eth2Provider:     *client,
		DBClient:         dbClient,
		AttHistory:       make(map[phase0.Slot]map[phase0.CommitteeIndex]bitfield.Bitlist),
		BlockRootHistory: make(map[phase0.Slot]phase0.Root),
		log:              log.WithField("label", label),
		EpochData:        additional_structs.NewEpochData(client.Api),
		CurrentHeadSlot:  0,
		ProcessNewHead:   make(chan struct{}),
		Monitoring:       MonitoringMetrics{},
		client:           clientName,
		blocksDir:        fmt.Sprintf("%s/%s/%s/", blocksBaseDir, label, clientName),
		label:            fmt.Sprintf("%s_%s", label, cliEndpoint),
	}
	analyzer.CheckBlocksFolder()

	return analyzer, nil
}

// Asks for a block proposal to the client and stores score in the database
func (b *ClientLiveData) ProposeNewBlock(slot phase0.Slot) {
	log := b.log.WithField("task", "generate-block")
	log.Debugf("processing new block: %d\n", slot)

	for i := range b.AttHistory { // TODO: 32 must be a constant
		if i+32 < slot { // attestations can only reference 32 slots back
			delete(b.AttHistory, i) // remove old entries from the map
		}
	}

	for i := range b.BlockRootHistory { // TODO: 64 must be a constant
		if i+64 < slot { // attestations can only reference 32 slots back
			delete(b.BlockRootHistory, i) // remove old entries from the map
		}
	}

	// Infinity randao always required
	randaoReveal := phase0.BLSSignature{}
	bs, err := hex.DecodeString(utils.InfinityRandaoReveal)
	if err == nil {
		copy(randaoReveal[:], bs)
	}
	skipRandaoVerification := false // only needed for Lighthouse, Nimbus and Grandine

	if b.client == utils.LighthouseClient ||
		b.client == utils.NimbusClient ||
		b.client == utils.GrandineClient {
		skipRandaoVerification = true
	}

	graffiti := make([]byte, 32) // TODO: 32 must be a constant

	proposalOpts := api.ProposalOpts{
		Slot:                   slot,
		RandaoReveal:           randaoReveal,
		Graffiti:               ([32]byte)(graffiti),
		SkipRandaoVerification: skipRandaoVerification,
	}

	metrics := postgresql.BlockMetricsModel{
		Slot:  int(slot),
		Label: b.label,
		Score: -1,
	}

	snapshot := time.Now()
	block, err := b.Eth2Provider.Api.Proposal(b.ctx, &proposalOpts) // ask for block proposal
	blockTime := time.Since(snapshot).Seconds()                     // time to generate block

	if err != nil {
		log.Errorf("error requesting block from %s: %s", b.label, err)
		b.Monitoring.ProposalStatus = 0

	} else {

		// for now we just have Capella
		newMetrics, err := b.BlockMetrics(block.Data, blockTime)
		if err != nil {
			log.Errorf("error analyzing block from %s: %s", b.label, err)
			b.Monitoring.ProposalStatus = 0
		} else {
			b.Monitoring.ProposalStatus = 1
			metrics = newMetrics
			log.Infof("Block Generation Time: %f", blockTime)
			log.Infof("Metrics: %+v", metrics)
		}

	}

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
	params = append(params, metrics.AttesterSlashings)
	params = append(params, metrics.ProposerSlashings)
	params = append(params, metrics.ProposerSlashingScore)
	params = append(params, metrics.AttesterSlashingScore)
	params = append(params, metrics.SyncScore)

	writeTask := postgresql.WriteTask{
		QueryString: postgresql.InsertNewScore,
		Params:      params,
	}
	b.Monitoring.ProposalStatus = 1

	b.DBClient.WriteChan <- writeTask

	if block != nil {
		b.PersistBlock(*block.Data)
	}

	// We block the update attestations as new head could impact attestations of the proposed block
	// b.ProcessNewHead <- struct{}{} // Allow the new head to update attestations
}

func (b ClientLiveData) GetLabel() string {
	return b.label
}

type MonitoringMetrics struct {
	ProposalStatus int
}
