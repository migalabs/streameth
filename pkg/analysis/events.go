package analysis

import (
	"encoding/hex"
	"fmt"
	"time"

	api_v1 "github.com/attestantio/go-eth2-client/api/v1"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/migalabs/streameth/pkg/postgresql"
	"github.com/migalabs/streameth/pkg/utils"
)

func (b *ClientLiveData) HandleHeadEvent(event *api_v1.Event) {
	timestamp := time.Now()
	log := b.log.WithField("routine", "head-event")

	if event.Data == nil {
		return
	}

	data := event.Data.(*api_v1.HeadEvent) // cast to head event
	log.Infof("Received a new event: slot %d", data.Slot)
	// <-b.ProcessNewHead // wait for the block proposal to be done
	// we only receive the block hash, get the new block
	newBlock, err := b.Eth2Provider.Api.SignedBeaconBlock(b.ctx, &api.SignedBeaconBlockOpts{
		Block: fmt.Sprintf("%#x", data.Block),
	})

	if newBlock == nil {
		log.Errorf("the block is not available: %d", data.Slot)
		return
	}
	if err != nil || newBlock == nil {
		log.Errorf("could not request new block: %s", err)
		return
	}
	b.UpdateAttestations(*newBlock.Data) // now update the attestations with the new head block in the chain

	// Track if there is any missing slot
	if b.CurrentHeadSlot != 0 && // we are not at the beginning of the run
		data.Slot-phase0.Slot(b.CurrentHeadSlot) > 1 { // there a gap bigger than 1 with the new head
		for i := b.CurrentHeadSlot + 1; i < uint64(data.Slot); i++ {
			params := make([]interface{}, 0)
			params = append(params, i)
			params = append(params, b.label)
			writeTask := postgresql.WriteTask{
				QueryString: postgresql.InsertNewMissedBlock,
				Params:      params,
			}
			b.DBClient.WriteChan <- writeTask // store
		}
	}
	b.CurrentHeadSlot = uint64(data.Slot)
	if b.CurrentHeadSlot%utils.SlotsPerEpoch == (utils.SlotsPerEpoch - 1) {
		// last slot of epoch, prepare next
		//  epoch is the next slot divided by 32
		epoch := phase0.Epoch((b.CurrentHeadSlot + 1) / utils.SlotsPerEpoch)

		go b.ProcessEpochTasks(epoch)
	}
	params := make([]interface{}, 0)
	params = append(params, int(data.Slot))
	params = append(params, b.label)
	params = append(params, timestamp)
	writeTask := postgresql.WriteTask{
		QueryString: postgresql.InsertNewBlock,
		Params:      params,
	}
	b.DBClient.WriteChan <- writeTask // store

}

// When the node receives a new attestation, this function is tiggered
func (b *ClientLiveData) HandleAttestationEvent(event *api_v1.Event) {
	timestamp := time.Now()

	log := b.log.WithField("routine", "attestation-event")

	if event.Data == nil {
		log.Errorf("attestation event does not contain anything")
	}

	data := event.Data.(*phase0.Attestation) // cast
	log.Debugf("Received a new event: slot %d, committee: %d", uint64(data.Data.Slot), uint64(data.Data.Index))
	// With the beacon committee we can identify the attesting validators
	// Will not track this for now
	// beaconCommittee := b.EpochData.GetBeaconCommittee(uint64(data.Data.Slot), uint64(data.Data.Index))

	// if beaconCommittee == nil {
	// 	log.Errorf("could not retrieve beacon committee at slot %d", uint64(data.Data.Slot))
	// 	return
	// }
	// attestingVals := make([]phase0.ValidatorIndex, 0)

	// for _, bit := range data.AggregationBits.BitIndices() {
	// 	attestingVals = append(attestingVals, beaconCommittee[bit])
	// }

	// create params to be written, same for all validators (same attestation)
	baseParams := make([]interface{}, 0)
	baseParams = append(baseParams, b.label)
	baseParams = append(baseParams, uint64(data.Data.Slot))
	baseParams = append(baseParams, uint64(data.Data.Index))
	baseParams = append(baseParams, timestamp)
	baseParams = append(baseParams, hex.EncodeToString(data.Signature[:]))
	baseParams = append(baseParams, hex.EncodeToString(data.Data.Source.Root[:]))
	baseParams = append(baseParams, hex.EncodeToString(data.Data.Target.Root[:]))
	baseParams = append(baseParams, hex.EncodeToString(data.Data.BeaconBlockRoot[:]))

	// for each attesting validator, not use for now
	// for _, item := range attestingVals {
	// 	params := append(baseParams, uint64(item)) // append the validator index
	// 	writeTask := postgresql.WriteTask{
	// 		QueryString: postgresql.InsertNewAtt,
	// 		Params:      params,
	// 	}
	// 	b.DBClient.WriteChan <- writeTask // send task to be written
	// }

	writeTask := postgresql.WriteTask{
		QueryString: postgresql.InsertNewAtt,
		Params:      baseParams,
	}

	b.DBClient.WriteChan <- writeTask // send task to be written

	log.Tracef("Finished processing event in %f seconds", time.Since(timestamp).Seconds())

}

func (b *ClientLiveData) HandleReOrgEvent(event *api_v1.Event) {
	timestamp := time.Now()
	log := b.log.WithField("routine", "reorg-event")

	if event.Data == nil {
		return
	}

	data := event.Data.(*api_v1.ChainReorgEvent) // cast to head event
	log.Debugf("New Reorg Evenet")

	baseParams := make([]interface{}, 0)
	baseParams = append(baseParams, b.label)
	baseParams = append(baseParams, uint64(data.Slot))
	baseParams = append(baseParams, hex.EncodeToString(data.OldHeadBlock[:]))
	baseParams = append(baseParams, hex.EncodeToString(data.NewHeadBlock[:]))
	baseParams = append(baseParams, uint64(data.Depth))
	baseParams = append(baseParams, timestamp)

	writeTask := postgresql.WriteTask{
		QueryString: postgresql.InsertNewReorg,
		Params:      baseParams,
	}

	b.DBClient.WriteChan <- writeTask // send task to be written

}

func (b *ClientLiveData) ProcessEpochTasks(epoch phase0.Epoch) {
	// retrieve duties
	duties, err := b.Eth2Provider.ProposerDuties(epoch)

	if err != nil {
		log.Errorf("could not process epoch %d tasks: %s", epoch, err)
	}
	vals := make([]phase0.ValidatorIndex, len(duties))

	for i, item := range duties {
		vals[i] = item.ValidatorIndex
	}

	log.Debugf("submitting proposal preparation...")
	err = b.Eth2Provider.SubmitProposalPreparation(vals)

	if err != nil {
		log.Errorf("could not submit a proposal peparation: %s", err)
	}
}
