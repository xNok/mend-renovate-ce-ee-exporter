package schemas

import (
	"time"
)

// TaskType represents the type of a task.
type TaskType string

// Tasks can be used to keep track of tasks.
type Tasks map[TaskType]map[string]interface{}

// TaskSchedulingStatus represent the stat of the queued tasks.
type TaskSchedulingStatus struct {
	Last time.Time
	Next time.Time
}
