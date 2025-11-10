package telemetry

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics holds all Prometheus metrics
type Metrics struct {
	// Device metrics
	DeviceStatus *prometheus.GaugeVec
	DeviceUp     *prometheus.GaugeVec

	// Job metrics
	JobsTotal         *prometheus.CounterVec
	JobDuration       *prometheus.HistogramVec
	JobsInProgress    *prometheus.GaugeVec

	// Payment metrics
	PaymentsTotal     *prometheus.CounterVec
	PaymentAmountTotal *prometheus.CounterVec
	PaymentDuration   *prometheus.HistogramVec

	// Scanner metrics
	ScansTotal        *prometheus.CounterVec

	// System metrics
	Up                prometheus.Gauge

	registry *prometheus.Registry
}

// NewMetrics creates and registers all metrics
func NewMetrics() *Metrics {
	registry := prometheus.NewRegistry()

	m := &Metrics{
		DeviceStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "device_bridge_device_status",
				Help: "Current device status (0=unknown, 1=ready, 2=degraded, 3=offline, 4=error)",
			},
			[]string{"device_id", "kind"},
		),

		DeviceUp: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "device_bridge_device_up",
				Help: "Device availability (0=down, 1=up)",
			},
			[]string{"device_id", "kind"},
		),

		JobsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "device_bridge_jobs_total",
				Help: "Total number of jobs by device, type, and status",
			},
			[]string{"device_id", "type", "status"},
		),

		JobDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "device_bridge_job_duration_seconds",
				Help:    "Job execution duration in seconds",
				Buckets: prometheus.ExponentialBuckets(0.001, 2, 15), // 1ms to ~16s
			},
			[]string{"device_id", "type"},
		),

		JobsInProgress: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "device_bridge_jobs_in_progress",
				Help: "Current number of jobs in progress",
			},
			[]string{"device_id", "type"},
		),

		PaymentsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "device_bridge_payments_total",
				Help: "Total number of payments by provider and status",
			},
			[]string{"provider", "status"},
		),

		PaymentAmountTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "device_bridge_payment_amount_total",
				Help: "Total payment amount by provider and currency",
			},
			[]string{"provider", "currency"},
		),

		PaymentDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "device_bridge_payment_duration_seconds",
				Help:    "Payment processing duration in seconds",
				Buckets: prometheus.ExponentialBuckets(0.1, 2, 10), // 100ms to ~51s
			},
			[]string{"provider"},
		),

		ScansTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "device_bridge_scans_total",
				Help: "Total number of scans by device and symbology",
			},
			[]string{"device_id", "symbology"},
		),

		Up: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "device_bridge_up",
				Help: "Service is up (1) or down (0)",
			},
		),

		registry: registry,
	}

	// Register all metrics
	registry.MustRegister(
		m.DeviceStatus,
		m.DeviceUp,
		m.JobsTotal,
		m.JobDuration,
		m.JobsInProgress,
		m.PaymentsTotal,
		m.PaymentAmountTotal,
		m.PaymentDuration,
		m.ScansTotal,
		m.Up,
	)

	// Add Go runtime metrics
	registry.MustRegister(prometheus.NewGoCollector())
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))

	// Set service as up
	m.Up.Set(1)

	return m
}

// Handler returns HTTP handler for metrics endpoint
func (m *Metrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{
		EnableOpenMetrics: true,
	})
}

// RecordJobStart records job start
func (m *Metrics) RecordJobStart(deviceID, jobType string) {
	m.JobsInProgress.WithLabelValues(deviceID, jobType).Inc()
}

// RecordJobComplete records successful job completion
func (m *Metrics) RecordJobComplete(deviceID, jobType string, duration time.Duration) {
	m.JobsInProgress.WithLabelValues(deviceID, jobType).Dec()
	m.JobsTotal.WithLabelValues(deviceID, jobType, "completed").Inc()
	m.JobDuration.WithLabelValues(deviceID, jobType).Observe(duration.Seconds())
}

// RecordJobFailure records job failure
func (m *Metrics) RecordJobFailure(deviceID, jobType string, duration time.Duration) {
	m.JobsInProgress.WithLabelValues(deviceID, jobType).Dec()
	m.JobsTotal.WithLabelValues(deviceID, jobType, "failed").Inc()
	m.JobDuration.WithLabelValues(deviceID, jobType).Observe(duration.Seconds())
}

// SetDeviceStatus sets device status
func (m *Metrics) SetDeviceStatus(deviceID, kind string, status int) {
	m.DeviceStatus.WithLabelValues(deviceID, kind).Set(float64(status))
	if status == 2 { // Ready
		m.DeviceUp.WithLabelValues(deviceID, kind).Set(1)
	} else if status == 4 { // Offline
		m.DeviceUp.WithLabelValues(deviceID, kind).Set(0)
	}
}

// RecordScan records a scanner event
func (m *Metrics) RecordScan(deviceID, symbology string) {
	m.ScansTotal.WithLabelValues(deviceID, symbology).Inc()
}

// RecordPayment records a payment completion
func (m *Metrics) RecordPayment(provider, status string, amount float64, currency string, duration time.Duration) {
	m.PaymentsTotal.WithLabelValues(provider, status).Inc()
	if status == "approved" {
		m.PaymentAmountTotal.WithLabelValues(provider, currency).Add(amount)
	}
	m.PaymentDuration.WithLabelValues(provider).Observe(duration.Seconds())
}
