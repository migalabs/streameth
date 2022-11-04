package analysis

import (
	"sort"

	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/prysmaticlabs/go-bitfield"
	"github.com/tdahar/eth-cl-live-metrics/pkg/postgresql"
	"github.com/tdahar/eth-cl-live-metrics/pkg/utils"
)

const (
	TIMELY_SOURCE_WEIGHT = 14
	TIMELY_TARGET_WEIGHT = 26
	TIMELY_HEAD_WEIGHT   = 14
	PROPOSER_WEIGHT      = 8
	WEIGHT_DENOMINATOR   = 64
	SYNC_REWARD_WEIGHT   = 2
)

// https://github.com/attestantio/vouch/blob/0c75ee8315dc4e5df85eb2aa09b4acc2b4436661/strategies/beaconblockproposal/best/score.go#L222
// This function receives a new block proposal and ouputs a block score and metrics about the block
func (b *ClientLiveData) BellatrixBlockMetrics(block *bellatrix.BeaconBlock, duration float64) (postgresql.BlockMetricsModel, error) {
	// log := b.log.WithField("task", "bellatrix-block-score") // add extra log for function
	totalNewVotes := 0
	totalScore := float64(0)
	attScore := float64(0)
	totalCorrectSource := 0
	totalCorrectTarget := 0
	totalCorrectHead := 0
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
			totalCorrectSource += newVotes
		}
		if utils.IsCorrectTarget(*attestation, *block, b.BlockRootHistory) {
			score += newVotes * TIMELY_TARGET_WEIGHT
			totalCorrectTarget += newVotes
		}
		if utils.IsCorrectHead(*attestation, *block) {
			score += newVotes * TIMELY_HEAD_WEIGHT
			totalCorrectHead += newVotes
		}

		totalNewVotes += newVotes
		// denominator := (WEIGHT_DENOMINATOR - PROPOSER_WEIGHT) * WEIGHT_DENOMINATOR / PROPOSER_WEIGHT
		attScore += float64(score) / float64(WEIGHT_DENOMINATOR)

	}

	syncCommitteeScore := float64(block.Body.SyncAggregate.SyncCommitteeBits.Count()) * float64(SYNC_REWARD_WEIGHT) / float64(WEIGHT_DENOMINATOR)

	attesterSlashingScore, proposerSlashingScore := scoreSlashings(block.Body.AttesterSlashings, block.Body.ProposerSlashings)

	totalScore = attScore + syncCommitteeScore + attesterSlashingScore + proposerSlashingScore

	return postgresql.BlockMetricsModel{
		Slot:                  int(block.Slot),
		Label:                 b.Eth2Provider.Label,
		CorrectSource:         totalCorrectSource,
		CorrectTarget:         totalCorrectTarget,
		CorrectHead:           totalCorrectHead,
		Score:                 float64(totalScore),
		Duration:              duration,
		NewVotes:              totalNewVotes,
		AttNum:                len(block.Body.Attestations),
		Sync1Bits:             int(block.Body.SyncAggregate.SyncCommitteeBits.Count()),
		AttesterSlashings:     len(block.Body.AttesterSlashings),
		ProposerSlashings:     len(block.Body.ProposerSlashings),
		ProposerSlashingScore: proposerSlashingScore,
		AttesterSlashingScore: attScore,
		SyncScore:             syncCommitteeScore,
	}, nil
}

// https://github.com/attestantio/vouch/blob/0c75ee8315dc4e5df85eb2aa09b4acc2b4436661/strategies/beaconblockproposal/best/score.go#L312
func scoreSlashings(attesterSlashings []*phase0.AttesterSlashing,
	proposerSlashings []*phase0.ProposerSlashing,
) (float64, float64) {
	// Slashing reward will be at most MAX_EFFECTIVE_BALANCE/WHISTLEBLOWER_REWARD_QUOTIENT,
	// which is 0.0625 Ether.
	// Individual attestation reward at 250K validators will be around 23,000 GWei, or .000023 Ether.
	// So we state that a single slashing event has the same weight as about 2,700 attestations.
	slashingWeight := float64(2700)

	// Add proposer slashing scores.
	proposerSlashingScore := float64(len(proposerSlashings)) * slashingWeight

	// Add attester slashing scores.
	indicesSlashed := 0
	for _, slashing := range attesterSlashings {
		indicesSlashed += len(intersection(slashing.Attestation1.AttestingIndices, slashing.Attestation2.AttestingIndices))
	}
	attesterSlashingScore := slashingWeight * float64(indicesSlashed)

	return attesterSlashingScore, proposerSlashingScore
}

// https://github.com/attestantio/vouch/blob/0c75ee8315dc4e5df85eb2aa09b4acc2b4436661/strategies/beaconblockproposal/best/score.go#L426
// intersection returns a list of items common between the two sets.
func intersection(set1 []uint64, set2 []uint64) []uint64 {
	sort.Slice(set1, func(i, j int) bool { return set1[i] < set1[j] })
	sort.Slice(set2, func(i, j int) bool { return set2[i] < set2[j] })
	res := make([]uint64, 0)

	set1Pos := 0
	set2Pos := 0
	for set1Pos < len(set1) && set2Pos < len(set2) {
		switch {
		case set1[set1Pos] < set2[set2Pos]:
			set1Pos++
		case set2[set2Pos] < set1[set1Pos]:
			set2Pos++
		default:
			res = append(res, set1[set1Pos])
			set1Pos++
			set2Pos++
		}
	}

	return res
}
