package analysis

import (
	"context"
	"fmt"
	"time"

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
		log:              log.WithField("label", label).WithField("clientName", clientName),
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

	metrics := postgresql.BlockMetricsModel{
		Slot:  int(slot),
		Label: b.label,
		Score: -1,
	}

	if slot > (phase0.Slot(b.CurrentHeadSlot) + utils.SlotsPerEpoch) {
		// beacon node is not synced
		b.Monitoring.ProposalStatus = 0
		log.Errorf("node is not synced(proposal slot: %d, node head slot: %d), not proposing", slot, b.CurrentHeadSlot)
		return
	}

	block, blockTime, err := b.Eth2Provider.ProposeNewBlock(slot, b.client)

	if err != nil {
		log.Errorf("error requesting block from %s: %s", b.label, err)
		b.Monitoring.ProposalStatus = 0

	} else {

		newMetrics, err := b.BlockMetrics(block, blockTime)
		if err != nil {
			log.Errorf("error analyzing block from %s: %s", b.label, err)
			b.Monitoring.ProposalStatus = 0
		} else {
			b.Monitoring.ProposalStatus = 1
			metrics = newMetrics
			log.Infof("Block Generation Time: %fs", blockTime.Seconds())
			log.Infof("Metrics: %+v", metrics)
		}

	}
	b.DBClient.PersisBlockScoreMetrics(metrics)

	if block != nil {
		b.PersistBlock(*block)
	}

	// We block the update attestations as new head could impact attestations of the proposed block
	// b.ProcessNewHead <- struct{}{} // Allow the new head to update attestations
}

func (b *ClientLiveData) GetLabel() string {
	return b.label
}

func (b *ClientLiveData) GetClient() string {
	return b.client
}

type MonitoringMetrics struct {
	ProposalStatus int
}
