package analysis

import (
	"fmt"
	"strings"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/migalabs/streameth/pkg/utils"
	"github.com/prysmaticlabs/go-bitfield"
)

// This method receives a new head block and updates the attestation in the history
// So, when a block is to be proposed, we can check the history to identify new votes
func (b *ClientLiveData) UpdateAttestations(block spec.VersionedSignedBeaconBlock) {

	slot, err := block.Slot()

	if err != nil {
		log.Errorf("could not get block slot from block proposal: %s", err)
	}

	blockBody, err := utils.BlockBodyFromVersionedBlock(block)
	if err != nil {
		log.Errorf("could not get block body from block proposal: %s", err)
	}

	log.Tracef("updating attestations using block: %d", slot)

	for _, attestation := range blockBody.Attestations {
		slot := attestation.Data.Slot

		if _, exists := b.AttHistory[slot]; !exists {
			// add slot to map
			b.AttHistory[slot] = make(map[phase0.CommitteeIndex]bitfield.Bitlist)
		}

		committeIndex := attestation.Data.Index
		if _, exists := b.AttHistory[slot][committeIndex]; !exists {
			b.AttHistory[slot][committeIndex] = bitfield.NewBitlist(attestation.AggregationBits.Len())
		}

		attestingIndices := attestation.AggregationBits.BitIndices()

		for _, idx := range attestingIndices {
			if b.AttHistory[slot][committeIndex].BitAt(uint64(idx)) {
				// already registered vote
				continue
			}
			b.AttHistory[slot][committeIndex].SetBitAt(uint64(idx), true)
		}

	}

}

// This function is only called at the beginning of the run, so we build an initial attestation history
// to judge new block proposals
func (b *ClientLiveData) BuildHistory() bool {
	log := b.log.WithField("routine", "history-build")

	headOpts := api.BeaconBlockHeaderOpts{
		Block: "head",
	}

	currentHead, err := b.Eth2Provider.Api.BeaconBlockHeader(b.ctx, &headOpts)

	if err != nil {
		log.Panicf("could not retrieve current head: ", err)
	}
	headSlot := currentHead.Data.Header.Message.Slot

	if _, ok := b.BlockRootHistory[headSlot]; ok {
		// at this point we have already filled the historical records
		return true
	}

	// at this point there are no historical records

	for i := headSlot; i >= headSlot-64; i-- {
		if _, ok := b.BlockRootHistory[i]; ok {
			// at this point we have already filled the historical records
			return false // but we had to fill something
		}

		log.Debugf("filling block history, slot: %d\n", i)
		block, err := b.Eth2Provider.Api.SignedBeaconBlock(b.ctx, &api.SignedBeaconBlockOpts{
			Block: fmt.Sprintf("%d", i),
		})

		if err != nil {
			if strings.Contains(err.Error(), "404") {
				log.Debugf("Missed block!")
				continue
			} else {
				log.Panicf("could not retrieve historical block at slot: %d: ", i, err)
			}
		}
		if block == nil {
			log.Debugf("Unknown error with block %d\n", i)
			continue
		}
		root, err := block.Data.Root()
		if err != nil {
			log.Panicf("could not retrieve block root from block %d: ", i, err)
		}
		b.BlockRootHistory[i] = root

		if i+32 >= headSlot {
			b.UpdateAttestations(*block.Data)
		}

	}

	return false
}
