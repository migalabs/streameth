package cmd

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/migalabs/streameth/pkg/app"
	"github.com/migalabs/streameth/pkg/utils"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

const (
	DefaultBlocksDir = "./proposal_blocks_data"
)

var AnalyzerCommand = &cli.Command{
	Name:   "live-metrics",
	Usage:  "Receive Block proposals from clients and evaluate score, as well as other metrics",
	Action: LaunchBlockAnalyzer,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			Usage:       "info,debug,warn",
			DefaultText: "info",
		},
		&cli.StringFlag{
			Name:  "bn-endpoints",
			Usage: "beacon node endpoints (client/label/endpoint,client/label/endpoint)",
		},
		&cli.StringFlag{
			Name:  "db-endpoint",
			Usage: "postgresql database endpoint: postgresql://user:password@localhost:5432/beaconchain",
		},
		&cli.StringFlag{
			Name:        "db-workers",
			Usage:       "10",
			DefaultText: "1",
		},
		&cli.StringFlag{
			Name:        "metrics",
			Usage:       "proposals,attestations",
			DefaultText: "proposals,attestations",
		},
		&cli.StringFlag{
			Name:        "blocks-dir",
			Usage:       "Folder where to store proposal blocks by label",
			DefaultText: DefaultBlocksDir,
		},
		&cli.StringFlag{
			Name:        "prometheus-port",
			Usage:       "Port where to listen for metrics",
			DefaultText: DefaultBlocksDir,
		}},
}

var QueryTimeout = 90 * time.Second

func LaunchBlockAnalyzer(c *cli.Context) error {
	dbWorkers := 1
	blocksDir := DefaultBlocksDir
	metrics := make([]string, 0)
	prometheuPort := app.DefaultMetricsPort
	logLauncher := log.WithField(
		"module", "ScorerCommand",
	)

	logLauncher.Info("parsing flags")
	if !c.IsSet("log-level") {
		logLauncher.Infof("Setting log level to Info")
		logrus.SetLevel(logrus.InfoLevel)
	} else {
		logrus.SetLevel(utils.ParseLogLevel(c.String("log-level")))
	}
	// check if a beacon node is set
	if !c.IsSet("bn-endpoints") {
		return errors.New("bn endpoint not provided")
	}

	if !c.IsSet("db-endpoint") {
		return errors.New("db endpoint not provided")
	}

	if !c.IsSet("db-workers") {
		logLauncher.Warnf("no database workers configured, default is 1")
		metrics = append(metrics, "proposals")
		metrics = append(metrics, "attestations")
	} else {
		metricsInput := strings.Split(c.String("metrics"), ",")

		for _, item := range metricsInput {
			metrics = append(metrics, item)
		}
	}

	if !c.IsSet("metrics") {
		logLauncher.Warnf("no metrics configured, measuring all")
	} else {
		dbWorkers = c.Int("db-workers")
	}

	if c.IsSet("prometheus-port") {
		prometheuPort = c.Int("prometheus-port")
	}

	if !c.IsSet("blocks-dir") {
		logLauncher.Warnf("no blocks dir configured, defaulting to ./proposal_blocks_data")
	} else {
		blocksDir = c.String("blocks-dir")
	}

	bnEndpoints := strings.Split(c.String("bn-endpoints"), ",")
	dbEndpoint := c.String("db-endpoint")

	service, err := app.NewAppService(c.Context, bnEndpoints, dbEndpoint, dbWorkers, metrics, blocksDir, prometheuPort)
	if err != nil {
		log.Fatal("could not start app: %s", err.Error())
	}

	procDoneC := make(chan struct{})
	sigtermC := make(chan os.Signal)

	signal.Notify(sigtermC, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)

	go func() {
		service.Run()
		procDoneC <- struct{}{}
	}()

	select {
	case <-sigtermC:
		logLauncher.Info("Sudden shutdown detected, controlled shutdown of the cli triggered")
		service.Close()

	case <-procDoneC:
		logLauncher.Info("Process successfully finish!")
	}
	close(sigtermC)
	close(procDoneC)

	return nil
}
