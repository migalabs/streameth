package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/migalabs/streameth/pkg/app"
	"github.com/migalabs/streameth/pkg/config"
	"github.com/migalabs/streameth/pkg/utils"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

var (
	log = logrus.WithField(
		"module", "cmd",
	)
)

var AnalyzerCommand = &cli.Command{
	Name:   "live-metrics",
	Usage:  "Receive Block proposals from clients and evaluate score, as well as other metrics",
	Action: LaunchBlockAnalyzer,
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:        "log-level",
			Usage:       "info,debug,warn",
			DefaultText: config.DefaultLogLevel,
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
			DefaultText: fmt.Sprintf("%d", config.DefaultDbWorkers),
		},
		&cli.StringFlag{
			Name:        "metrics",
			Usage:       "proposals,attestations",
			DefaultText: config.DefaultMetrics,
		},
		&cli.StringFlag{
			Name:        "blocks-dir",
			Usage:       "Folder where to store proposal blocks by label",
			DefaultText: config.DefaultBlocksDir,
		},
		&cli.StringFlag{
			Name:        "prometheus-port",
			Usage:       "Port where to listen for metrics",
			DefaultText: fmt.Sprintf("%d", config.DefaultPrometheusPort),
		}},
}

var QueryTimeout = 90 * time.Second

func LaunchBlockAnalyzer(c *cli.Context) error {

	conf := config.NewStreamethConfig()
	conf.Apply(c)

	logrus.SetLevel(utils.ParseLogLevel(conf.LogLevel))

	service, err := app.NewAppService(c.Context, *conf)
	if err != nil {
		log.Fatalf("could not start app: %s", err.Error())
	}

	procDoneC := make(chan struct{})
	sigtermC := make(chan os.Signal, 1)

	signal.Notify(sigtermC, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)

	go func() {
		service.Run()
		procDoneC <- struct{}{}
	}()

	select {
	case <-sigtermC:
		log.Info("Sudden shutdown detected, controlled shutdown of the cli triggered")
		service.Close()

	case <-procDoneC:
		log.Info("Process successfully finish!")
	}
	close(sigtermC)
	close(procDoneC)

	return nil
}
