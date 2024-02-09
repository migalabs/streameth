package main

import (
	"context"
	"fmt"
	"os"

	"github.com/migalabs/streameth/pkg/cmd"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

var (
	Version = "v1.0.0"
	CliName = "Eth CL Live Metrics"
	log     = logrus.WithField(
		"cli", "CliName",
	)
)

func main() {
	fmt.Println(CliName, Version)

	// Set the general log configurations for the entire tool
	logrus.SetLevel(logrus.InfoLevel)

	app := &cli.App{
		Name:      CliName,
		Usage:     "Tinny client that requests and processes the Beacon Block proposals for each client.",
		UsageText: "live-metrics [commands] [arguments...]",
		Authors: []*cli.Author{
			{
				Name:  "Tarun",
				Email: "tarsuno@gmail.com",
			},
		},
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			cmd.AnalyzerCommand,
		},
	}
	// generate the block analyzer
	if err := app.RunContext(context.Background(), os.Args); err != nil {
		log.Errorf("error: %v\n", err)
		os.Exit(1)
	}
}
