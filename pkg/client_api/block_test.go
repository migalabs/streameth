package client_api

import (
	"context"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec/capella"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/migalabs/streameth/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestBlockSSZRead(t *testing.T) {

	cli, err := NewAPIClient(context.Background(), "test", "http://localhost:5052", 15*time.Second)

	if err != nil {
		t.Errorf("could not create cli: %s", err)
		return
	}

	head, err := cli.Api.BeaconBlockHeader(cli.ctx, &api.BeaconBlockHeaderOpts{
		Block: "head",
	})
	proposeSlot := head.Data.Header.Message.Slot + 1

	if err != nil {
		t.Errorf("could not download the block at the head")
		return
	}

	// Infinity randao always required
	randaoReveal := phase0.BLSSignature{}
	bs, err := hex.DecodeString(utils.InfinityRandaoReveal)
	if err == nil {
		copy(randaoReveal[:], bs)
	}

	proposeBlock, err := cli.Api.Proposal(cli.ctx, &api.ProposalOpts{
		Slot:                   head.Data.Header.Message.Slot + 1,
		RandaoReveal:           utils.CreateInfinityRandaoReveal(),
		Graffiti:               utils.GraffitiFromString(""),
		SkipRandaoVerification: true,
	})

	if err != nil {
		t.Errorf("could not propose the block at the head: %s", err)
		return
	}

	blockPath := "./test_block.ssz"

	blockWBytes, err := utils.BlockToSSZ(*proposeBlock.Data)
	if err != nil {
		log.Errorf("could not export block %d to ssz: %s", head.Data.Header.Message.Slot, err)
	}

	err = os.WriteFile(blockPath, blockWBytes, 0644)
	if err != nil {
		log.Errorf("error writing block to %s: %s", blockPath, err)
		return
	}

	log.Debugf("wrote %d bytes to %s", len(blockWBytes), blockPath)

	blockRBytes, err := os.ReadFile(blockPath)
	if err != nil {
		t.Errorf("could not read file: %s", blockPath)
		return
	}

	block := &capella.BeaconBlock{}
	err = block.UnmarshalSSZ(blockRBytes)

	if err != nil {
		t.Errorf("could not read slot from block: %s", err)
		return
	}

	slot := block.Slot

	assert.Equal(t, proposeSlot, slot)

	e := os.Remove(blockPath)
	if e != nil {
		log.Fatal(e)
	}
}
