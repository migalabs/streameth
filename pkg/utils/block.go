package utils

import (
	"encoding/hex"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/bellatrix"
	"github.com/attestantio/go-eth2-client/spec/deneb"
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

func BlockBodyFromProposal(block api.VersionedProposal) deneb.BeaconBlockBody {

	return *block.Deneb.Block.Body
}

func BlockBodyFromVersionedBlock(block spec.VersionedSignedBeaconBlock) deneb.BeaconBlockBody {
	return *block.Deneb.Message.Body
}

func BlockToSSZ(block api.VersionedProposal) ([]byte, error) {

	return block.Deneb.MarshalSSZ()

}
