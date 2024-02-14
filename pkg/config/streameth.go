package config

import (
	cli "github.com/urfave/cli/v2"
)

type StreamethConfig struct {
	LogLevel       string `json:"log-level"`
	BnEndpoints    string `json:"bn-endpoints"`
	DBEndpoint     string `json:"db-endpoint"`
	DbWorkers      int    `json:"db-workers"`
	Metrics        string `json:"metrics"`
	BlocksDir      string `json:"blocks-dir"`
	PrometheusPort int    `json:"prometheus-port"`
}

// TODO: read from config-file
func NewStreamethConfig() *StreamethConfig {
	// Return Default values for the ethereum configuration
	return &StreamethConfig{
		LogLevel:       DefaultLogLevel,
		BnEndpoints:    DefaultBnEndpoints,
		DBEndpoint:     DefaultDBEndpoint,
		DbWorkers:      DefaultDbWorkers,
		Metrics:        DefaultMetrics,
		BlocksDir:      DefaultBlocksDir,
		PrometheusPort: DefaultPrometheusPort,
	}
}

func (c *StreamethConfig) Apply(ctx *cli.Context) {
	// apply to the existing Default configuration the set flags
	// log level
	if ctx.IsSet("log-level") {
		c.LogLevel = ctx.String("log-level")
	}
	// cl url
	if ctx.IsSet("bn-endpoints") {
		c.BnEndpoints = ctx.String("bn-endpoints")
	}
	// db url
	if ctx.IsSet("db-endpoint") {
		c.DBEndpoint = ctx.String("db-endpoint")
	}
	// worker num
	if ctx.IsSet("db-workers") {
		c.DbWorkers = ctx.Int("db-workers")
	}
	// metrics
	if ctx.IsSet("metrics") {
		c.Metrics = ctx.String("metrics")
	}
	// blocksDir
	if ctx.IsSet("blocks-dir") {
		c.BlocksDir = ctx.String("blocks-dir")
	}
	// prometheus port
	if ctx.IsSet("prometheus-port") {
		c.PrometheusPort = ctx.Int("prometheus-port")
	}
}
