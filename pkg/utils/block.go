package utils

import (
	"fmt"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec"
	"github.com/attestantio/go-eth2-client/spec/capella"
)

const (
	InfinityRandaoReveal = "c00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
)

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
