package cmd

import (
	"strings"
	"time"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/tdahar/block-scorer/pkg/app"
	cli "github.com/urfave/cli/v2"
)

var AnalyzerCommand = &cli.Command{
	Name:   "block-scorer",
	Usage:  "Receive Block proposals from clients and evaluate score",
	Action: LaunchBlockAnalyzer,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "bn-endpoints",
			Usage:       "beacon node endpoints (label/endpoint,label/endpoint)",
			DefaultText: "lh/localhost:5052",
		}},
}

var QueryTimeout = 90 * time.Second

func LaunchBlockAnalyzer(c *cli.Context) error {
	logLauncher := log.WithField(
		"module", "ScorerCommand",
	)
	logLauncher.Info("parsing flags")
	// check if a beacon node is set
	if !c.IsSet("bn-endpoints") {
		return errors.New("bn endpoint not provided")
	}

	bnEndpoints := strings.Split(c.String("bn-endpoints"), ",")

	service, err := app.NewAppService(c.Context, bnEndpoints)
	if err != nil {
		log.Fatal("could not start app: %s", err.Error())
	}
	service.Run()

	return nil
}
