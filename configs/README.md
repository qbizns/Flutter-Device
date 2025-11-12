# Payment Terminal Driver - Configuration Files

This directory contains configuration files for the Payment Terminal Driver monitoring and deployment.

## Directory Structure

```
configs/
├── grafana/               # Grafana dashboards
│   ├── payment-overview.json
│   └── payment-transactions.json
├── prometheus/            # Prometheus configuration
│   ├── prometheus.yml
│   ├── payment_alerts.yml
│   └── payment_recording_rules.yml
├── payment/              # Payment driver configuration
│   ├── config.development.yaml
│   ├── config.staging.yaml
│   └── config.production.yaml
└── README.md             # This file
```

## Grafana Dashboards

### payment-overview.json
Comprehensive overview dashboard with:
- Transaction rate and success metrics
- P95 transaction duration
- Active connections status
- Error rates
- Transaction breakdown by type
- Terminal connection status table

**Import:** Grafana UI → Dashboards → Import → Upload JSON file

### payment-transactions.json
Detailed transaction monitoring dashboard with:
- Transaction rate by device and provider
- Success rate trends
- Duration percentiles (P95) by device and type
- Declined transactions by response code
- Transaction amount by currency
- Card scheme distribution

**Import:** Grafana UI → Dashboards → Import → Upload JSON file

### Dashboard Variables
Both dashboards support filtering by:
- `$device` - Filter by device ID (multi-select)
- `$provider` - Filter by payment provider (multi-select)

## Prometheus Configuration

### prometheus.yml
Main Prometheus configuration with:
- Scrape configurations for payment driver metrics
- High-frequency scraping for critical metrics
- Node exporter integration
- Alertmanager integration
- 15-day retention with 50GB storage limit

**Usage:**
```bash
prometheus --config.file=configs/prometheus/prometheus.yml
```

### payment_alerts.yml
Alerting rules covering:
- **Critical:** High error rate, terminal disconnected, timeout spikes, settlement failures
- **Warning:** Elevated errors, slow transactions, connection flapping, high void/refund rates
- **Performance:** High volume alerts, settlement due notifications
- **Security:** Auth failures, suspicious decline patterns
- **Business:** Low volume during business hours, unusual amounts

**Thresholds:**
- Error rate: >5% critical, >2% warning
- Success rate: <80% critical, <90% warning
- Transaction duration: >5s warning
- Settlement: >20 hours since last settlement

### payment_recording_rules.yml
Pre-computed queries for dashboard performance:
- Transaction rates (by device, provider, type)
- Success rates and error ratios
- Duration percentiles (P50, P95, P99)
- Connection metrics
- Settlement metrics
- Hourly aggregations
- SLI/SLO tracking

**Benefits:**
- Faster dashboard loading
- Reduced Prometheus query load
- Consistent metric definitions

## Payment Driver Configuration

Configuration files for different environments with appropriate security and monitoring settings.

### config.development.yaml
Development environment settings:
- **Simulator:** Enabled (no real terminals needed)
- **TLS:** Disabled for local development
- **Logging:** Debug level, console output
- **Monitoring:** All features enabled including profiling
- **Settlement:** Manual (auto-settle disabled)
- **Security:** Rate limiting disabled, debug endpoints enabled

**Usage:**
```bash
device-bridge --config=configs/payment/config.development.yaml
```

### config.staging.yaml
Staging environment settings:
- **Terminals:** Real staging terminals
- **TLS:** Enabled with certificate validation
- **Logging:** Info level, file output with rotation
- **Monitoring:** Full monitoring + 10% tracing
- **Settlement:** Automated daily at 23:00
- **Security:** Rate limiting enabled, API key auth

**Environment Variables Required:**
```bash
MADA_TERMINAL_ID=xxx
MADA_MERCHANT_ID=xxx
MADA_AUTH_KEY=xxx
KNET_TERMINAL_ID=xxx
KNET_MERCHANT_ID=xxx
KNET_AUTH_KEY=xxx
```

**Usage:**
```bash
export MADA_TERMINAL_ID="..."
export MADA_MERCHANT_ID="..."
# ... set other variables
device-bridge --config=configs/payment/config.staging.yaml
```

### config.production.yaml
Production environment settings:
- **Terminals:** Production terminal endpoints
- **TLS:** TLS 1.3 only, certificate pinning, mTLS auth
- **Logging:** Warn level, comprehensive audit logs (1 year retention)
- **Monitoring:** Full observability, 1% tracing for performance
- **Settlement:** Automated with retry and notifications
- **Security:** Strict rate limiting, IP whitelist, PCI DSS compliance
- **HA:** Circuit breaker, failover endpoints, graceful shutdown

