package app

import (
	"github.com/migalabs/streameth/pkg/exporter"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	modName    = "app"
	modDetails = "general metrics about streameth"

	ProposalsUp = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "clients",
		Name:      "proposals_up",
		Help:      "Block Proposals up",
	},
		[]string{"clientName", "label"},
	)
)

func (c *AppService) GetPrometheusMetrics() *exporter.MetricsModule {
	metricsMod := exporter.NewMetricsModule(
		modName,
		modDetails,
	)
	// compose all the metrics

	metricsMod.AddIndvMetric(c.getProposalsUp())

	return metricsMod
}

func (s *AppService) getProposalsUp() *exporter.IndvMetrics {

	initFn := func() error {
		prometheus.MustRegister(ProposalsUp)
		return nil
	}

	updateFn := func() (interface{}, error) {
		countUp := 0

		for _, item := range s.Analyzers {
			ProposalsUp.With(
				prometheus.Labels{
					"clientName": item.GetClient(),
					"label":      item.GetLabel(),
				},
			).Set(float64(item.Monitoring.ProposalStatus))
			if item.Monitoring.ProposalStatus == 1 {
				countUp += 1
			}
		}
		return countUp, nil
	}

	indvMetr, err := exporter.NewIndvMetrics(
		"proposals_up",
		initFn,
		updateFn,
	)
	if err != nil {
		log.Error(errors.Wrap(err, "unable to init proposals_up"))
		return nil
	}

	return indvMetr
}
