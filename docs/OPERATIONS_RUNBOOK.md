# Device Bridge v2 - Operations Runbook

**Version:** 2.0
**Date:** 2025-11-11
**Status:** ✅ Production Ready

---

## Table of Contents

1. [Overview](#overview)
2. [System Architecture](#system-architecture)
3. [Deployment Procedures](#deployment-procedures)
4. [Monitoring & Alerting](#monitoring--alerting)
5. [Common Operations](#common-operations)
6. [Troubleshooting](#troubleshooting)
7. [Incident Response](#incident-response)
8. [Maintenance Procedures](#maintenance-procedures)
9. [Backup & Recovery](#backup--recovery)
10. [Security Operations](#security-operations)

---

## Overview

### Purpose

This runbook provides operational procedures for Device Bridge v2 in production environments.

### Service Level Objectives (SLOs)

| Metric | Target | Measurement Period |
|--------|--------|-------------------|
| Availability | 99.9% | Monthly |
| Transaction Success Rate | >95% | Daily |
| API Response Time (P95) | <2s | Hourly |
| Payment Transaction Time (P95) | <5s | Hourly |
| Error Rate | <1% | Hourly |

### On-Call Responsibilities

- Monitor alerts and respond within SLA
- Investigate and resolve incidents
- Escalate critical issues
- Document all incidents
- Perform routine maintenance

---

## System Architecture

### Components

```
┌─────────────────┐
│   Load Balancer │
└────────┬────────┘
         │
    ┌────┴────┬────────┬────────┐
    │         │        │        │
┌───┴──┐  ┌───┴──┐ ┌───┴──┐ ┌───┴──┐
│ DB-1 │  │ DB-2 │ │ DB-3 │ │ DB-N │  Device Bridge Instances
└──┬───┘  └──┬───┘ └──┬───┘ └──┬───┘
   │         │        │        │
   └─────────┴────────┴────────┘
             │
    ┌────────┴────────┐
    │   Prometheus    │  Metrics
    └────────┬────────┘
             │
    ┌────────┴────────┐
    │     Grafana     │  Dashboards
    └─────────────────┘
```

### Dependencies

- **External:** Payment networks (Mada, KNET), Certificate Authority
- **Internal:** Prometheus, Grafana, Logstash, Database
- **Hardware:** USB devices (scanners, printers), Serial devices (scales), Payment terminals

---

## Deployment Procedures

### Pre-Deployment Checklist

- [ ] Code review completed and approved
- [ ] All tests passing (unit, integration, security)
- [ ] Security scan completed (no critical vulnerabilities)
- [ ] Configuration updated and reviewed
- [ ] Secrets rotated (if applicable)
- [ ] Database migrations tested
- [ ] Rollback plan documented
- [ ] Monitoring dashboards updated
- [ ] Alert rules configured
- [ ] Documentation updated
- [ ] Stakeholders notified

### Deployment Steps (Kubernetes)

```bash
# 1. Backup current configuration
kubectl get all -n device-bridge -o yaml > backup-$(date +%Y%m%d-%H%M%S).yaml

# 2. Apply configuration changes
kubectl apply -f deploy/kubernetes/config.yaml

# 3. Deploy new version
kubectl set image deployment/device-bridge \
  device-bridge=device-bridge:v2.0-$(git rev-parse --short HEAD) \
  -n device-bridge

# 4. Monitor rollout
kubectl rollout status deployment/device-bridge -n device-bridge

# 5. Verify pods are healthy
kubectl get pods -n device-bridge
kubectl logs -f deployment/device-bridge -n device-bridge

# 6. Run smoke tests
curl -f https://api.device-bridge.example.com/health
curl -f https://api.device-bridge.example.com/api/devices

# 7. Monitor metrics
# Check Grafana dashboards for anomalies
# Check error rates in Prometheus
```

### Rollback Procedure

```bash
# Immediate rollback
kubectl rollout undo deployment/device-bridge -n device-bridge

# Rollback to specific revision
kubectl rollout history deployment/device-bridge -n device-bridge
kubectl rollout undo deployment/device-bridge --to-revision=N -n device-bridge

# Verify rollback
kubectl rollout status deployment/device-bridge -n device-bridge
```

---

## Monitoring & Alerting

### Key Metrics to Monitor

**Application Metrics:**
- `payment_transactions_total` - Total transactions
- `payment_transaction_duration_seconds` - Transaction latency
- `payment_errors_total` - Error count
- `payment_connection_status` - Terminal connection status
- `payment_active_connections` - Active connections

**System Metrics:**
- CPU utilization
- Memory usage
- Disk I/O
- Network throughput
- Pod restarts

### Critical Alerts

| Alert | Threshold | Action |
|-------|-----------|--------|
| High Error Rate | >5% for 2min | Investigate immediately |
| Terminal Disconnected | 1min | Check terminal connection |
| Very Low Success Rate | <80% for 5min | Emergency response |
| Service Down | Health check fails | Page on-call |
| High Latency | P95 >5s | Investigate performance |

### Accessing Dashboards

**Grafana:**
```
URL: https://grafana.example.com
Dashboard: Payment Terminal - Overview
Dashboard: Payment Terminal - Transaction Details
```

**Prometheus:**
```
URL: https://prometheus.example.com
Query: rate(payment_transactions_total[5m])
```

---

## Common Operations

### Restarting the Service

**systemd (Linux):**
```bash
sudo systemctl restart device-bridge
sudo systemctl status device-bridge
sudo journalctl -u device-bridge -f
```

**Windows:**
```powershell
Restart-Service -Name DeviceBridge
Get-Service -Name DeviceBridge
```

**Kubernetes:**
```bash
kubectl rollout restart deployment/device-bridge -n device-bridge
```

### Viewing Logs

**systemd:**
```bash
# Real-time logs
sudo journalctl -u device-bridge -f

# Last 100 lines
sudo journalctl -u device-bridge -n 100

# Filter by time
sudo journalctl -u device-bridge --since "2025-11-11 10:00:00"

# Filter by level
sudo journalctl -u device-bridge -p err
```

**Kubernetes:**
```bash
# Real-time logs
kubectl logs -f deployment/device-bridge -n device-bridge

# All pods
kubectl logs -l app=device-bridge -n device-bridge --tail=100

# Previous container (crashed)
kubectl logs pod-name --previous -n device-bridge
```

### Configuration Updates

**Kubernetes:**
```bash
# Edit ConfigMap
kubectl edit configmap device-bridge-config -n device-bridge

# Restart pods to pick up changes
kubectl rollout restart deployment/device-bridge -n device-bridge
```

**systemd:**
```bash
# Edit config
sudo nano /etc/device-bridge/config.yaml

# Restart service
sudo systemctl restart device-bridge
```

### Scaling

**Kubernetes:**
```bash
# Manual scaling
kubectl scale deployment/device-bridge --replicas=5 -n device-bridge

# Check HPA status
kubectl get hpa -n device-bridge

# Adjust HPA
kubectl edit hpa device-bridge -n device-bridge
```

---

## Troubleshooting

### High Error Rate

**Symptoms:**
- Alert: `PaymentHighErrorRate`
- Dashboard shows >5% error rate

**Investigation:**
```bash
# 1. Check error types
kubectl logs -l app=device-bridge -n device-bridge | grep ERROR | tail -50

# 2. Query Prometheus
# rate(payment_errors_total[5m]) by (error_type)

# 3. Check terminal connectivity
kubectl logs -l app=device-bridge -n device-bridge | grep "connection"
```

**Common Causes:**
- Network connectivity issues
- Payment network downtime
- Configuration errors
- Certificate expiration

**Resolution:**
1. Verify network connectivity to payment terminals
2. Check payment network status pages
3. Review recent configuration changes
4. Check TLS certificates expiration

### Terminal Disconnected

**Symptoms:**
- Alert: `PaymentTerminalDisconnected`
- `payment_connection_status` = 0

**Investigation:**
```bash
# Check device status
curl https://api.device-bridge.example.com/api/devices/payment-001/status

# Check logs for disconnect reason
kubectl logs -l app=device-bridge -n device-bridge | grep "disconnect"
```

**Resolution:**
1. Physical check: Terminal powered on, cables connected
2. Network check: Terminal has network connectivity
3. Restart terminal if necessary
4. Restart Device Bridge if persistent

### High Latency

**Symptoms:**
- Alert: `PaymentSlowTransactions`
- P95 >5 seconds

**Investigation:**
```bash
# Check resource usage
kubectl top pods -n device-bridge

# Check database performance
# Query slow queries

# Check network latency
ping payment-terminal.example.com
```

**Resolution:**
1. Scale up if CPU/memory constrained
2. Optimize database queries if applicable
3. Check network path to payment terminals
4. Review transaction logs for bottlenecks

### Service Won't Start

**Investigation:**
```bash
# systemd
sudo systemctl status device-bridge
sudo journalctl -u device-bridge -n 50

# Kubernetes
kubectl describe pod <pod-name> -n device-bridge
kubectl logs <pod-name> -n device-bridge
```

**Common Causes:**
- Configuration syntax error
- Missing secrets/credentials
- Port already in use
- Insufficient permissions

**Resolution:**
1. Validate configuration: `yamllint config.yaml`
2. Check secret exists: `kubectl get secret payment-credentials -n device-bridge`
3. Check port availability: `sudo lsof -i :8080`
4. Review file permissions

---

## Incident Response

### Severity Levels

| Severity | Description | Response Time | Example |
|----------|-------------|---------------|---------|
| **P1 - Critical** | Service down, major business impact | 15 minutes | All payments failing |
| **P2 - High** | Degraded performance, partial outage | 1 hour | One terminal down |
| **P3 - Medium** | Minor issues, workaround available | 4 hours | Slow responses |
| **P4 - Low** | Cosmetic issues, no business impact | Next business day | UI glitch |

### Incident Response Process

1. **Acknowledge** - Acknowledge alert within SLA
2. **Assess** - Determine severity and impact
3. **Communicate** - Notify stakeholders
4. **Investigate** - Gather logs, metrics, traces
5. **Mitigate** - Implement temporary fix
6. **Resolve** - Implement permanent fix
7. **Document** - Create post-incident report
8. **Review** - Conduct post-mortem

### Communication Template

```
INCIDENT: [Title]
SEVERITY: [P1/P2/P3/P4]
STATUS: [Investigating/Identified/Monitoring/Resolved]
IMPACT: [Description]
CURRENT ACTIONS: [What we're doing]
NEXT UPDATE: [Time]
```

### Post-Incident Report Template

```markdown
# Incident Report - [Date]

## Summary
[Brief description]

## Timeline
- HH:MM - Event occurred
- HH:MM - Alert triggered
- HH:MM - Engineer acknowledged
- HH:MM - Root cause identified
- HH:MM - Fix deployed
- HH:MM - Incident resolved

## Impact
- Duration: X hours Y minutes
- Affected users: N
- Failed transactions: M

## Root Cause
[Detailed explanation]

## Resolution
[How it was fixed]

## Action Items
- [ ] Item 1 (Owner, Deadline)
- [ ] Item 2 (Owner, Deadline)

## Lessons Learned
[What we learned]
```

---

## Maintenance Procedures

### Certificate Renewal

**Timeline:** 30 days before expiration

```bash
# Check certificate expiration
openssl s_client -connect api.device-bridge.example.com:443 | openssl x509 -noout -dates

# Renew certificate (Let's Encrypt)
certbot renew

# Update Kubernetes secret
kubectl create secret tls device-bridge-tls \
  --cert=fullchain.pem \
  --key=privkey.pem \
  --dry-run=client -o yaml | kubectl apply -f -

# Restart to pick up new cert
kubectl rollout restart deployment/device-bridge -n device-bridge
```

### Database Maintenance

```bash
# Backup database
pg_dump devicebridge > backup-$(date +%Y%m%d).sql

# Vacuum database
psql -c "VACUUM ANALYZE;"

# Check table sizes
psql -c "SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size FROM pg_tables ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC LIMIT 10;"
```

### Log Rotation

**systemd (automatic via journald):**
```bash
# Check journal size
sudo journalctl --disk-usage

# Vacuum old logs
sudo journalctl --vacuum-time=30d
sudo journalctl --vacuum-size=1G
```

### Dependency Updates

**Monthly security patches:**
```bash
# Update base image
docker pull alpine:latest

# Rebuild
docker build -t device-bridge:latest .

# Test in staging
# Deploy to production (follow deployment procedure)
```

---

## Backup & Recovery

### What to Backup

- [ ] Configuration files
- [ ] Secrets and credentials (encrypted)
- [ ] Database (if applicable)
- [ ] Audit logs
- [ ] Certificates and keys

### Backup Procedures

**Configuration Backup:**
```bash
# Kubernetes
kubectl get all -n device-bridge -o yaml > backup-config-$(date +%Y%m%d).yaml
kubectl get configmap,secret -n device-bridge -o yaml > backup-secrets-$(date +%Y%m%d).yaml

# Encrypt secrets backup
gpg --encrypt --recipient ops@example.com backup-secrets-$(date +%Y%m%d).yaml
```

**Audit Log Backup:**
```bash
# Archive logs
tar -czf audit-logs-$(date +%Y%m).tar.gz /var/log/device-bridge/audit*.log

# Upload to S3
aws s3 cp audit-logs-$(date +%Y%m).tar.gz s3://backups/device-bridge/
```

### Recovery Procedures

**Disaster Recovery:**
```bash
# 1. Provision new cluster/server
# 2. Restore configuration
kubectl apply -f backup-config-YYYYMMDD.yaml

# 3. Restore secrets
gpg --decrypt backup-secrets-YYYYMMDD.yaml.gpg | kubectl apply -f -

# 4. Deploy application
kubectl apply -f deploy/kubernetes/

# 5. Verify
kubectl get all -n device-bridge
curl https://api.device-bridge.example.com/health
```

---

## Security Operations

### Security Monitoring

**Daily:**
- Review failed authentication attempts
- Check for unusual traffic patterns
- Verify certificate validity

**Weekly:**
- Review access logs
- Check for CVEs affecting dependencies
- Audit user access

**Monthly:**
- Security patch updates
- Access control review
- Penetration testing review

### Incident Handling

**Security Incident Detected:**
1. **Isolate:** Stop affected instances
2. **Preserve:** Capture logs, memory dumps
3. **Analyze:** Determine breach scope
4. **Contain:** Patch vulnerability
5. **Eradicate:** Remove malicious code
6. **Recover:** Restore from clean backup
7. **Report:** Notify stakeholders, authorities

### Credential Rotation

**Quarterly rotation:**
```bash
# Generate new credentials
# Update secrets
kubectl create secret generic payment-credentials \
  --from-literal=mada-auth-key=$NEW_KEY \
  --dry-run=client -o yaml | kubectl apply -f -

# Rolling restart
kubectl rollout restart deployment/device-bridge -n device-bridge

# Verify
kubectl logs -l app=device-bridge -n device-bridge | grep "authentication"
```

---

## Emergency Contacts

| Role | Name | Phone | Email | Availability |
|------|------|-------|-------|--------------|
| On-Call Engineer | [Name] | +XXX | [email] | 24/7 |
| Engineering Manager | [Name] | +XXX | [email] | Business hours |
| Security Team | [Name] | +XXX | [email] | 24/7 |
| Mada Support | - | +XXX | [email] | Business hours |
| KNET Support | - | +XXX | [email] | Business hours |

---

## Appendix

### Useful Commands Quick Reference

```bash
# Health check
curl https://api.device-bridge.example.com/health

# Metrics
curl https://api.device-bridge.example.com:9090/metrics

# List devices
curl https://api.device-bridge.example.com/api/devices

# Check service status (systemd)
sudo systemctl status device-bridge

# View logs (Kubernetes)
kubectl logs -f -l app=device-bridge -n device-bridge

# Scale up (Kubernetes)
kubectl scale deployment/device-bridge --replicas=5 -n device-bridge

# Rollback (Kubernetes)
kubectl rollout undo deployment/device-bridge -n device-bridge
```

---

**Document Version:** 2.0
**Last Updated:** 2025-11-11
**Next Review:** 2026-02-11