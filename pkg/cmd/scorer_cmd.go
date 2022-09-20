package cmd

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tdahar/block-scorer/pkg/app"
	cli "github.com/urfave/cli/v2"
)

var ScorerCommand = &cli.Command{
	Name:   "block-scorer",
	Usage:  "Receive Block proposals from clients and evaluate score",
	Action: LaunchBlockScorer,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "bn-endpoints",
			Usage:       "beacon node endpoints (label/endpoint,label/endpoint)",
			DefaultText: "lh/localhost:5052",
		}},
}

var logRewardsRewards = logrus.WithField(
	"module", "ScorerCommand",
)

var QueryTimeout = 90 * time.Second

// CrawlAction is the function that is called when running `eth2`.
func LaunchBlockScorer(c *cli.Context) error {
	logRewardsRewards.Info("parsing flags")
	// check if a config file is set
	if !c.IsSet("bn-endpoints") {
		return errors.New("bn endpoint not provided")
	}

	bnEndpoints := strings.Split(c.String("bn-endpoints"), ",")
	fmt.Printf("%v", bnEndpoints)

	service, err := app.NewAppService(c.Context, bnEndpoints)
	if err != nil {
		log.Fatal("could not start app: %s", err.Error())
	}
	service.Run()

	return nil
}
