package analysis

import (
	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/tdahar/block-scorer/pkg/postgresql"
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

func (b *ClientLiveData) BellatrixBlockMetrics(block *bellatrix.BeaconBlock) (postgresql.BlockMetricsModel, error) {
	// log := b.log.WithField("task", "bellatrix-block-score") // add extra log for function
	totalNewVotes := 0
	totalScore := 0
	attested := make(map[phase0.Slot]map[phase0.CommitteeIndex]bitfield.Bitlist) // for current block
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
			if b.AttHistory[slot][committeIndex].BitAt(uint64(idx)) {
				// already registered vote in a previous block
				continue
			}
			if attested[slot][committeIndex].BitAt(uint64(idx)) {
				// already registered vote in a same block
				continue
			}
			// we do not touch the history, as this is just a proposed block, but we do not know if it will be included in the chain
			attested[slot][committeIndex].SetBitAt(uint64(idx), true) // register as attested in the current block
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

	// syncCommitteeScore := float64(block.Body.SyncAggregate.SyncCommitteeBits.Count()) * float64(SYNC_REWARD_WEIGHT) / float64(WEIGHT_DENOMINATOR)

	return postgresql.BlockMetricsModel{
		Slot:          int(block.Slot),
		Label:         b.Eth2Provider.Label,
		CorrectSource: 0,
		CorrectTarget: 0,
		CorrectHead:   0,
		Score:         float64(totalScore),
		NewVotes:      totalNewVotes,
	}, nil
}
