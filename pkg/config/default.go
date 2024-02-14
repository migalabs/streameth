package config

var (
	DefaultLogLevel       string = "info"
	DefaultBnEndpoints    string = ""
	DefaultDBEndpoint     string = "postgres://user:password@localhost:5432/goteth"
	DefaultDbWorkers      int    = 1
	DefaultMetrics        string = "proposal"
	DefaultBlocksDir      string = "./block_proposals"
	DefaultPrometheusPort int    = 9080
)
