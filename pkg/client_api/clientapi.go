package client_api

import (
	"context"
	"time"

	"github.com/attestantio/go-eth2-client/http"
	"github.com/rs/zerolog"
	"github.com/sirupsen/logrus"
)

var (
	moduleName = "API-Cli"
	log        = logrus.WithField(
		"module", moduleName)
)

type APIClient struct {
	ctx   context.Context
	Api   *http.Service
	Label string
}

func NewAPIClient(ctx context.Context, label string, cliEndpoint string, timeout time.Duration) (*APIClient, error) {
	log.Debugf("generating http client at %s", cliEndpoint)
	httpCli, err := http.New(
		ctx,
		http.WithAddress(cliEndpoint),
		http.WithLogLevel(zerolog.WarnLevel),
		http.WithTimeout(timeout),
	)
	if err != nil {
		return &APIClient{}, err
	}

	hc, ok := httpCli.(*http.Service)
	if !ok {
		log.Error("gernerating the http api client")
	}
	return &APIClient{
		ctx:   ctx,
		Api:   hc,
		Label: label,
	}, nil
}

func (p APIClient) String() string {
	return p.Label + "->" + p.Api.Address()
}
