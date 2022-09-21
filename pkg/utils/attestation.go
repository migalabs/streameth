package utils

import (
	"bytes"

	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

func IsCorrectSource(attestation phase0.Attestation, block bellatrix.BeaconBlock) bool {
	return true
}

func IsCorrectTarget(attestation phase0.Attestation, block bellatrix.BeaconBlock) bool {

	return true

}

func IsCorrectHead(attestation phase0.Attestation, block bellatrix.BeaconBlock) bool {
	if bytes.Equal(block.ParentRoot[:], attestation.Data.BeaconBlockRoot[:]) {
		if block.Slot-attestation.Data.Slot == 1 {
			return true
		}
	}
	return false
}
