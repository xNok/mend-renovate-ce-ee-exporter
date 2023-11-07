package controller

import (
	"context"
	"net/http"

	"github.com/heptiolabs/healthcheck"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/trace"
)

// HealthCheckHandler ..
func (c *Controller) HealthCheckHandler(ctx context.Context) (h healthcheck.Handler) {
	h = healthcheck.NewHandler()

	return
}

// MetricsHandler ..
func (c *Controller) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)

	defer span.End()

	registry := NewRegistry(ctx)

	metrics, err := c.Store.Metrics(ctx)
	if err != nil {
		log.WithContext(ctx).
			WithError(err).
			Error()
	}

	if err := registry.ExportInternalMetrics(
		ctx,
		c.Store,
	); err != nil {
		log.WithContext(ctx).
			WithError(err).
			Warn()
	}

	registry.ExportMetrics(metrics)

	otelhttp.NewHandler(
		promhttp.HandlerFor(
			registry, promhttp.HandlerOpts{
				Registry:          registry,
				EnableOpenMetrics: c.Config.Server.Metrics.EnableOpenmetricsEncoding,
			},
		),
		"/metrics",
	).ServeHTTP(w, r)
}

// WebhookHandler ..
func (c *Controller) WebhookHandler(w http.ResponseWriter, r *http.Request) {
	span := trace.SpanFromContext(r.Context())
	defer span.End()

	// We create a new background context instead of relying on the request one which has a short cancellation TTL
	ctx := trace.ContextWithSpan(context.Background(), span)

	logger := log.
		WithContext(ctx).
		WithFields(
			log.Fields{
				"ip-address": r.RemoteAddr,
				"user-agent": r.UserAgent(),
			},
		)

	logger.Debug("webhook request")
}
