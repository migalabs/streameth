package client_api

import (
	"fmt"
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/migalabs/streameth/pkg/utils"
)

// Asks for a block proposal to the client and stores score in the database
func (b *APIClient) ProposeNewBlock(slot phase0.Slot, client string) (api.VersionedProposal, time.Duration, error) {
	log.Debugf("proposing new block: %d\n", slot)

	skipRandaoVerification := false // only needed for Lighthouse, Nimbus and Grandine

	if client == utils.LighthouseClient ||
		client == utils.NimbusClient ||
		client == utils.GrandineClient {
		skipRandaoVerification = true
	}

	proposalOpts := api.ProposalOpts{
		Slot:                   slot,
		RandaoReveal:           utils.CreateInfinityRandaoReveal(),
		Graffiti:               utils.GraffitiFromString(""),
		SkipRandaoVerification: skipRandaoVerification,
	}

	snapshot := time.Now()
	block, err := b.Api.Proposal(b.ctx, &proposalOpts) // ask for block proposal
	blockTime := time.Since(snapshot)                  // time to generate block

	if err != nil {
		return api.VersionedProposal{}, 0 * time.Second, fmt.Errorf("error proposing block at slot %d: %s", slot, err)
	}

	return *block.Data, blockTime, nil

}

func (s APIClient) SubmitProposalPreparation() error {
	err := s.Api.SubmitProposalPreparations(s.ctx, []*v1.ProposalPreparation{
		&v1.ProposalPreparation{
			ValidatorIndex: 0,
			FeeRecipient:   utils.CreateEmptyFeeRecipient(),
		},
	})

	return err

}
