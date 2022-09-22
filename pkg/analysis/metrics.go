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
	PROPOSER_WEIGHT      = 8
	WEIGHT_DENOMINATOR   = 64
	SYNC_REWARD_WEIGHT   = 2
)

func (b *BlockAnalyzer) BellatrixBlockMetrics(block *bellatrix.BeaconBlock) (BeaconBlockMetrics, error) {
	// log = b.log.WithField("method", "bellatrix-block") // add extra log for function
	totalNewVotes := 0
	totalScore := 0

	for _, attestation := range block.Body.Attestations {
		newVotes := 0
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
			newVotes++
		}
		score := 0
		if utils.IsCorrectSource(*attestation, *block) {
			score += newVotes * TIMELY_SOURCE_WEIGHT
		}
		if utils.IsCorrectTarget(*attestation, *block, b.BlockRootHistory) {
			score += newVotes * TIMELY_TARGET_WEIGHT
		}
		if utils.IsCorrectHead(*attestation, *block) {
			score += newVotes * TIMELY_HEAD_WEIGHT
		}

		totalNewVotes += newVotes
		// denominator := (WEIGHT_DENOMINATOR - PROPOSER_WEIGHT) * WEIGHT_DENOMINATOR / PROPOSER_WEIGHT
		totalScore += score / WEIGHT_DENOMINATOR

	}

	syncCommitteeScore := float64(block.Body.SyncAggregate.SyncCommitteeBits.Count()) * float64(SYNC_REWARD_WEIGHT) / float64(WEIGHT_DENOMINATOR)

	return BeaconBlockMetrics{
		AttScore:  float64(totalScore),
		SyncScore: float64(syncCommitteeScore),
	}, nil
}
