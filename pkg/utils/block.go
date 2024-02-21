package utils

import (
	"encoding/hex"
	"fmt"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

const (
	InfinityRandaoReveal = "c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
	EmptyFeeRecipient    = "0x0000000000000000000000000000000000000000"
	SlotsPerEpoch        = 32
)

func CreateInfinityRandaoReveal() phase0.BLSSignature {
	// Infinity randao always required
	randaoReveal := phase0.BLSSignature{}
	bs, err := hex.DecodeString(InfinityRandaoReveal)
	if err == nil {
		copy(randaoReveal[:], bs)
	}

	return randaoReveal
}

func CreateEmptyFeeRecipient() bellatrix.ExecutionAddress {
	// Infinity randao always required
	feeRecipient := bellatrix.ExecutionAddress{}
	bs, err := hex.DecodeString(EmptyFeeRecipient)
	if err == nil {
		copy(feeRecipient[:], bs)
	}

	return feeRecipient
}

func GraffitiFromString(graffitiStr string) [32]byte {
	graffiti := [32]byte{}
	copy(graffiti[:], graffitiStr)
	return graffiti

}

func BlockBodyFromProposal(block api.VersionedProposal) (*capella.BeaconBlockBody, error) {
	switch block.Version {

	case spec.DataVersionCapella:
		return block.Capella.Body, nil
	default:
		return nil, fmt.Errorf("could not figure out the BlockBody Fork Version: %s", block.Version.String())
	}
}

func BlockBodyFromVersionedBlock(block spec.VersionedSignedBeaconBlock) (*capella.BeaconBlockBody, error) {
	switch block.Version {

	case spec.DataVersionCapella:
		return block.Capella.Message.Body, nil
	default:
		return nil, fmt.Errorf("could not figure out the Verioned Block Body Fork Version: %s", block.Version.String())
	}
}

func BlockToSSZ(block api.VersionedProposal) ([]byte, error) {
	switch block.Version.String() {

	case spec.DataVersionCapella.String():
		return block.Capella.MarshalSSZ()
	default:
		return []byte{}, fmt.Errorf("could not figure out data version")
	}

}
