package additional_structs

import (
	"context"
	"fmt"
	"sync"

	"github.com/attestantio/go-eth2-client/api"
	api_v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/http"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	"github.com/sirupsen/logrus"
)

var (
	log = logrus.WithField(
		"module", "Epoch Data")
)

type EpochStructs struct {
	mu                       sync.Mutex
	Api                      *http.Service
	CurrentBeaconCommittees  []*api_v1.BeaconCommittee
	CurrentEpoch             uint64
	PreviousBeaconCommittees []*api_v1.BeaconCommittee
	PreviousEpoch            uint64
}

func NewEpochData(iApi *http.Service) EpochStructs {

	return EpochStructs{
		Api:                      iApi,
		CurrentBeaconCommittees:  make([]*api_v1.BeaconCommittee, 0),
		CurrentEpoch:             0,
		PreviousBeaconCommittees: make([]*api_v1.BeaconCommittee, 0),
		PreviousEpoch:            0,
	}
}

func (e *EpochStructs) RequestNewBeaconCommittee(slot uint64) error {
	epochCommittees, err := e.Api.BeaconCommittees(context.Background(), &api.BeaconCommitteesOpts{
		State: fmt.Sprintf("%d", slot),
	})

	if err != nil {
		return fmt.Errorf("could not request beacon committees for epoch %d: %s", int(slot/32), err)
	}

	// keep in mind we can only receive attestations to 32 blocks before
	e.PreviousBeaconCommittees = e.CurrentBeaconCommittees
	e.PreviousEpoch = e.CurrentEpoch
	e.CurrentBeaconCommittees = epochCommittees.Data
	e.CurrentEpoch = uint64(slot / 32)

	return nil

}

func (e *EpochStructs) GetBeaconCommittee(slot uint64, index uint64) []phase0.ValidatorIndex {
	log := log.WithField("routine", "epoch-structs")
	e.mu.Lock()
	// if the epoch requested is newer than the data we have
	if slot/32 > e.CurrentEpoch {
		log.Debugf("Requesting new beacon committee for %d", slot/32)
		e.RequestNewBeaconCommittee(slot)
	}
	e.mu.Unlock()

	committeeList := e.PreviousBeaconCommittees
	if slot/32 == e.CurrentEpoch {
		committeeList = e.CurrentBeaconCommittees
	}

	for _, item := range committeeList {
		if item.Slot == phase0.Slot(slot) && item.Index == phase0.CommitteeIndex(index) {
			return item.Validators
		}
	}

	return nil
}
