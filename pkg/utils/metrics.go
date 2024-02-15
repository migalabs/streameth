package utils

import (
	"fmt"
	"strings"
)

const (
	AttestationMetric = "attestations"
	ProposalMetric    = "proposals"
	ReorgMetric       = "reorgs"
)

func ParseMetrics(metricsInput string) ([]string, error) {

	metrics := make([]string, 0)
	metricsArr := strings.Split(metricsInput, ",")
	for _, item := range metricsArr {
		if !checkValidMetric(item) {
			return make([]string, 0), fmt.Errorf("metric is not valid: %s", item)
		}
		metrics = append(metrics, item)
	}
	return metrics, nil
}

func checkValidMetric(metricInput string) bool {
	switch metricInput {
	case AttestationMetric:
		return true
	case ProposalMetric:
		return true
	case ReorgMetric:
		return true
	default:
		return false
	}
}
