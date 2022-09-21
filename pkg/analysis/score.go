package analysis

import (
	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/tdahar/block-scorer/pkg/utils"
)

const (
	TIMELY_SOURCE_WEIGHT = 14
	TIMELY_TARGET_WEIGHT = 26
	TIMELY_HEAD_WEIGHT   = 14
)

func BellatrixBlockMetrics(block *bellatrix.BeaconBlock) (BeaconBlockMetrics, error) {
	totalNewVotes := 0
	totalScore := 0
	// Map is attestation slot -> committee index -> validator committee index -> aggregate.
	attested := make(map[phase0.Slot]map[phase0.CommitteeIndex]bitfield.Bitlist)
	for _, attestation := range block.Body.Attestations {
		newVotes := 0
		slot := attestation.Data.Slot

		if _, exists := attested[slot]; !exists {
			// add slot to map
			attested[slot] = make(map[phase0.CommitteeIndex]bitfield.Bitlist)
		}

		committeIndex := attestation.Data.Index
		if _, exists := attested[slot][committeIndex]; !exists {
			attested[slot][committeIndex] = bitfield.NewBitlist(attestation.AggregationBits.Len())
		}

		attestingIndices := attestation.AggregationBits.BitIndices()

		for _, idx := range attestingIndices {
			if attested[slot][committeIndex].BitAt(uint64(idx)) {
				// already registered vote
				continue
			}
			attested[slot][committeIndex].SetBitAt(uint64(idx), true)
			newVotes++
		}
		score := 0
		if utils.IsCorrectHead(*attestation, *block) {
			score += newVotes * TIMELY_HEAD_WEIGHT
		}

		if utils.IsCorrectTarget(*attestation, *block) {
			score += newVotes * TIMELY_TARGET_WEIGHT
		}

		totalNewVotes += newVotes
		totalScore += score

	}

	return BeaconBlockMetrics{
		Score: float64(totalScore),
	}, nil
}
