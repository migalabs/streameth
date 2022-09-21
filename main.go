package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/tdahar/block-scorer/pkg/cmd"
	cli "github.com/urfave/cli/v2"
)

var (
	Version = "v1.0.0"
	CliName = "Eth2 Block Scorer"
	log     = logrus.WithField(
		"cli", "CliName",
	)
)

func main() {
	fmt.Println(CliName, Version)

	// Set the general log configurations for the entire tool
	logrus.SetLevel(logrus.DebugLevel)

	app := &cli.App{
		Name:      CliName,
		Usage:     "Tinny client that requests and processes the Beacon Block proposals for each client.",
		UsageText: "block-scorer [commands] [arguments...]",
		Authors: []*cli.Author{
			{
				Name:  "Tarun",
				Email: "tarsuno@gmail.com",
			},
		},
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			cmd.ScorerCommand,
		},
	}
	// generate the crawler
	if err := app.RunContext(context.Background(), os.Args); err != nil {
		log.Errorf("error: %v\n", err)
		os.Exit(1)
	}
}
