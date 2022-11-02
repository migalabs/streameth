package analysis

import (
	"fmt"
	"strings"

	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/prysmaticlabs/go-bitfield"
)

func (b *ClientLiveData) UpdateAttestations(block bellatrix.BeaconBlock) {

	for _, attestation := range block.Body.Attestations {
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

func (b *ClientLiveData) BuildHistory() bool {
	log := b.log.WithField("routine", "history-build")
	currentHead, err := b.Eth2Provider.Api.BeaconBlockHeader(b.ctx, "head")
	headSlot := currentHead.Header.Message.Slot

	if err != nil {
		log.Panicf("could not retrieve current head: ", err)

	}

	if _, ok := b.BlockRootHistory[currentHead.Header.Message.Slot]; ok {
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
		block, err := b.Eth2Provider.Api.SignedBeaconBlock(b.ctx, fmt.Sprintf("%d", i))

		if err != nil {
			if strings.Contains(err.Error(), "Could not find requested block") {
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
		root, err := block.Root()
		if err != nil {
			log.Panicf("could not retrieve block root from block %d: ", i, err)
		}
		b.BlockRootHistory[i] = root

		if i+32 >= headSlot {
			b.UpdateAttestations(*block.Bellatrix.Message)
		}

	}

	return false
}
