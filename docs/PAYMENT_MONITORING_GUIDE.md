# Payment Terminal Driver - Monitoring & Observability Guide

**Version:** 1.0
**Date:** November 11, 2025
**Audience:** DevOps Engineers, SRE Teams, Operations

---

## Table of Contents

- [1. Overview](#1-overview)
- [2. Metrics Collection](#2-metrics-collection)
- [3. Monitoring Dashboards](#3-monitoring-dashboards)
- [4. Alerting Rules](#4-alerting-rules)
- [5. Log Management](#5-log-management)
- [6. Distributed Tracing](#6-distributed-tracing)
- [7. Health Checks](#7-health-checks)
- [8. Performance Monitoring](#8-performance-monitoring)
- [9. Incident Response](#9-incident-response)
- [10. Best Practices](#10-best-practices)

---

## 1. Overview

This guide covers monitoring, metrics, logging, and observability for the Device Bridge v2 payment terminal driver. Proper monitoring ensures high availability, quick incident response, and proactive issue detection.

### Monitoring Philosophy

1. **Four Golden Signals:**
   - **Latency:** Transaction processing time
   - **Traffic:** Transactions per second
   - **Errors:** Failed transaction rate
   - **Saturation:** Connection pool usage

2. **RED Method:**
   - **Rate:** Request rate (TPS)
   - **Errors:** Error rate (%)
   - **Duration:** Response time (ms)

3. **USE Method:**
   - **Utilization:** Resource usage (%)
   - **Saturation:** Queue depth
   - **Errors:** Error count

---

## 2. Metrics Collection

### 2.1 Prometheus Metrics

**Key Metrics to Export:**

```go
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Transaction Metrics
	TransactionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_transactions_total",
			Help: "Total number of payment transactions",
		},
		[]string{"device_id", "terminal_id", "type", "status", "provider"},
	)

	TransactionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_transaction_duration_seconds",
			Help:    "Transaction processing duration in seconds",
			Buckets: []float64{0.1, 0.5, 1.0, 2.0, 5.0, 10.0, 30.0},
		},
		[]string{"device_id", "type", "provider"},
	)

	TransactionAmount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_transaction_amount",
			Help:    "Transaction amounts in smallest currency unit",
			Buckets: prometheus.ExponentialBuckets(100, 2, 15),
		},
		[]string{"device_id", "currency", "provider"},
	)

	// Connection Metrics
	ConnectionsActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "payment_connections_active",
			Help: "Number of active terminal connections",
		},
		[]string{"device_id", "terminal_id"},
	)

	ConnectionErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_connection_errors_total",
			Help: "Total number of connection errors",
		},
		[]string{"device_id", "terminal_id", "error_type"},
	)

	ConnectionDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "payment_connection_duration_seconds",
			Help:    "Connection establishment duration",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
		},
		[]string{"device_id", "terminal_id"},
	)

	// Settlement Metrics
	SettlementsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_settlements_total",
			Help: "Total number of settlements",
		},
		[]string{"device_id", "terminal_id", "status"},
	)

	SettlementTransactionCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "payment_settlement_transaction_count",
			Help: "Number of transactions in last settlement",
		},
		[]string{"device_id", "terminal_id"},
	)

	SettlementAmount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "payment_settlement_amount",
			Help: "Total amount in last settlement",
		},
		[]string{"device_id", "terminal_id", "currency"},
	)

	// Audit Metrics
	AuditEventsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_audit_events_total",
			Help: "Total number of audit events",
		},
		[]string{"device_id", "event_type", "action"},
	)

	AuditBufferSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "payment_audit_buffer_size",
			Help: "Current size of audit buffer",
		},
		[]string{"device_id"},
	)

	// Error Metrics
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_errors_total",
			Help: "Total number of errors",
		},
		[]string{"device_id", "error_type", "error_code"},
	)

	DeclinedTransactions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "payment_declined_transactions_total",
			Help: "Total number of declined transactions",
		},
		[]string{"device_id", "decline_reason", "response_code"},
	)

	// Performance Metrics
	STANCounter = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "payment_stan_counter",
			Help: "Current STAN counter value",
		},
		[]string{"device_id", "terminal_id"},
	)

	BatchNumber = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "payment_batch_number",
			Help: "Current batch number",
		},
		[]string{"device_id", "terminal_id"},
	)

	// System Metrics
	SystemInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "payment_system_info",
			Help: "System information",
		},
		[]string{"version", "go_version", "build_date"},
	)
)
```

### 2.2 Metrics Endpoint Configuration

**prometheus.yml:**
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'device-bridge-payment'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: '/metrics'
    scrape_interval: 10s
    scrape_timeout: 5s
```

### 2.3 Custom Metrics Implementation

**Example: Recording Transaction:**
```go
import (
	"time"
	"github.com/prometheus/client_golang/prometheus"
)

func (d *Driver) ProcessTransactionWithMetrics(ctx context.Context, req TransactionRequest) (*TransactionResponse, error) {
	start := time.Now()

	// Process transaction
	resp, err := d.ProcessTransaction(ctx, req)

	// Record metrics
	duration := time.Since(start).Seconds()
	status := "success"
	if err != nil || !resp.Success {
		status = "failed"
	}

	// Increment counter
	metrics.TransactionsTotal.WithLabelValues(
		d.id,
		d.config.TerminalID,
		string(req.Type),
		status,
		d.config.Provider,
	).Inc()

	// Record duration
	metrics.TransactionDuration.WithLabelValues(
		d.id,
		string(req.Type),
		d.config.Provider,
	).Observe(duration)

	// Record amount
	if resp != nil {
		metrics.TransactionAmount.WithLabelValues(
			d.id,
			resp.Currency,
			d.config.Provider,
		).Observe(float64(resp.Amount))
	}

	return resp, err
}
```

---

## 3. Monitoring Dashboards

### 3.1 Grafana Dashboard - Payment Overview

**dashboard.json** (excerpt):
```json
{
  "dashboard": {
    "title": "Payment Terminal Overview",
    "panels": [
      {
        "title": "Transactions Per Second",
        "targets": [
          {
            "expr": "rate(payment_transactions_total[1m])",
            "legendFormat": "{{device_id}} - {{type}}"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Transaction Success Rate",
        "targets": [
          {
            "expr": "sum(rate(payment_transactions_total{status=\"success\"}[5m])) / sum(rate(payment_transactions_total[5m])) * 100",
            "legendFormat": "Success Rate %"
          }
        ],
        "type": "gauge",
        "thresholds": [
          {"value": 95, "color": "red"},
          {"value": 98, "color": "yellow"},
          {"value": 99, "color": "green"}
        ]
      },
      {
        "title": "Average Transaction Duration",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(payment_transaction_duration_seconds_bucket[5m]))",
            "legendFormat": "p95"
          },
          {
            "expr": "histogram_quantile(0.50, rate(payment_transaction_duration_seconds_bucket[5m]))",
            "legendFormat": "p50"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Active Connections",
        "targets": [
          {
            "expr": "payment_connections_active",
            "legendFormat": "{{terminal_id}}"
          }
        ],
        "type": "stat"
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(payment_errors_total[5m])",
            "legendFormat": "{{error_type}}"
          }
        ],
        "type": "graph"
      }
    ]
  }
}
```

### 3.2 Key Dashboard Panels

**1. Transaction Volume**
- Total transactions (counter)
- Transactions per second (rate)
- Transactions by type (pie chart)
- Transactions by provider (bar chart)

**2. Transaction Success**
- Success rate (gauge)
- Failure rate (gauge)
- Declined transactions (counter)
- Error breakdown (pie chart)

**3. Performance**
- Transaction latency (p50, p95, p99)
- Connection latency
- Settlement duration
- STAN counter progression

**4. System Health**
- Active connections (gauge)
- Connection errors (counter)
- Audit buffer size (gauge)
- Memory usage (graph)

**5. Business Metrics**
- Transaction amounts (histogram)
- Settlement amounts (counter)
- Average ticket size (stat)
- Peak transaction times (heatmap)

### 3.3 Dashboard Templates

**Variables:**
```yaml
variables:
  - name: device_id
    type: query
    query: label_values(payment_transactions_total, device_id)

  - name: terminal_id
    type: query
    query: label_values(payment_transactions_total{device_id="$device_id"}, terminal_id)

  - name: provider
    type: query
    query: label_values(payment_transactions_total, provider)

  - name: time_range
    type: interval
    options: [5m, 15m, 1h, 6h, 24h, 7d]
```

---

## 4. Alerting Rules

### 4.1 Prometheus Alert Rules

**payment_alerts.yml:**
```yaml
groups:
  - name: payment_critical
    interval: 30s
    rules:
      - alert: PaymentHighErrorRate
        expr: |
          (
            sum(rate(payment_transactions_total{status="failed"}[5m]))
            /
            sum(rate(payment_transactions_total[5m]))
          ) > 0.05
        for: 2m
        labels:
          severity: critical
          component: payment
        annotations:
          summary: "High payment error rate"
          description: "Error rate is {{ $value | humanizePercentage }} (threshold: 5%)"

      - alert: PaymentServiceDown
        expr: up{job="device-bridge-payment"} == 0
        for: 1m
        labels:
          severity: critical
          component: payment
        annotations:
          summary: "Payment service is down"
          description: "Device Bridge payment service has been down for 1 minute"

      - alert: PaymentHighLatency
        expr: |
          histogram_quantile(0.95,
            rate(payment_transaction_duration_seconds_bucket[5m])
          ) > 5
        for: 3m
        labels:
          severity: warning
          component: payment
        annotations:
          summary: "High payment processing latency"
          description: "P95 latency is {{ $value }}s (threshold: 5s)"

      - alert: PaymentConnectionErrors
        expr: rate(payment_connection_errors_total[5m]) > 0.1
        for: 2m
        labels:
          severity: warning
          component: payment
        annotations:
          summary: "High connection error rate"
          description: "Connection errors: {{ $value }} per second"

      - alert: PaymentNoTransactions
        expr: |
          sum(rate(payment_transactions_total[10m])) == 0
          and
          hour() >= 8 and hour() <= 22
        for: 10m
        labels:
          severity: warning
          component: payment
        annotations:
          summary: "No payment transactions"
          description: "No transactions processed in last 10 minutes during business hours"

      - alert: PaymentAuditBufferFull
        expr: payment_audit_buffer_size >= 0.9 * 1000
        for: 1m
        labels:
          severity: warning
          component: payment
        annotations:
          summary: "Audit buffer nearly full"
          description: "Audit buffer is {{ $value }} events (threshold: 900)"

      - alert: PaymentHighDeclineRate
        expr: |
          (
            sum(rate(payment_declined_transactions_total[5m]))
            /
            sum(rate(payment_transactions_total[5m]))
          ) > 0.20
        for: 5m
        labels:
          severity: warning
          component: payment
        annotations:
          summary: "High transaction decline rate"
          description: "Decline rate is {{ $value | humanizePercentage }} (threshold: 20%)"

  - name: payment_warning
    interval: 1m
    rules:
      - alert: PaymentSlowTransactions
        expr: |
          histogram_quantile(0.50,
            rate(payment_transaction_duration_seconds_bucket[5m])
          ) > 2
        for: 5m
        labels:
          severity: info
          component: payment
        annotations:
          summary: "Slow payment transactions"
          description: "Median latency is {{ $value }}s (threshold: 2s)"

      - alert: PaymentSettlementDelayed
        expr: |
          (time() - payment_batch_number * 86400) > 93600
        labels:
          severity: info
          component: payment
        annotations:
          summary: "Settlement may be delayed"
          description: "Last settlement was over 26 hours ago"
```

### 4.2 Alertmanager Configuration

**alertmanager.yml:**
```yaml
global:
  resolve_timeout: 5m
  slack_api_url: 'https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK'

route:
  receiver: 'default'
  group_by: ['alertname', 'device_id']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h

  routes:
    - match:
        severity: critical
      receiver: 'pagerduty'
      continue: true

    - match:
        severity: critical
      receiver: 'slack-critical'

    - match:
        severity: warning
      receiver: 'slack-warning'

receivers:
  - name: 'default'
    email_configs:
      - to: 'ops@company.com'
        from: 'alertmanager@company.com'
        smarthost: 'smtp.company.com:587'

  - name: 'pagerduty'
    pagerduty_configs:
      - service_key: 'YOUR_PAGERDUTY_KEY'

  - name: 'slack-critical'
    slack_configs:
      - channel: '#payment-critical'
        title: '🚨 Critical Payment Alert'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'

  - name: 'slack-warning'
    slack_configs:
      - channel: '#payment-alerts'
        title: '⚠️ Payment Warning'
        text: '{{ range .Alerts }}{{ .Annotations.description }}{{ end }}'
```

---

## 5. Log Management

### 5.1 Structured Logging

**Log Format:**
```json
{
  "timestamp": "2025-11-11T10:30:45.123Z",
  "level": "info",
  "component": "payment-driver",
  "device_id": "terminal-001",
  "terminal_id": "TERM001",
  "transaction_id": "123456789012",
  "event": "transaction_processed",
  "duration_ms": 234,
  "amount": 10000,
  "currency": "SAR",
  "status": "approved",
  "response_code": "00"
}
```

### 5.2 Log Levels

- **DEBUG:** Detailed diagnostic information
- **INFO:** General operational events
- **WARN:** Warning events (non-critical)
- **ERROR:** Error events (needs attention)
- **FATAL:** Critical failures (service stop)

### 5.3 Log Aggregation (ELK Stack)

**Filebeat configuration:**
```yaml
filebeat.inputs:
  - type: log
    enabled: true
    paths:
      - /var/log/device-bridge/payment/*.log
    json.keys_under_root: true
    json.add_error_key: true
    fields:
      service: device-bridge
      component: payment

output.elasticsearch:
  hosts: ["elasticsearch:9200"]
  index: "device-bridge-payment-%{+yyyy.MM.dd}"

processors:
  - drop_fields:
      fields: ["host", "agent"]
  - add_cloud_metadata: ~
```

### 5.4 Log Queries (Kibana)

**Useful queries:**
```
# Failed transactions
status:failed AND component:payment

# High value transactions
amount:>100000 AND currency:SAR

# Declined transactions
response_code:51 OR response_code:54

# Slow transactions
duration_ms:>5000

# Connection errors
event:connection_error
```

---

## 6. Distributed Tracing

### 6.1 OpenTelemetry Integration

**Trace example:**
```go
import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (d *Driver) ProcessTransactionWithTracing(ctx context.Context, req TransactionRequest) (*TransactionResponse, error) {
	tracer := otel.Tracer("payment-driver")
	ctx, span := tracer.Start(ctx, "ProcessTransaction")
	defer span.End()

	span.SetAttributes(
		attribute.String("device.id", d.id),
		attribute.String("terminal.id", d.config.TerminalID),
		attribute.String("transaction.type", string(req.Type)),
		attribute.Int64("transaction.amount", req.Amount),
		attribute.String("transaction.currency", req.Currency),
	)

	// Process transaction
	resp, err := d.ProcessTransaction(ctx, req)

	if err != nil {
		span.RecordError(err)
		span.SetAttributes(attribute.String("error", "true"))
	} else {
		span.SetAttributes(
			attribute.String("response.code", resp.ResponseCode),
			attribute.Bool("response.success", resp.Success),
		)
	}

	return resp, err
}
```

---

## 7. Health Checks

### 7.1 Health Check Endpoints

**GET /health:**
```json
{
  "status": "healthy",
  "timestamp": "2025-11-11T10:30:45Z",
  "version": "2.0.0",
  "uptime_seconds": 86400,
  "checks": {
    "database": "healthy",
    "payment_terminals": "healthy",
    "audit_log": "healthy"
  }
}
```

**GET /health/live:**
```json
{
  "status": "alive"
}
```

**GET /health/ready:**
```json
{
  "status": "ready",
  "ready_for_traffic": true
}
```

### 7.2 Kubernetes Health Checks

```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

---

## 8. Performance Monitoring

### 8.1 Key Performance Indicators (KPIs)

| Metric | Target | Alert Threshold |
|--------|--------|----------------|
| Transaction Success Rate | > 99% | < 98% |
| P95 Latency | < 2s | > 5s |
| P99 Latency | < 5s | > 10s |
| Connection Success Rate | > 99.9% | < 99% |
| Decline Rate | < 10% | > 20% |
| Settlement Success Rate | 100% | < 100% |

### 8.2 SLI/SLO/SLA

**Service Level Indicators (SLI):**
- Transaction success rate
- Transaction latency (p95)
- Service availability

**Service Level Objectives (SLO):**
- 99.5% of transactions complete successfully
- 95% of transactions complete within 2 seconds
- 99.9% service uptime

**Service Level Agreement (SLA):**
- 99% uptime guarantee
- Maximum 1 hour downtime per month
- Response time: Critical incidents within 15 minutes

---

## 9. Incident Response

### 9.1 Incident Severity Levels

**P1 - Critical:**
- Payment service completely down
- Data breach or security incident
- PCI compliance violation

**P2 - High:**
- Error rate > 10%
- Latency > 10 seconds
- Multiple terminal failures

**P3 - Medium:**
- Error rate > 5%
- Latency > 5 seconds
- Single terminal failure

**P4 - Low:**
- Minor performance degradation
- Non-critical warnings

### 9.2 Runbook Example

**High Error Rate Runbook:**

1. **Verify:**
   ```bash
   # Check current error rate
   promtool query instant 'rate(payment_errors_total[5m])'

   # Check recent errors
   kubectl logs -l app=device-bridge --tail=100 | grep ERROR
   ```

2. **Identify:**
   - Check which terminals are affected
   - Check error types and codes
   - Review recent deployments

3. **Mitigate:**
   - If specific terminal: Disable terminal
   - If network issue: Check connectivity
   - If code issue: Rollback deployment

4. **Resolve:**
   - Fix root cause
   - Verify error rate returns to normal
   - Document incident

5. **Follow-up:**
   - Post-mortem analysis
   - Update runbook
   - Implement preventive measures

---

## 10. Best Practices

### 10.1 Monitoring Checklist

- [ ] Prometheus metrics exported
- [ ] Grafana dashboards configured
- [ ] Alert rules defined
- [ ] Alertmanager configured
- [ ] Log aggregation set up
- [ ] Distributed tracing enabled
- [ ] Health checks implemented
- [ ] SLI/SLO defined
- [ ] Runbooks documented
- [ ] On-call rotation established

### 10.2 Dashboard Best Practices

1. **Keep it simple:** Focus on key metrics
2. **Use consistent colors:** Red = bad, Green = good
3. **Add context:** Include thresholds and targets
4. **Time ranges:** Default to last 1 hour, allow customization
5. **Group logically:** Related metrics together

### 10.3 Alerting Best Practices

1. **Alert on symptoms, not causes**
2. **Reduce noise:** Avoid alert fatigue
3. **Use appropriate severity:** Critical = wake someone up
4. **Provide context:** Include troubleshooting steps
5. **Test alerts:** Regularly test alert delivery

---

## Appendix A: Quick Reference

### Prometheus Queries

```promql
# Transaction rate (TPS)
rate(payment_transactions_total[1m])

# Success rate
sum(rate(payment_transactions_total{status="success"}[5m]))
/ sum(rate(payment_transactions_total[5m]))

# P95 latency
histogram_quantile(0.95, rate(payment_transaction_duration_seconds_bucket[5m]))

# Error rate
rate(payment_errors_total[5m])

# Active connections
payment_connections_active
```

### Dashboard Links

- Payment Overview: `http://grafana/d/payment-overview`
- Transaction Analysis: `http://grafana/d/payment-transactions`
- Performance Metrics: `http://grafana/d/payment-performance`
- System Health: `http://grafana/d/payment-health`

---

**Document Version:** 1.0
**Last Updated:** 2025-11-11
**Next Review:** 2025-12-11
**Document Owner:** DevOps Team
