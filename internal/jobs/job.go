package jobs

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Job represents an asynchronous device operation
type Job struct {
	ID              string
	DeviceID        string
	Type            string
	Status          Status
	Payload         interface{}
	Result          interface{}
	Error           error
	IdempotencyKey  string
	CreatedAt       time.Time
	StartedAt       *time.Time
	CompletedAt     *time.Time
	Timeout         time.Duration
}

// Status represents job execution state
type Status int

const (
	StatusPending Status = iota
	StatusInProgress
	StatusCompleted
	StatusFailed
	StatusCancelled
)

// String returns string representation of status
func (s Status) String() string {
	switch s {
	case StatusPending:
		return "pending"
	case StatusInProgress:
		return "in_progress"
	case StatusCompleted:
		return "completed"
	case StatusFailed:
		return "failed"
	case StatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// NewJob creates a new job
func NewJob(deviceID, jobType string, payload interface{}) *Job {
	return &Job{
		ID:        uuid.New().String(),
		DeviceID:  deviceID,
		Type:      jobType,
		Status:    StatusPending,
		Payload:   payload,
		CreatedAt: time.Now(),
		Timeout:   30 * time.Second,
	}
}

// NewJobWithIdempotency creates a new job with idempotency key
func NewJobWithIdempotency(deviceID, jobType string, payload interface{}, idempotencyKey string) *Job {
	job := NewJob(deviceID, jobType, payload)
	job.IdempotencyKey = idempotencyKey
	return job
}

// Start marks job as started
func (j *Job) Start() {
	now := time.Now()
	j.StartedAt = &now
	j.Status = StatusInProgress
}

// Complete marks job as completed
func (j *Job) Complete(result interface{}) {
	now := time.Now()
	j.CompletedAt = &now
	j.Status = StatusCompleted
	j.Result = result
}

// Fail marks job as failed
func (j *Job) Fail(err error) {
	now := time.Now()
	j.CompletedAt = &now
	j.Status = StatusFailed
	j.Error = err
}

// Cancel marks job as cancelled
func (j *Job) Cancel() {
	now := time.Now()
	j.CompletedAt = &now
	j.Status = StatusCancelled
}

// Duration returns job execution duration
func (j *Job) Duration() time.Duration {
	if j.StartedAt == nil {
		return 0
	}
	if j.CompletedAt == nil {
		return time.Since(*j.StartedAt)
	}
	return j.CompletedAt.Sub(*j.StartedAt)
}

// IsComplete returns true if job is in a terminal state
func (j *Job) IsComplete() bool {
	return j.Status == StatusCompleted || j.Status == StatusFailed || j.Status == StatusCancelled
}

// JobExecutor is a function that executes a job
type JobExecutor func(ctx context.Context, job *Job) (interface{}, error)
