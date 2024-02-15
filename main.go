package main

import (
	"context"
	"fmt"
	"os"

	"github.com/migalabs/streameth/cmd"
	"github.com/migalabs/streameth/pkg/utils"
	"github.com/sirupsen/logrus"
	cli "github.com/urfave/cli/v2"
)

var (
	log = logrus.WithField(
		"cli", "CliName",
	)
)

func main() {
	fmt.Println(utils.CliName, utils.Version)

	// Set the general log configurations for the entire tool
	logrus.SetLevel(logrus.InfoLevel)

	app := &cli.App{
		Name:      utils.CliName,
		Usage:     "Tinny client that requests and processes the Beacon Block proposals for each client.",
		UsageText: "streameth [commands] [arguments...]",
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
