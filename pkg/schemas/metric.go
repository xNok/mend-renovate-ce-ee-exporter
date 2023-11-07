package schemas

import (
	"hash/crc32"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

// MetricKind ..
type MetricKind int32

// Metric ..
type Metric struct {
	Kind   MetricKind
	Labels prometheus.Labels
	Value  float64
}

// MetricKey ..
type MetricKey string

// Metrics ..
type Metrics map[MetricKey]Metric

// Key is used to build the key based on the metric kind
// Keys should be build using a set of labels that identify the metric
// By default a checksum is used if no key pattern is defined.
func (m Metric) Key() MetricKey {
	key := strconv.Itoa(int(m.Kind))

	return MetricKey(strconv.Itoa(int(crc32.ChecksumIEEE([]byte(key)))))
}
