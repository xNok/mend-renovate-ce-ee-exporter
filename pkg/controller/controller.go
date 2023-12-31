package controller

import (
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/taskq/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.7.0"
	"google.golang.org/grpc"

	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/config"
	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/schemas"
	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/store"
)

type Controller struct {
	Config config.Config

	// UUID is used to identify this controller/process amongst others when
	// the exporter is running in cluster mode, leveraging Redis.
	UUID           uuid.UUID
	Redis          *redis.Client
	TaskController TaskController
	Store          store.Store

	Collectors RegistryCollectors
}

// New creates a new controller.
func New(ctx context.Context, cfg config.Config, version string) (c Controller, err error) {
	c.Config = cfg
	c.UUID = uuid.New()
	c.Collectors = make(RegistryCollectors)

	if err = configureTracing(ctx, &cfg.OpenTelemetry); err != nil {
		return
	}

	if err = c.configureRedis(ctx, cfg.Redis.URL); err != nil {
		return
	}

	c.TaskController = NewTaskController(ctx, c.Redis, cfg)
	c.Store = store.New(ctx, c.Redis)

	return
}

// configureTracing setup OTEL endpoint.
func configureTracing(ctx context.Context, cfg *config.OpenTelemetry) error {
	if len(cfg.GRPCEndpoint) == 0 {
		log.Debug("opentelemetry.grpc_endpoint is not configured, skipping open telemetry support")

		return nil
	}

	log.WithFields(
		log.Fields{
			"opentelemetry_grpc_endpoint": cfg.GRPCEndpoint,
		},
	).Info("opentelemetry gRPC endpoint provided, initializing connection..")

	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(cfg.GRPCEndpoint),
		otlptracegrpc.WithDialOption(grpc.WithBlock()),
	)

	traceExp, err := otlptrace.New(ctx, traceClient)
	if err != nil {
		return err
	}

	res, err := resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceNameKey),
		),
	)
	if err != nil {
		return err
	}

	bsp := sdktrace.NewBatchSpanProcessor(traceExp)
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)

	otel.SetTracerProvider(tracerProvider)

	return nil
}

// configureRedis is used in distributed mode, in that case the jobs/task backend is redis instead of in-memory.
func (c *Controller) configureRedis(ctx context.Context, url string) (err error) {
	ctx, span := otel.Tracer(c.Config.OpenTelemetry.ServiceNameKey).Start(ctx, "controller:configureRedis")
	defer span.End()

	if len(url) <= 0 {
		log.Debug("redis url is not configured, skipping configuration & using local driver")

		return
	}

	log.Info("redis url configured, initializing connection..")

	var opt *redis.Options

	if opt, err = redis.ParseURL(url); err != nil {
		return
	}

	c.Redis = redis.NewClient(opt)

	if err = redisotel.InstrumentTracing(c.Redis); err != nil {
		return
	}

	if _, err := c.Redis.Ping(ctx).Result(); err != nil {
		return errors.Wrap(err, "connecting to redis")
	}

	log.Info("connected to redis")

	return
}

// RegisterTasks is used to load the list of tasks to be handled.
func (c *Controller) RegisterTasks(n schemas.TaskType, h interface{}) {
	_, _ = c.TaskController.TaskMap.Register(
		string(n), &taskq.TaskConfig{
			Handler:    h,
			RetryLimit: 1,
		},
	)
}

// RegisterCollector is used to add collectors to the registry
func (c *Controller) RegisterCollector(ctx context.Context, collectors RegistryCollectors) {
	for kind, collector := range collectors {
		if _, ok := c.Collectors[kind]; ok {
			log.WithContext(ctx).Warn("Duplicated Collector key - skipping")
		}

		c.Collectors[kind] = collector
	}
}

func (c *Controller) UnqueueTask(ctx context.Context, tt schemas.TaskType, uniqueID string) {
	if err := c.Store.UnqueueTask(ctx, tt, uniqueID); err != nil {
		log.WithContext(ctx).
			WithFields(
				log.Fields{
					"task_type":      tt,
					"task_unique_id": uniqueID,
				},
			).
			WithError(err).
			Warn("unqueuing task")
	}
}
