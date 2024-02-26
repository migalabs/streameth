package client_api

import (
	"fmt"

	"github.com/attestantio/go-eth2-client/api"
	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

func (s *APIClient) ProposerDuties(epoch phase0.Epoch) ([]*v1.ProposerDuty, error) {

	proposerDuties, err := s.Api.ProposerDuties(s.ctx, &api.ProposerDutiesOpts{
		Epoch: epoch,
	})

	if err != nil {
		return nil, fmt.Errorf("could not get proposer duties in epoch %d: %s", epoch, err)
	}

	duties := make([]*v1.ProposerDuty, len(proposerDuties.Data))

	for i, item := range proposerDuties.Data {
		duties[i] = item
	}

	return duties, nil
}
