package store

import (
	"context"

	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"

	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/schemas"
)

// Store ..
type Store interface {
	// Metrics ..
	Metrics(context.Context) (schemas.Metrics, error)
	SetMetric(context.Context, schemas.Metric) error
	DelMetric(context.Context, schemas.MetricKey) error
	GetMetric(context.Context, *schemas.Metric) error
	MetricExists(context.Context, schemas.MetricKey) (bool, error)
	MetricsCount(context.Context) (int64, error)
	// QueueTask Helpers to keep track of currently queued tasks and avoid scheduling them
	// twice at the risk of ending up with loads of dangling goroutines being locked
	QueueTask(context.Context, schemas.TaskType, string, string) (bool, error)
	UnqueueTask(context.Context, schemas.TaskType, string) error
	CurrentlyQueuedTasksCount(context.Context) (uint64, error)
	ExecutedTasksCount(context.Context) (uint64, error)
}

// NewLocalStore ..
func NewLocalStore() Store {
	return &Local{}
}

// NewRedisStore ..
func NewRedisStore(client *redis.Client) Store {
	return &Redis{
		Client: client,
	}
}

// New creates a new store and populates it with
// provided []schemas.Project.
func New(
	ctx context.Context,
	r *redis.Client,
) (s Store) {
	_, span := otel.Tracer("mend-renovate-ce-ee-exporter").Start(ctx, "store:New")
	defer span.End()

	if r != nil {
		s = NewRedisStore(r)
	} else {
		s = NewLocalStore()
	}

	return
}
