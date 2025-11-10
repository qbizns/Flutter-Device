package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Macber-eg/Flutter-Device/internal/telemetry"
)

// Queue manages job submission and execution
type Queue struct {
	jobs       chan *Job
	workers    int
	executors  map[string]JobExecutor
	history    *History
	idempotent map[string]*Job
	idempMu    sync.RWMutex
	wg         sync.WaitGroup
	logger     *telemetry.Logger
	metrics    *telemetry.Metrics
	ctx        context.Context
	cancel     context.CancelFunc
}

// QueueConfig contains queue configuration
type QueueConfig struct {
	Workers     int
	QueueSize   int
	HistorySize int
}

// DefaultQueueConfig returns default configuration
func DefaultQueueConfig() QueueConfig {
	return QueueConfig{
		Workers:     10,
		QueueSize:   100,
		HistorySize: 1000,
	}
}

// NewQueue creates a new job queue
func NewQueue(config QueueConfig, logger *telemetry.Logger, metrics *telemetry.Metrics) *Queue {
	ctx, cancel := context.WithCancel(context.Background())

	return &Queue{
		jobs:       make(chan *Job, config.QueueSize),
		workers:    config.Workers,
		executors:  make(map[string]JobExecutor),
		history:    NewHistory(config.HistorySize),
		idempotent: make(map[string]*Job),
		logger:     logger,
		metrics:    metrics,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// RegisterExecutor registers a job executor for a job type
func (q *Queue) RegisterExecutor(jobType string, executor JobExecutor) {
	q.executors[jobType] = executor
}

// Submit submits a job to the queue
func (q *Queue) Submit(job *Job) error {
	// Check idempotency
	if job.IdempotencyKey != "" {
		q.idempMu.RLock()
		existing, exists := q.idempotent[job.IdempotencyKey]
		q.idempMu.RUnlock()

		if exists {
			q.logger.Info("idempotent job already exists",
				telemetry.String("job_id", existing.ID),
				telemetry.String("idempotency_key", job.IdempotencyKey),
			)
			// Return existing job ID in the new job
			job.ID = existing.ID
			job.Status = existing.Status
			job.Result = existing.Result
			job.Error = existing.Error
			return nil
		}

		// Store for idempotency
		q.idempMu.Lock()
		q.idempotent[job.IdempotencyKey] = job
		q.idempMu.Unlock()
	}

	// Add to history
	q.history.Add(job)

	// Submit to queue
	select {
	case q.jobs <- job:
		q.logger.Info("job submitted",
			telemetry.String("job_id", job.ID),
			telemetry.String("device_id", job.DeviceID),
			telemetry.String("type", job.Type),
		)
		return nil
	case <-q.ctx.Done():
		return fmt.Errorf("queue is shutting down")
	default:
		return fmt.Errorf("queue is full")
	}
}

// Start starts the worker pool
func (q *Queue) Start() {
	q.logger.Info("starting job queue", telemetry.Int("workers", q.workers))

	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
}

// Stop stops the queue and waits for workers to finish
func (q *Queue) Stop() {
	q.logger.Info("stopping job queue")
	q.cancel()
	close(q.jobs)
	q.wg.Wait()
	q.logger.Info("job queue stopped")
}

// worker processes jobs from the queue
func (q *Queue) worker(id int) {
	defer q.wg.Done()

	logger := q.logger.WithFields(telemetry.Int("worker_id", id))
	logger.Info("worker started")

	for job := range q.jobs {
		q.processJob(job, logger)
	}

	logger.Info("worker stopped")
}

// processJob executes a single job
func (q *Queue) processJob(job *Job, logger *telemetry.Logger) {
	// Get executor
	executor, exists := q.executors[job.Type]
	if !exists {
		job.Fail(fmt.Errorf("no executor registered for job type: %s", job.Type))
		q.history.Update(job)
		logger.Error("no executor for job type",
			telemetry.String("job_id", job.ID),
			telemetry.String("type", job.Type),
		)
		return
	}

	// Start job
	job.Start()
	q.history.Update(job)

	logger.Info("executing job",
		telemetry.String("job_id", job.ID),
		telemetry.String("device_id", job.DeviceID),
		telemetry.String("type", job.Type),
	)

	// Record metrics
	if q.metrics != nil {
		q.metrics.RecordJobStart(job.DeviceID, job.Type)
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(q.ctx, job.Timeout)
	defer cancel()

	result, err := executor(ctx, job)

	// Complete job
	if err != nil {
		job.Fail(err)
		logger.Error("job failed",
			telemetry.String("job_id", job.ID),
			telemetry.String("device_id", job.DeviceID),
			telemetry.String("type", job.Type),
			telemetry.Error(err),
		)

		if q.metrics != nil {
			q.metrics.RecordJobFailure(job.DeviceID, job.Type, job.Duration())
		}
	} else {
		job.Complete(result)
		logger.Info("job completed",
			telemetry.String("job_id", job.ID),
			telemetry.String("device_id", job.DeviceID),
			telemetry.String("type", job.Type),
			telemetry.Duration("duration", job.Duration()),
		)

		if q.metrics != nil {
			q.metrics.RecordJobComplete(job.DeviceID, job.Type, job.Duration())
		}
	}

	q.history.Update(job)

	// Clean up idempotency after some time
	if job.IdempotencyKey != "" {
		go func() {
			time.Sleep(24 * time.Hour)
			q.idempMu.Lock()
			delete(q.idempotent, job.IdempotencyKey)
			q.idempMu.Unlock()
		}()
	}
}

// Get returns a job by ID
func (q *Queue) Get(jobID string) (*Job, error) {
	return q.history.Get(jobID)
}

// GetByDevice returns jobs for a device
func (q *Queue) GetByDevice(deviceID string, limit int) []*Job {
	return q.history.GetByDevice(deviceID, limit)
}

// Stats returns queue statistics
func (q *Queue) Stats() QueueStats {
	return QueueStats{
		QueueLength: len(q.jobs),
		Workers:     q.workers,
		HistorySize: q.history.Size(),
	}
}

// QueueStats contains queue statistics
type QueueStats struct {
	QueueLength int
	Workers     int
	HistorySize int
}
