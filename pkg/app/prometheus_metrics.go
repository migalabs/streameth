package app

import (
	"github.com/prometheus/client_golang/prometheus"
)

// List of metrics that we are going to export
var (
	ProposalsUp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "clients",
		Name:      "proposals_up",
		Help:      "Block Proposals up",
	},
		[]string{"controldist"},
	)
)
