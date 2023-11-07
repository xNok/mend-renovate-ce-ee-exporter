package controller

import "github.com/prometheus/client_golang/prometheus"

// NewInternalCollectorCurrentlyQueuedTasksCount returns a new collector for the gcpe_currently_queued_tasks_count metric.
func NewInternalCollectorCurrentlyQueuedTasksCount() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gcpe_currently_queued_tasks_count",
			Help: "Number of tasks in the queue",
		},
		[]string{},
	)
}

// NewInternalCollectorExecutedTasksCount returns a new collector for the gcpe_executed_tasks_count metric.
func NewInternalCollectorExecutedTasksCount() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gcpe_executed_tasks_count",
			Help: "Number of tasks executed",
		},
		[]string{},
	)
}

// NewInternalCollectorMetricsCount returns a new collector for the gcpe_metrics_count metric.
func NewInternalCollectorMetricsCount() prometheus.Collector {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "gcpe_metrics_count",
			Help: "Number of GitLab pipelines metrics being exported",
		},
		[]string{},
	)
}
