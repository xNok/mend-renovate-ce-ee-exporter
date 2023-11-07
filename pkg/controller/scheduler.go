package controller

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
	"github.com/vmihailenco/taskq/memqueue/v4"
	"github.com/vmihailenco/taskq/redisq/v4"
	"github.com/vmihailenco/taskq/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/config"
	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/schemas"
	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/store"
)

// TaskController holds task related clients.
type TaskController struct {
	Factory                  taskq.Factory
	Queue                    taskq.Queue
	TaskMap                  *taskq.TaskMap
	TaskSchedulingMonitoring map[schemas.TaskType]*schemas.TaskSchedulingStatus
}

// NewTaskController initializes and returns a new TaskController object.
func NewTaskController(ctx context.Context, r *redis.Client, cfg config.Config) (t TaskController) {
	ctx, span := otel.Tracer(cfg.OpenTelemetry.ServiceNameKey).Start(ctx, "controller:NewTaskController")
	defer span.End()

	t.TaskMap = &taskq.TaskMap{}

	queueOptions := &taskq.QueueConfig{
		Name:                 "default",
		PauseErrorsThreshold: 3,
		Handler:              t.TaskMap,
		BufferSize:           cfg.Scheduler.MaximumJobsQueueSize,
	}

	if r != nil {
		t.Factory = redisq.NewFactory()
		queueOptions.Redis = r
	} else {
		t.Factory = memqueue.NewFactory()
	}

	t.Queue = t.Factory.RegisterQueue(queueOptions)

	// Purge the queue when we start
	// I am only partially convinced this will not cause issues in HA fashion
	if err := t.Queue.Purge(ctx); err != nil {
		log.WithContext(ctx).
			WithError(err).
			Error("purging the pulling queue")
	}

	if r != nil {
		if err := t.Factory.StartConsumers(context.TODO()); err != nil {
			log.WithContext(ctx).
				WithError(err).
				Fatal("starting consuming the task queue")
		}
	}

	t.TaskSchedulingMonitoring = make(map[schemas.TaskType]*schemas.TaskSchedulingStatus)

	return
}

// Schedule ..
func (c *Controller) Schedule(ctx context.Context, pull config.Pull, gc config.GarbageCollect) {
	ctx, span := otel.Tracer(c.Config.OpenTelemetry.ServiceNameKey).Start(ctx, "controller:Schedule")
	defer span.End()

	for tt, cfg := range map[schemas.TaskType]config.SchedulerConfig{} {
		if cfg.OnInit {
			c.ScheduleTask(ctx, tt, "_")
		}

		if cfg.Scheduled {
			c.ScheduleTaskWithTicker(ctx, tt, cfg.IntervalSeconds)
		}

		if c.Redis != nil {
			c.ScheduleRedisSetKeepalive(ctx)
		}
	}
}

// ScheduleTask ..
func (c *Controller) ScheduleTask(ctx context.Context, tt schemas.TaskType, uniqueID string, args ...interface{}) {
	ctx, span := otel.Tracer(c.Config.OpenTelemetry.ServiceNameKey).Start(ctx, "controller:ScheduleTask")
	defer span.End()

	span.SetAttributes(attribute.String("task_type", string(tt)))
	span.SetAttributes(attribute.String("task_unique_id", uniqueID))

	logFields := log.Fields{
		"task_type":      tt,
		"task_unique_id": uniqueID,
	}
	task := c.TaskController.TaskMap.Get(string(tt))
	msg := task.NewJob(args...)

	qlen, err := c.TaskController.Queue.Len(ctx)
	if err != nil {
		log.WithContext(ctx).
			WithFields(logFields).
			Warn("unable to read task queue length, skipping scheduling of task..")

		return
	}

	if qlen >= c.TaskController.Queue.Options().BufferSize {
		log.WithContext(ctx).
			WithFields(logFields).
			Warn("queue buffer size exhausted, skipping scheduling of task..")

		return
	}

	queued, err := c.Store.QueueTask(ctx, tt, uniqueID, c.UUID.String())
	if err != nil {
		log.WithContext(ctx).
			WithFields(logFields).
			Warn("unable to declare the queueing, skipping scheduling of task..")

		return
	}

	if !queued {
		log.WithFields(logFields).
			Debug("task already queued, skipping scheduling of task..")

		return
	}

	go func(job *taskq.Job) {
		if err := c.TaskController.Queue.AddJob(ctx, job); err != nil {
			log.WithContext(ctx).
				WithError(err).
				Warn("scheduling task")
		}
	}(msg)
}

// ScheduleTaskWithTicker ..
func (c *Controller) ScheduleTaskWithTicker(ctx context.Context, tt schemas.TaskType, intervalSeconds int) {
	ctx, span := otel.Tracer(c.Config.OpenTelemetry.ServiceNameKey).Start(ctx, "controller:ScheduleTaskWithTicker")
	defer span.End()
	span.SetAttributes(attribute.String("task_type", string(tt)))
	span.SetAttributes(attribute.Int("interval_seconds", intervalSeconds))

	if intervalSeconds <= 0 {
		log.WithContext(ctx).
			WithField("task", tt).
			Warn("task scheduling misconfigured, currently disabled")

		return
	}

	log.WithFields(
		log.Fields{
			"task":             tt,
			"interval_seconds": intervalSeconds,
		},
	).Debug("task scheduled")

	c.TaskController.monitorNextTaskScheduling(tt, intervalSeconds)

	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Duration(intervalSeconds) * time.Second)

		for {
			select {
			case <-ctx.Done():
				log.WithField("task", tt).Info("scheduling of task stopped")

				return
			case <-ticker.C:
				c.ScheduleTask(ctx, tt, "_")
				c.TaskController.monitorNextTaskScheduling(tt, intervalSeconds)
			}
		}
	}(ctx)
}

// ScheduleRedisSetKeepalive will ensure that whilst the process is running,
// a key is periodically updated within Redis to let other instances know this
// one is alive and processing tasks.
func (c *Controller) ScheduleRedisSetKeepalive(ctx context.Context) {
	ctx, span := otel.Tracer(c.Config.OpenTelemetry.ServiceNameKey).Start(ctx, "controller:ScheduleRedisSetKeepalive")
	defer span.End()

	go func(ctx context.Context) {
		ticker := time.NewTicker(time.Duration(5) * time.Second)

		for {
			select {
			case <-ctx.Done():
				log.Info("stopped redis keepalive")

				return
			case <-ticker.C:
				if _, err := c.Store.(*store.Redis).SetKeepalive(ctx, c.UUID.String(), time.Duration(10)*time.Second); err != nil {
					log.WithContext(ctx).
						WithError(err).
						Fatal("setting keepalive")
				}
			}
		}
	}(ctx)
}

func (tc *TaskController) monitorNextTaskScheduling(tt schemas.TaskType, duration int) {
	if _, ok := tc.TaskSchedulingMonitoring[tt]; !ok {
		tc.TaskSchedulingMonitoring[tt] = &schemas.TaskSchedulingStatus{}
	}

	tc.TaskSchedulingMonitoring[tt].Next = time.Now().Add(time.Duration(duration) * time.Second)
}

func (tc *TaskController) monitorLastTaskScheduling(tt schemas.TaskType) {
	if _, ok := tc.TaskSchedulingMonitoring[tt]; !ok {
		tc.TaskSchedulingMonitoring[tt] = &schemas.TaskSchedulingStatus{}
	}

	tc.TaskSchedulingMonitoring[tt].Last = time.Now()
}
