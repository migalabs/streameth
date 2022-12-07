package app

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/tdahar/eth-cl-live-metrics/pkg/exporter"
)

// ServeMetrics:
// This method will serve the global peerstore values to the
// local prometheus instance.
func (s *AppService) ServeMetrics() {
	// Generate new ticker

	// generate the Prometheus exporter
	exptr, _ := exporter.NewMetricsExporter(
		s.ctx,
		"Client-Metrics-Prometheus",
		"Expose in Prometheus the metrics of the clients",
		s.initClientsPrometheusMetrics,
		s.runClientsPrometheusMetrics,
		func() {},
		exporter.MetricLoopInterval,
	)
	// add the new exptr to the ExporterService
	s.ExporterService.AddNewExporter(exptr)

}

func (s *AppService) initClientsPrometheusMetrics() {
	// register variables
	prometheus.MustRegister(ProposalsUp)
}
func (s *AppService) runClientsPrometheusMetrics() {
	ticker := time.NewTicker(exporter.MetricLoopInterval)
	// routine to loop
	go func() {
		for {
			select {
			case <-ticker.C:

				for _, item := range s.Analyzers {
					ProposalsUp.WithLabelValues(item.Eth2Provider.Label).Set(float64(item.Monitoring.ProposalStatus))
				}

			case <-s.ctx.Done():
				log.Info("Closing the prometheus metrics export service")
				// closing the routine in a ordened way
				ticker.Stop()
				return
			}
		}
	}()
}
