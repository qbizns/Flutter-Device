# Payment Terminal Driver - Security Best Practices Guide

**Version:** 1.0
**Date:** November 11, 2025
**Audience:** Developers, DevOps Engineers, Security Teams

---

## Table of Contents

- [1. Introduction](#1-introduction)
- [2. Deployment Security](#2-deployment-security)
- [3. Configuration Security](#3-configuration-security)
- [4. Network Security](#4-network-security)
- [5. Data Protection](#5-data-protection)
- [6. Authentication & Authorization](#6-authentication--authorization)
- [7. Monitoring & Incident Response](#7-monitoring--incident-response)
- [8. Compliance](#8-compliance)
- [9. Security Checklists](#9-security-checklists)

---

## 1. Introduction

This guide provides security best practices for deploying and operating the Device Bridge v2 payment terminal driver. Following these guidelines ensures PCI DSS compliance and protects sensitive payment card data.

### Security Principles

1. **Defense in Depth:** Multiple layers of security controls
2. **Least Privilege:** Minimal necessary access rights
3. **Fail Secure:** Safe defaults, secure error handling
4. **Audit Everything:** Comprehensive logging and monitoring
5. **Encrypt Always:** End-to-end encryption for sensitive data

---

## 2. Deployment Security

### 2.1 Production Environment

**✅ DO:**
- Deploy in a PCI DSS compliant environment
- Use dedicated servers/containers for payment processing
- Implement network segmentation (payment network isolated)
- Enable firewall rules (allow only necessary ports)
- Keep systems patched and up-to-date
- Use configuration management (Infrastructure as Code)

**❌ DON'T:**
- Run on shared hosting or multi-tenant environments
- Deploy on developer workstations
- Use default passwords or credentials
- Expose payment systems directly to the internet
- Mix payment and non-payment workloads

### 2.2 Container Security

```yaml
# Secure Docker deployment example
version: '3.8'
services:
  device-bridge:
    image: device-bridge:latest
    security_opt:
      - no-new-privileges:true
    read_only: true
    cap_drop:
      - ALL
    cap_add:
      - NET_BIND_SERVICE
    tmpfs:
      - /tmp
    networks:
      - payment-network
    secrets:
      - terminal_cert
      - terminal_key

networks:
  payment-network:
    driver: bridge
    internal: true

secrets:
  terminal_cert:
    external: true
  terminal_key:
    external: true
```

### 2.3 System Hardening

```bash
# Linux system hardening
# 1. Disable unnecessary services
sudo systemctl disable <unnecessary-service>

# 2. Configure firewall
sudo ufw default deny incoming
sudo ufw default allow outgoing
sudo ufw allow from <payment-network> to any port 50051 proto tcp
sudo ufw enable

# 3. Set file permissions
sudo chmod 600 /etc/device-bridge/config.yaml
sudo chmod 600 /etc/device-bridge/certs/*.key
sudo chmod 644 /etc/device-bridge/certs/*.crt

# 4. Create dedicated user
sudo useradd -r -s /usr/sbin/nologin devicebridge
sudo chown -R devicebridge:devicebridge /var/lib/device-bridge
```

---

## 3. Configuration Security

### 3.1 Secure Configuration

**Production config.yaml:**
```yaml
server:
  grpc:
    address: "0.0.0.0:50051"
    tls:
      enabled: true
      cert_file: "/etc/device-bridge/certs/server.crt"
      key_file: "/etc/device-bridge/certs/server.key"
      client_auth: "require"  # Mutual TLS

security:
  audit_log:
    enabled: true
    path: "/var/log/device-bridge/audit.log"
    rotation: "daily"
    retention_days: 365  # PCI DSS requires 1 year

  rate_limiting:
    enabled: true
    requests_per_second: 10
    burst_size: 20

devices:
  - id: terminal-001
    name: "Payment Terminal 1"
    kind: payment.tcp
    enabled: true
    metadata:
      host: "terminal001.internal.company.com"
      port: "3000"
      terminal_id: "${TERMINAL_ID}"  # From environment/secrets
      merchant_id: "${MERCHANT_ID}"  # From environment/secrets
      provider: "mada"
      use_tls: "true"
      tls_verify: "true"
      timeout: "30s"
      max_retries: 3
```

### 3.2 Secrets Management

**✅ DO:**
- Store secrets in environment variables or secret management systems
- Use HashiCorp Vault, AWS Secrets Manager, or Kubernetes Secrets
- Rotate secrets regularly (at least every 90 days)
- Never commit secrets to version control
- Use `.gitignore` for sensitive files

**Environment variables:**
```bash
# /etc/device-bridge/env
export TERMINAL_ID="<terminal-id>"
export MERCHANT_ID="<merchant-id>"
export TLS_CERT_PATH="/etc/device-bridge/certs/terminal.crt"
export TLS_KEY_PATH="/etc/device-bridge/certs/terminal.key"
export DB_PASSWORD="<strong-password>"
```

**Using Kubernetes Secrets:**
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: payment-terminal-secrets
type: Opaque
stringData:
  terminal-id: "TERM001"
  merchant-id: "MERCHANT001"
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: device-bridge
spec:
  template:
    spec:
      containers:
      - name: device-bridge
        env:
        - name: TERMINAL_ID
          valueFrom:
            secretKeyRef:
              name: payment-terminal-secrets
              key: terminal-id
        - name: MERCHANT_ID
          valueFrom:
            secretKeyRef:
              name: payment-terminal-secrets
              key: merchant-id
```

---

## 4. Network Security

### 4.1 TLS Configuration

**Strong TLS Configuration:**
```go
// Secure TLS config
tlsConfig := &tls.Config{
    MinVersion: tls.VersionTLS13, // TLS 1.3 only
    CipherSuites: []uint16{
        tls.TLS_AES_256_GCM_SHA384,
        tls.TLS_AES_128_GCM_SHA256,
        tls.TLS_CHACHA20_POLY1305_SHA256,
    },
    PreferServerCipherSuites: true,
    CurvePreferences: []tls.CurveID{
        tls.X25519,
        tls.CurveP256,
    },
    // Certificate pinning (optional but recommended)
    VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
        // Implement certificate pinning logic
        return nil
    },
}
```

### 4.2 Certificate Management

**Generate certificates:**
```bash
# 1. Generate CA
openssl genrsa -out ca.key 4096
openssl req -new -x509 -days 365 -key ca.key -out ca.crt

# 2. Generate terminal certificate
openssl genrsa -out terminal.key 4096
openssl req -new -key terminal.key -out terminal.csr
openssl x509 -req -days 365 -in terminal.csr -CA ca.crt -CAkey ca.key -set_serial 01 -out terminal.crt

# 3. Set permissions
chmod 600 terminal.key
chmod 644 terminal.crt
```

**Certificate rotation (every 90 days):**
```bash
#!/bin/bash
# cert-rotate.sh

OLD_CERT="/etc/device-bridge/certs/terminal.crt"
NEW_CERT="/etc/device-bridge/certs/terminal-new.crt"

# Generate new certificate
generate_new_cert

# Test new certificate
test_certificate "$NEW_CERT"

# Rotate
mv "$OLD_CERT" "$OLD_CERT.old"
mv "$NEW_CERT" "$OLD_CERT"

# Reload service
systemctl reload device-bridge

# Cleanup after 7 days
find /etc/device-bridge/certs -name "*.old" -mtime +7 -delete
```

### 4.3 Network Segmentation

```
┌─────────────────────────────────────────┐
│         Public Internet                 │
└──────────────┬──────────────────────────┘
               │
               ↓
    ┌──────────────────────┐
    │  Firewall/WAF        │  ← External firewall
    └──────────┬───────────┘
               │
               ↓
    ┌──────────────────────┐
    │  DMZ (Web Server)    │  ← Web tier
    └──────────┬───────────┘
               │
               ↓
    ┌──────────────────────┐
    │  Internal Firewall   │  ← Internal firewall
    └──────────┬───────────┘
               │
               ↓
    ┌──────────────────────┐
    │  Device Bridge       │  ← Application tier
    └──────────┬───────────┘
               │
               ↓
    ┌──────────────────────┐
    │  Payment Terminal    │  ← Payment network (isolated)
    └──────────────────────┘
```

---

## 5. Data Protection

### 5.1 Card Data Handling

**PCI DSS Requirements:**
- ✅ NEVER store full track data (magnetic stripe, chip data)
- ✅ NEVER store card verification code (CVV2, CVC2, CID)
- ✅ NEVER store PIN or PIN block
- ✅ Mask PAN (Primary Account Number) in all logs
- ✅ Only store last 4 digits for receipts
- ✅ Encrypt data in transit (TLS 1.2+)

**Code Example:**
```go
// ✅ CORRECT: Masked card number
response.CardNumberMasked = "****1234"

// ❌ WRONG: Full PAN
response.CardNumber = "4111111111111111" // NEVER DO THIS

// ✅ CORRECT: No CVV storage
// CVV is only transmitted, never stored

// ✅ CORRECT: No PIN storage
// PIN is encrypted by terminal, never logged
```

### 5.2 Audit Log Protection

**Audit log configuration:**
```yaml
audit_log:
  enabled: true
  path: "/var/log/device-bridge/audit.log"
  format: "json"
  rotation:
    max_size: 100MB
    max_age: 365  # Days
    max_backups: 12
    compress: true
  permissions: "0600"
  owner: "devicebridge"
  group: "devicebridge"
```

**Log rotation script:**
```bash
# /etc/logrotate.d/device-bridge
/var/log/device-bridge/audit.log {
    daily
    rotate 365
    compress
    delaycompress
    notifempty
    create 0600 devicebridge devicebridge
    postrotate
        systemctl reload device-bridge
    endscript
}
```

### 5.3 Data Encryption at Rest

**Encrypt audit logs:**
```bash
# Using LUKS for log volume
cryptsetup luksFormat /dev/sdb1
cryptsetup open /dev/sdb1 audit-logs
mkfs.ext4 /dev/mapper/audit-logs
mount /dev/mapper/audit-logs /var/log/device-bridge
```

---

## 6. Authentication & Authorization

### 6.1 Terminal Authentication

**Strong authentication example:**
```yaml
devices:
  - id: terminal-001
    auth:
      type: "certificate"  # Certificate-based auth
      cert_file: "/etc/device-bridge/certs/terminal-001.crt"
      key_file: "/etc/device-bridge/certs/terminal-001.key"
      ca_file: "/etc/device-bridge/certs/ca.crt"

    # Alternative: Token-based auth
    # auth:
    #   type: "token"
    #   token_file: "/run/secrets/terminal-token"
    #   expires: "2025-12-31T23:59:59Z"
```

### 6.2 API Authentication

**gRPC with mTLS:**
```go
// Server-side
creds, err := credentials.NewServerTLSFromFile(certFile, keyFile)
server := grpc.NewServer(grpc.Creds(creds))

// Client-side
creds, err := credentials.NewClientTLSFromFile(certFile, "")
conn, err := grpc.Dial(address, grpc.WithTransportCredentials(creds))
```

### 6.3 Access Control

**Role-based access:**
```yaml
users:
  - name: "operator"
    role: "operator"
    permissions:
      - "transaction:process"
      - "status:read"

  - name: "admin"
    role: "admin"
    permissions:
      - "transaction:*"
      - "config:*"
      - "audit:read"

  - name: "auditor"
    role: "auditor"
    permissions:
      - "audit:read"
      - "status:read"
```

---

## 7. Monitoring & Incident Response

### 7.1 Security Monitoring

**Key metrics to monitor:**
- Failed authentication attempts
- Unusual transaction patterns
- High error rates
- Connection failures
- Configuration changes
- Certificate expiration

**Prometheus metrics example:**
```go
var (
    transactionsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "payment_transactions_total",
            Help: "Total number of transactions",
        },
        []string{"device", "type", "status"},
    )

    failedAuthAttempts = prometheus.NewCounter(
        prometheus.CounterOpts{
            Name: "payment_auth_failures_total",
            Help: "Total number of failed auth attempts",
        },
    )
)
```

### 7.2 Alerting Rules

**Example Prometheus alerts:**
```yaml
groups:
  - name: payment_security
    rules:
      - alert: HighFailureRate
        expr: rate(payment_transactions_total{status="failed"}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High payment failure rate"

      - alert: FailedAuthentication
        expr: increase(payment_auth_failures_total[5m]) > 10
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Multiple failed authentication attempts"

      - alert: CertificateExpiringSoon
        expr: (payment_cert_expiry_timestamp - time()) < 604800
        labels:
          severity: warning
        annotations:
          summary: "Certificate expires in less than 7 days"
```

### 7.3 Incident Response

**Security incident checklist:**

1. **Detect**
   - [ ] Monitor alerts and logs
   - [ ] Automated anomaly detection
   - [ ] Regular security scans

2. **Respond**
   - [ ] Isolate affected systems
   - [ ] Preserve evidence (logs, memory dumps)
   - [ ] Notify security team
   - [ ] Document timeline

3. **Recover**
   - [ ] Patch vulnerabilities
   - [ ] Rotate credentials
   - [ ] Restore from backups if needed
   - [ ] Verify system integrity

4. **Review**
   - [ ] Post-incident analysis
   - [ ] Update procedures
   - [ ] Improve monitoring
   - [ ] Train team

---

## 8. Compliance

### 8.1 PCI DSS Compliance Checklist

**Requirements:**
- [x] Build and Maintain a Secure Network
  - [x] Firewall configuration
  - [x] No default passwords
- [x] Protect Cardholder Data
  - [x] Card data masked
  - [x] Encryption in transit
- [x] Maintain a Vulnerability Management Program
  - [x] Anti-virus (system level)
  - [x] Secure code practices
- [x] Implement Strong Access Control Measures
  - [x] Need-to-know access
  - [x] Unique IDs
  - [x] Physical access controls (environment)
- [x] Regularly Monitor and Test Networks
  - [x] Audit logging
  - [x] Security testing
- [x] Maintain an Information Security Policy
  - [x] Security policy documented
  - [x] Staff trained

### 8.2 Audit Preparation

**Documents to maintain:**
1. System architecture diagram
2. Data flow diagram
3. Network diagram
4. Access control matrix
5. Audit logs (365 days)
6. Penetration test reports
7. Vulnerability scan reports
8. Incident response logs
9. Training records
10. Policy documentation

---

## 9. Security Checklists

### 9.1 Pre-Deployment Checklist

- [ ] TLS 1.3 enabled
- [ ] Strong cipher suites configured
- [ ] Certificates generated and secured
- [ ] Secrets stored securely (not in code)
- [ ] Firewall rules configured
- [ ] Audit logging enabled
- [ ] Log rotation configured
- [ ] Monitoring and alerting set up
- [ ] Backup and recovery tested
- [ ] Security scan completed
- [ ] Penetration test passed
- [ ] Documentation complete
- [ ] Team trained

### 9.2 Daily Operations Checklist

- [ ] Review audit logs
- [ ] Check alert status
- [ ] Verify backup completion
- [ ] Monitor transaction metrics
- [ ] Check certificate expiration dates
- [ ] Review failed transactions
- [ ] Verify system health

### 9.3 Monthly Security Review

- [ ] Review access logs
- [ ] Update security patches
- [ ] Rotate credentials
- [ ] Test disaster recovery
- [ ] Review and update firewall rules
- [ ] Conduct security awareness training
- [ ] Review incident reports
- [ ] Update documentation

### 9.4 Quarterly Security Audit

- [ ] Penetration testing
- [ ] Vulnerability scanning
- [ ] Code security review
- [ ] Access control audit
- [ ] Configuration review
- [ ] Compliance assessment
- [ ] Update risk assessment
- [ ] Review and update policies

---

## Appendix A: Security Tools

### Static Analysis
```bash
# Install gosec
go install github.com/securego/gosec/v2/cmd/gosec@latest

# Run security scan
gosec ./...

# Check for known vulnerabilities
govulncheck ./...
```

### Dependency Scanning
```bash
# Audit Go modules
go list -json -m all | nancy sleuth
```

### TLS Testing
```bash
# Test TLS configuration
nmap --script ssl-enum-ciphers -p 50051 localhost

# Test with testssl.sh
./testssl.sh localhost:50051
```

---

## Appendix B: Secure Configuration Template

See `configs/config.payment.secure.yaml` for a complete secure configuration template.

---

## Appendix C: Emergency Contacts

- **Security Team:** security@company.com
- **On-Call:** +1-XXX-XXX-XXXX
- **PCI Assessor:** assessor@company.com
- **Incident Response:** ir@company.com

---

**Last Updated:** 2025-11-11
**Next Review:** 2025-12-11
**Document Owner:** Security Team
