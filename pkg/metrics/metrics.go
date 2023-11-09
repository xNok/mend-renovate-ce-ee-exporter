package metrics

import (
	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/schemas"
)

// List of all the metrics supported by this exporter
const (
	// MetricKindRenovateJobsQueueLength ..
	MetricKindRenovateJobsQueueLength schemas.MetricKind = iota
)
