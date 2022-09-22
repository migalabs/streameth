package chain_stats

import (
	"time"

	"github.com/attestantio/go-eth2-client/spec/phase0"
)

const (
	SLOT_DURATION = 12
)

type ChainTime struct {
	GenesisTime time.Time
}

// Calculate at which time a given slot happens
func (c ChainTime) SlotTime(slot phase0.Slot) time.Time {
	return c.GenesisTime.Add(time.Duration(slot) * SLOT_DURATION * time.Second)
}
