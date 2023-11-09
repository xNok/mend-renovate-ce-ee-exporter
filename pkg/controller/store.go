package controller

import (
	"context"

	log "github.com/sirupsen/logrus"

	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/schemas"
	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/store"
)

func metricLogFields(m schemas.Metric) log.Fields {
	return log.Fields{
		"metric-kind":   m.Kind,
		"metric-labels": m.Labels,
	}
}

func StoreGetMetric(ctx context.Context, s store.Store, m *schemas.Metric) {
	if err := s.GetMetric(ctx, m); err != nil {
		log.WithContext(ctx).
			WithFields(metricLogFields(*m)).
			WithError(err).
			Errorf("reading metric from the store")
	}
}

func StoreSetMetric(ctx context.Context, s store.Store, m schemas.Metric) {
	if err := s.SetMetric(ctx, m); err != nil {
		log.WithContext(ctx).
			WithFields(metricLogFields(m)).
			WithError(err).
			Errorf("writing metric from the store")
	}
}

func StoreDelMetric(ctx context.Context, s store.Store, m schemas.Metric) {
	if err := s.DelMetric(ctx, m.Key()); err != nil {
		log.WithContext(ctx).
			WithFields(metricLogFields(m)).
			WithError(err).
			Errorf("deleting metric from the store")
	}
}
