# Device Bridge v2 - Deployment Guide

**Version:** 2.0
**Date:** 2025-11-11
**Status:** ✅ Production Ready

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Deployment Options](#deployment-options)
4. [Linux (systemd) Deployment](#linux-systemd-deployment)
5. [Windows Service Deployment](#windows-service-deployment)
6. [macOS Deployment](#macos-deployment)
7. [Docker Deployment](#docker-deployment)
8. [Kubernetes Deployment](#kubernetes-deployment)
9. [Post-Deployment Verification](#post-deployment-verification)
10. [Production Checklist](#production-checklist)

---

## Overview

Device Bridge v2 can be deployed on multiple platforms:
- **Linux** - systemd service
- **Windows** - Windows Service
- **macOS** - launchd daemon
- **Docker** - Containerized deployment with Docker Compose
- **Kubernetes** - Cloud-native orchestrated deployment

---

## Prerequisites

### Hardware Requirements

| Component | Minimum | Recommended |
|-----------|---------|-------------|
| CPU | 2 cores | 4 cores |
| RAM | 2 GB | 4 GB |
| Disk | 10 GB | 50 GB |
| Network | 100 Mbps | 1 Gbps |

### Software Requirements

- **Go** 1.21+ (for building from source)
- **Git** (for version control)
- **Operating System:**
  - Linux: Ubuntu 20.04+, RHEL 8+, Debian 11+
  - Windows: Windows Server 2019+, Windows 10/11
  - macOS: macOS 11+
  - Docker: Docker 20.10+
  - Kubernetes: 1.24+

### Network Requirements

- **Inbound:**
  - Port 8080 (API/HTTP)
  - Port 9090 (Prometheus metrics)
  - Port 3000 (Grafana, if using monitoring stack)

- **Outbound:**
  - HTTPS (443) for payment networks
  - DNS (53) for name resolution

### Credentials Required

- Mada credentials (terminal_id, merchant_id, auth_key)
- KNET credentials (terminal_id, merchant_id, auth_key)
- TLS certificates (for production)
- API keys (for client authentication)

---

## Deployment Options

### Quick Comparison

| Method | Best For | Complexity | HA Support | Scaling |
|--------|----------|------------|------------|---------|
| systemd | Linux servers | Low | Manual | Manual |
| Windows Service | Windows servers | Low | Manual | Manual |
| macOS launchd | macOS workstations | Low | No | No |
| Docker Compose | Development, single-host | Medium | Limited | Manual |
| Kubernetes | Production, cloud | High | Yes | Automatic |

---

## Linux (systemd) Deployment

### 1. Build the Binary

```bash
# Clone repository
git clone https://github.com/Macber-eg/Flutter-Device.git
cd Flutter-Device

# Build for Linux
make build
# or manually:
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o device-bridge ./cmd/device-bridge
```

### 2. Run Installation Script

```bash
cd deploy/systemd
sudo ./install.sh
```

The script will:
- Create `devicebridge` user and group
- Create directories (`/opt/device-bridge`, `/etc/device-bridge`, etc.)
- Install binary to `/usr/local/bin/device-bridge`
- Install systemd service
- Configure USB device access

### 3. Configure

```bash
# Edit configuration
sudo nano /etc/device-bridge/config.yaml

# Edit environment variables
sudo nano /etc/device-bridge/environment
```

**Example environment file:**
```bash
MADA_TERMINAL_ID=MADA_TERM_001
MADA_MERCHANT_ID=MADA_MERCH_12345
MADA_AUTH_KEY=your_secret_key_here

KNET_TERMINAL_ID=KNET_TERM_001
KNET_MERCHANT_ID=KNET_MERCH_67890
KNET_AUTH_KEY=your_secret_key_here
```

### 4. Start Service

```bash
# Start service
sudo systemctl start device-bridge

# Enable on boot
sudo systemctl enable device-bridge

# Check status
sudo systemctl status device-bridge

# View logs
sudo journalctl -u device-bridge -f
```

### 5. Verify

```bash
# Health check
curl http://localhost:8080/health

# List devices
curl http://localhost:8080/api/devices

# Metrics
curl http://localhost:9090/metrics
```

---

## Windows Service Deployment

### 1. Build for Windows

```powershell
# On Windows with Go installed
$env:GOOS="windows"
$env:GOARCH="amd64"
$env:CGO_ENABLED="1"
go build -o device-bridge.exe .\cmd\device-bridge
```

Or download pre-built binary.

### 2. Run Installation Script

```powershell
# Open PowerShell as Administrator
cd deploy\windows
.\install-service.ps1
```

The script will:
- Create installation directory (`C:\Program Files\Device Bridge`)
- Copy binary and configuration
- Create Windows Service
- Add firewall rules

### 3. Configure

```powershell
# Edit configuration
notepad "C:\Program Files\Device Bridge\config\config.yaml"
```

Set environment variables:
```powershell
[System.Environment]::SetEnvironmentVariable("MADA_TERMINAL_ID", "YOUR_VALUE", "Machine")
[System.Environment]::SetEnvironmentVariable("MADA_MERCHANT_ID", "YOUR_VALUE", "Machine")
# ... repeat for all credentials
```

### 4. Start Service

```powershell
# Start service
Start-Service -Name DeviceBridge

# Check status
Get-Service -Name DeviceBridge

# View logs
Get-Content "C:\Program Files\Device Bridge\logs\device-bridge.log" -Tail 50 -Wait
```

### 5. Verify

```powershell
# Health check
Invoke-WebRequest -Uri http://localhost:8080/health

# Metrics
Invoke-WebRequest -Uri http://localhost:9090/metrics
```

---

## macOS Deployment

### 1. Build for macOS

```bash
# Build
make build-macos
# or
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -o device-bridge ./cmd/device-bridge
```

### 2. Run Installation Script

```bash
cd deploy/macos
sudo ./install.sh
```

### 3. Configure

```bash
# Edit configuration
sudo nano /usr/local/etc/device-bridge/config.yaml
```

### 4. Load and Start Service

```bash
# Load service
sudo launchctl load /Library/LaunchDaemons/com.devicebridge.service.plist

# Start service
sudo launchctl start com.devicebridge.service

# Check status
sudo launchctl list | grep devicebridge

# View logs
tail -f /usr/local/var/log/device-bridge/stdout.log
```

---

## Docker Deployment

### 1. Prepare Configuration

```bash
cd deploy/docker

# Copy environment template
cp .env.example .env

# Edit environment
nano .env
```

### 2. Copy Configuration Files

```bash
# Copy configuration
cp ../../configs/payment/config.production.yaml config.yaml

# Copy Prometheus configuration
cp ../../configs/prometheus/prometheus.yml .
cp ../../configs/prometheus/payment_alerts.yml .
cp ../../configs/prometheus/payment_recording_rules.yml .

# Copy Grafana dashboards
mkdir -p grafana
cp ../../configs/grafana/*.json grafana/
```

### 3. Build and Start

```bash
# Build image
docker-compose build

# Start stack
docker-compose up -d

# View logs
docker-compose logs -f device-bridge

# Check status
docker-compose ps
```

### 4. Verify

```bash
# Health check
curl http://localhost:8080/health

# Grafana (default: admin/admin)
open http://localhost:3000

# Prometheus
open http://localhost:9091
```

### 5. Managing the Stack

```bash
# Stop
docker-compose down

# Restart
docker-compose restart device-bridge

# Update
docker-compose pull
docker-compose up -d

# View logs
docker-compose logs -f
```

---

## Kubernetes Deployment

### 1. Prerequisites

- Kubernetes cluster (1.24+)
- kubectl configured
- Helm 3+ (optional, for cert-manager)
- Ingress controller (nginx, traefik, etc.)
- cert-manager (for TLS certificates)

### 2. Prepare Cluster

```bash
# Create namespace
kubectl create namespace device-bridge

# Install cert-manager (if not already installed)
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Create ClusterIssuer for Let's Encrypt
cat <<EOF | kubectl apply -f -
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: ops@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
EOF
```

### 3. Create Secrets

```bash
# Create payment credentials secret
kubectl create secret generic payment-credentials \
  --from-literal=mada-terminal-id=MADA_TERM_001 \
  --from-literal=mada-merchant-id=MADA_MERCH_12345 \
  --from-literal=mada-auth-key=SECRET_KEY \
  --from-literal=knet-terminal-id=KNET_TERM_001 \
  --from-literal=knet-merchant-id=KNET_MERCH_67890 \
  --from-literal=knet-auth-key=SECRET_KEY \
  -n device-bridge
```

### 4. Build and Push Image

```bash
# Build Docker image
docker build -t your-registry.com/device-bridge:v2.0 -f deploy/docker/Dockerfile .

# Push to registry
docker push your-registry.com/device-bridge:v2.0
```

### 5. Update Manifests

Edit `deploy/kubernetes/deployment.yaml` and update image:
```yaml
containers:
  - name: device-bridge
    image: your-registry.com/device-bridge:v2.0
```

### 6. Deploy

```bash
cd deploy/kubernetes

# Deploy all resources
kubectl apply -f config.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml
kubectl apply -f autoscaling.yaml
kubectl apply -f ingress.yaml

# Check deployment
kubectl get all -n device-bridge

# Watch rollout
kubectl rollout status deployment/device-bridge -n device-bridge
```

### 7. Verify

```bash
# Check pods
kubectl get pods -n device-bridge

# Check logs
kubectl logs -f -l app=device-bridge -n device-bridge

# Check service
kubectl get svc -n device-bridge

# Test health endpoint
kubectl run -it --rm debug --image=curlimages/curl --restart=Never -- \
  curl http://device-bridge.device-bridge.svc.cluster.local:8080/health
```

### 8. Access Application

```bash
# Get Ingress address
kubectl get ingress -n device-bridge

# Test API (update domain)
curl https://api.device-bridge.example.com/health
```

---

## Post-Deployment Verification

### Health Checks

```bash
# 1. Service health
curl http://localhost:8080/health
# Expected: {"status":"healthy","timestamp":"..."}

# 2. Readiness
curl http://localhost:8080/health/ready
# Expected: {"status":"ready"}

# 3. Liveness
curl http://localhost:8080/health/live
# Expected: {"status":"alive"}

# 4. List devices
curl http://localhost:8080/api/devices
# Expected: {"devices":[...]}
```

### Smoke Tests

```bash
# Test payment transaction (simulator)
curl -X POST http://localhost:8080/api/devices/payment-001/transaction \
  -H "Content-Type: application/json" \
  -H "X-API-Key: your-api-key" \
  -d '{
    "type": "sale",
    "amount": 10000,
    "currency": "SAR"
  }'
```

### Metrics Verification

```bash
# Check Prometheus metrics
curl http://localhost:9090/metrics | grep payment_

# Expected metrics:
# payment_transactions_total
# payment_transaction_duration_seconds
# payment_errors_total
# payment_connection_status
```

### Log Verification

Check logs for:
- ✅ No ERROR or FATAL messages
- ✅ Successful connection messages
- ✅ Configuration loaded successfully
- ✅ API server started

---

## Production Checklist

Before going to production, verify:

### Security
- [ ] TLS/SSL certificates installed and valid
- [ ] API keys configured and documented
- [ ] Payment credentials secured (secrets management)
- [ ] Firewall rules configured
- [ ] CORS configured with specific origins (no wildcard)
- [ ] Rate limiting enabled
- [ ] IP whitelisting configured (if applicable)
- [ ] Security headers configured
- [ ] Audit logging enabled

### Monitoring
- [ ] Prometheus scraping metrics successfully
- [ ] Grafana dashboards imported and displaying data
- [ ] Alert rules configured in Alertmanager
- [ ] Alert notifications configured (email, Slack, PagerDuty)
- [ ] Log aggregation configured (ELK, Splunk, etc.)
- [ ] Distributed tracing configured (Jaeger, if enabled)

### High Availability
- [ ] Multiple instances running (3+ for Kubernetes)
- [ ] Health checks configured
- [ ] Auto-scaling configured (Kubernetes HPA)
- [ ] Load balancer configured
- [ ] Backup procedures documented
- [ ] Disaster recovery plan documented

### Configuration
- [ ] Production configuration reviewed
- [ ] Environment-specific settings applied
- [ ] Resource limits set appropriately
- [ ] Timeouts configured
- [ ] Retry logic configured
- [ ] Settlement schedule configured

### Testing
- [ ] Smoke tests passed
- [ ] Integration tests with real terminals passed
- [ ] Load testing completed
- [ ] Failover testing completed
- [ ] Backup/restore tested

### Documentation
- [ ] Deployment documented
- [ ] Runbook created and reviewed
- [ ] Emergency contacts updated
- [ ] On-call procedures documented
- [ ] Rollback procedures documented

### Compliance
- [ ] PCI DSS compliance verified
- [ ] Data retention policies configured
- [ ] Privacy policies reviewed
- [ ] Audit requirements met

### Operations
- [ ] Monitoring dashboards accessible
- [ ] On-call rotation configured
- [ ] Incident response procedures documented
- [ ] Maintenance windows scheduled
- [ ] Change management process established

---

## Troubleshooting

### Service Won't Start

**Check logs:**
```bash
# systemd
sudo journalctl -u device-bridge -n 50

# Kubernetes
kubectl logs -l app=device-bridge -n device-bridge

# Docker
docker-compose logs device-bridge
```

**Common issues:**
- Configuration file syntax error → Validate YAML
- Missing credentials → Check secrets/environment variables
- Port already in use → Check `lsof -i :8080`
- Permission denied → Check file/directory permissions

### Cannot Connect to Payment Terminal

**Check:**
1. Network connectivity: `ping terminal.example.com`
2. Firewall rules allow outbound HTTPS
3. TLS certificates are valid
4. Terminal credentials are correct
5. Terminal is powered on and online

### High Memory Usage

**Actions:**
1. Check for memory leaks in logs
2. Review resource limits
3. Scale up if needed
4. Check for large payloads

---

## Support

For additional help:
- **Documentation:** [OPERATIONS_RUNBOOK.md](OPERATIONS_RUNBOOK.md)
- **Security:** [PAYMENT_SECURITY_GUIDE.md](PAYMENT_SECURITY_GUIDE.md)
- **Monitoring:** [PAYMENT_MONITORING_GUIDE.md](PAYMENT_MONITORING_GUIDE.md)
- **Arabic Guide:** [PAYMENT_TERMINAL_SETUP_AR.md](PAYMENT_TERMINAL_SETUP_AR.md)
- **Browser Integration:** [BROWSER_CORS_SECURITY.md](BROWSER_CORS_SECURITY.md)

---

**Document Version:** 2.0
**Last Updated:** 2025-11-11
**Status:** ✅ Production Ready