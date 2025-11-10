package jobs

import (
	"fmt"
	"sync"
)

// History stores job execution history
type History struct {
	jobs     map[string]*Job
	byDevice map[string][]string // deviceID -> []jobID
	maxSize  int
	mu       sync.RWMutex
}

// NewHistory creates a new history
func NewHistory(maxSize int) *History {
	return &History{
		jobs:     make(map[string]*Job),
		byDevice: make(map[string][]string),
		maxSize:  maxSize,
	}
}

// Add adds a job to history
func (h *History) Add(job *Job) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Store job
	h.jobs[job.ID] = job

	// Add to device index
	h.byDevice[job.DeviceID] = append(h.byDevice[job.DeviceID], job.ID)

	// Cleanup if too large
	if len(h.jobs) > h.maxSize {
		h.cleanup()
	}
}

// Update updates a job in history
func (h *History) Update(job *Job) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.jobs[job.ID] = job
}

// Get retrieves a job by ID
func (h *History) Get(jobID string) (*Job, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	job, exists := h.jobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

// GetByDevice returns jobs for a device
func (h *History) GetByDevice(deviceID string, limit int) []*Job {
	h.mu.RLock()
	defer h.mu.RUnlock()

	jobIDs := h.byDevice[deviceID]
	if len(jobIDs) == 0 {
		return nil
	}

	// Get most recent jobs
	start := len(jobIDs) - limit
	if start < 0 {
		start = 0
	}

	jobs := make([]*Job, 0, limit)
	for i := len(jobIDs) - 1; i >= start; i-- {
		if job, exists := h.jobs[jobIDs[i]]; exists {
			jobs = append(jobs, job)
		}
	}

	return jobs
}

// Size returns the number of jobs in history
func (h *History) Size() int {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.jobs)
}

// cleanup removes oldest completed jobs
func (h *History) cleanup() {
	// Remove 20% of oldest completed jobs
	toRemove := h.maxSize / 5

	// Find oldest completed jobs
	var oldest []string
	for id, job := range h.jobs {
		if job.IsComplete() && len(oldest) < toRemove {
			oldest = append(oldest, id)
		}
	}

	// Remove them
	for _, id := range oldest {
		job := h.jobs[id]
		delete(h.jobs, id)

		// Remove from device index
		deviceJobs := h.byDevice[job.DeviceID]
		for i, jid := range deviceJobs {
			if jid == id {
				h.byDevice[job.DeviceID] = append(deviceJobs[:i], deviceJobs[i+1:]...)
				break
			}
		}
	}
}

// Clear clears all history
func (h *History) Clear() {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.jobs = make(map[string]*Job)
	h.byDevice = make(map[string][]string)
}
