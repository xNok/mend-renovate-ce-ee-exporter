package store

import (
	"context"
	"sync"

	"github.com/xnok/mend-renovate-ce-ee-exporter/pkg/schemas"
)

type Local struct {
	metrics      schemas.Metrics
	metricsMutex sync.RWMutex

	tasks              schemas.Tasks
	tasksMutex         sync.RWMutex
	executedTasksCount uint64
}

// Metrics ..
func (l *Local) Metrics(_ context.Context) (metrics schemas.Metrics, err error) {
	metrics = make(schemas.Metrics)

	l.metricsMutex.RLock()
	defer l.metricsMutex.RUnlock()

	for k, v := range l.metrics {
		metrics[k] = v
	}

	return
}

// SetMetric ..
func (l *Local) SetMetric(_ context.Context, m schemas.Metric) error {
	l.metricsMutex.Lock()
	defer l.metricsMutex.Unlock()

	l.metrics[m.Key()] = m

	return nil
}

// DelMetric ..
func (l *Local) DelMetric(_ context.Context, k schemas.MetricKey) error {
	l.metricsMutex.Lock()
	defer l.metricsMutex.Unlock()

	delete(l.metrics, k)

	return nil
}

// GetMetric ..
func (l *Local) GetMetric(ctx context.Context, m *schemas.Metric) error {
	exists, _ := l.MetricExists(ctx, m.Key())

	if exists {
		l.metricsMutex.RLock()
		*m = l.metrics[m.Key()]
		l.metricsMutex.RUnlock()
	}

	return nil
}

// MetricExists ..
func (l *Local) MetricExists(_ context.Context, k schemas.MetricKey) (bool, error) {
	l.metricsMutex.RLock()
	defer l.metricsMutex.RUnlock()

	_, ok := l.metrics[k]

	return ok, nil
}

// MetricsCount ..
func (l *Local) MetricsCount(_ context.Context) (int64, error) {
	l.metricsMutex.RLock()
	defer l.metricsMutex.RUnlock()

	return int64(len(l.metrics)), nil
}

// CurrentlyQueuedTasksCount ..
func (l *Local) CurrentlyQueuedTasksCount(_ context.Context) (count uint64, err error) {
	l.tasksMutex.RLock()
	defer l.tasksMutex.RUnlock()

	for _, t := range l.tasks {
		count += uint64(len(t))
	}

	return
}

// isTaskAlreadyQueued assess if a task is already queued or not.
func (l *Local) isTaskAlreadyQueued(tt schemas.TaskType, uniqueID string) bool {
	l.tasksMutex.Lock()
	defer l.tasksMutex.Unlock()

	if l.tasks == nil {
		l.tasks = make(map[schemas.TaskType]map[string]interface{})
	}

	taskTypeQueue, ok := l.tasks[tt]
	if !ok {
		l.tasks[tt] = make(map[string]interface{})

		return false
	}

	if _, alreadyQueued := taskTypeQueue[uniqueID]; alreadyQueued {
		return true
	}

	return false
}

// QueueTask registers that we are queueing the task.
// It returns true if it managed to schedule it, false if it was already scheduled.
func (l *Local) QueueTask(_ context.Context, tt schemas.TaskType, uniqueID, _ string) (bool, error) {
	if !l.isTaskAlreadyQueued(tt, uniqueID) {
		l.tasksMutex.Lock()
		defer l.tasksMutex.Unlock()

		l.tasks[tt][uniqueID] = nil

		return true, nil
	}

	return false, nil
}

// UnqueueTask removes the task from the tracker.
func (l *Local) UnqueueTask(_ context.Context, tt schemas.TaskType, uniqueID string) error {
	if l.isTaskAlreadyQueued(tt, uniqueID) {
		l.tasksMutex.Lock()
		defer l.tasksMutex.Unlock()

		delete(l.tasks[tt], uniqueID)
		l.executedTasksCount++
	}

	return nil
}

// ExecutedTasksCount ..
func (l *Local) ExecutedTasksCount(_ context.Context) (uint64, error) {
	l.tasksMutex.RLock()
	defer l.tasksMutex.RUnlock()

	return l.executedTasksCount, nil
}