**Environment Variables Required:**
```bash
PAYMENT_TERMINAL_HOST=xxx
MADA_TERMINAL_ID=xxx
MADA_MERCHANT_ID=xxx
MADA_AUTH_KEY=xxx
MADA_ENDPOINT=xxx
KNET_TERMINAL_ID=xxx
KNET_MERCHANT_ID=xxx
KNET_AUTH_KEY=xxx
KNET_ENDPOINT=xxx
AUDIT_BACKUP_ENDPOINT=xxx
JAEGER_ENDPOINT=xxx
LOGSTASH_ENDPOINT=xxx
SETTLEMENT_NOTIFICATION_ENDPOINT=xxx
ALLOWED_IP_1=xxx
ALLOWED_IP_2=xxx
ENCRYPTION_KEY_ID=xxx
FAILOVER_ENDPOINT_1=xxx
FAILOVER_ENDPOINT_2=xxx
PANIC_NOTIFICATION_ENDPOINT=xxx
ALERTMANAGER_ENDPOINT=xxx
```

**Security Notes:**
- Never commit credentials to version control
- Use secrets manager (Vault, AWS Secrets Manager, etc.)
- Rotate credentials regularly
- Follow principle of least privilege

**Usage with Secrets:**
```bash
# Load secrets from secrets manager
eval $(secrets-manager export --env=production --app=payment-driver)

# Start with production config
device-bridge --config=configs/payment/config.production.yaml
```

## Configuration Best Practices

### 1. Environment Separation
- Never use production credentials in non-production environments
- Use separate terminals/merchants for each environment
- Maintain separate monitoring infrastructure

### 2. Secrets Management
- Store all credentials in a secrets manager
- Use environment variables or mounted secrets
- Rotate credentials regularly (90 days recommended)
- Never log or expose credentials

### 3. Monitoring
- Set up Prometheus before deploying
- Configure Alertmanager for critical alerts
- Import Grafana dashboards
- Test alerts in staging first

### 4. TLS/SSL
- Use TLS 1.3 in production
- Enable certificate pinning
- Rotate certificates before expiry
- Use mTLS for authentication

### 5. Audit Logging
- Enable audit logging in all environments
- Retain logs per compliance requirements (1 year minimum for PCI DSS)
- Backup logs to remote storage
- Monitor audit log integrity

### 6. Performance Tuning
- Adjust worker pool size based on load
- Monitor memory usage and set limits
- Use connection pooling
- Enable recording rules for complex queries

### 7. Testing Configuration
Before deploying new configuration:
```bash
# Validate YAML syntax
yamllint configs/payment/config.production.yaml

# Test configuration loading
device-bridge --config=configs/payment/config.production.yaml --validate

# Check Prometheus rules
promtool check rules configs/prometheus/payment_alerts.yml
promtool check config configs/prometheus/prometheus.yml
```

## Deployment

### Docker Compose Example
```yaml
version: '3.8'

services:
  device-bridge:
    image: device-bridge:latest
    volumes:
      - ./configs/payment/config.production.yaml:/etc/device-bridge/config.yaml:ro
      - /run/secrets:/run/secrets:ro
    environment:
      - PAYMENT_TERMINAL_HOST=${PAYMENT_TERMINAL_HOST}
      # ... other env vars
    ports:
      - "9090:9090"  # Prometheus metrics
      - "8080:8080"  # Health checks

  prometheus:
    image: prom/prometheus:latest
    volumes:
      - ./configs/prometheus:/etc/prometheus:ro
      - prometheus-data:/prometheus
    ports:
      - "9090:9090"

  grafana:
    image: grafana/grafana:latest
    volumes:
      - ./configs/grafana:/etc/grafana/provisioning/dashboards:ro
      - grafana-data:/var/lib/grafana
    ports:
      - "3000:3000"

volumes:
  prometheus-data:
  grafana-data:
```

### Kubernetes Example
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: payment-driver-config
data:
  config.yaml: |
    # Include production config here
    # Or mount from external ConfigMap

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: device-bridge
spec:
  replicas: 3
  template:
    spec:
      containers:
      - name: device-bridge
        image: device-bridge:latest
        volumeMounts:
        - name: config
          mountPath: /etc/device-bridge
          readOnly: true
        - name: secrets
          mountPath: /run/secrets
          readOnly: true
        ports:
        - containerPort: 9090
          name: metrics
        - containerPort: 8080
          name: health
      volumes:
      - name: config
        configMap:
          name: payment-driver-config
      - name: secrets
        secret:
          secretName: payment-driver-secrets
```

## Support

For issues or questions:
- See main project documentation: `/docs`
- Security guide: `/docs/PAYMENT_SECURITY_GUIDE.md`
- Monitoring guide: `/docs/PAYMENT_MONITORING_GUIDE.md`
- Architecture: `/docs/PAYMENT_ARCHITECTURE.md`

## Version History

- **1.0** (2025-11-11): Initial configuration files
  - Grafana dashboards (overview, transactions)
  - Prometheus configuration with alerts and recording rules
  - Environment-specific payment driver configs
