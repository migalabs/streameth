package analysis

import (
	"encoding/hex"
	"time"

	api "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/tdahar/block-scorer/pkg/postgresql"
)

func (b *ClientLiveData) HandleHeadEvent(event *api.Event) {
	timestamp := time.Now()
	log := b.log.WithField("routine", "head-event")

	if event.Data == nil {
		return
	}

	data := event.Data.(*api.HeadEvent)
	log.Debugf("Received a new head event")

	newBlock, err := b.Eth2Provider.Api.SignedBeaconBlock(b.ctx, hex.EncodeToString(data.Block[:]))
	b.DBClient.InsertNewBlock(int(data.Slot), b.Eth2Provider.Label, timestamp)
	<-b.ProcessNewHead
	b.UpdateAttestations(*newBlock.Bellatrix.Message)

	if err != nil {
		log.Errorf("could not retrieve the head block: %s", err)
	}

}

func (b *ClientLiveData) HandleAttestationEvent(event *api.Event) {
	timestamp := time.Now()

	log := b.log.WithField("routine", "attestation-event")
	log.Debugf("Received a new event")

	if event.Data == nil {
		log.Errorf("attestation event does not contain anything")
	}

	data := event.Data.(*phase0.Attestation)

	log.Debugf("Initiating processing event in %f seconds", time.Since(timestamp).Seconds())
	beaconCommittee := b.EpochData.GetBeaconCommittee(uint64(data.Data.Slot), uint64(data.Data.Index))

	if beaconCommittee == nil {
		log.Errorf("could not retrieve beacon committee at slot %d", uint64(data.Data.Slot))
	}
	attestingVals := make([]phase0.ValidatorIndex, 0)

	for _, bit := range data.AggregationBits.BitIndices() {
		attestingVals = append(attestingVals, beaconCommittee[bit])
	}
	log.Debugf("Initiating writing tasks event in %f seconds", time.Since(timestamp).Seconds())
	// create params to be written
	baseParams := make([]interface{}, 0)
	baseParams = append(baseParams, b.Eth2Provider.Label)
	baseParams = append(baseParams, uint64(data.Data.Slot))
	baseParams = append(baseParams, uint64(data.Data.Index))
	baseParams = append(baseParams, timestamp)

	// for each attesting validator
	for _, item := range attestingVals {
		params := append(baseParams, uint64(item))
		writeTask := postgresql.WriteTask{
			QueryString: postgresql.InsertNewAtt,
			Params:      params,
		}
		b.DBClient.WriteChan <- writeTask // send task to be written
	}

	log.Debugf("Finished processing event in %f seconds", time.Since(timestamp).Seconds())

}
