# Browser Integration & CORS Security Guide

**Version:** 1.0
**Date:** 2025-11-11
**Status:** ✅ Production Ready

---

## Table of Contents

1. [Overview](#overview)
2. [CORS Security Model](#cors-security-model)
3. [Security Best Practices](#security-best-practices)
4. [Browser Client Implementation](#browser-client-implementation)
5. [WebSocket Security](#websocket-security)
6. [Common Security Pitfalls](#common-security-pitfalls)
7. [Testing & Validation](#testing--validation)
8. [Production Configuration](#production-configuration)

---

## Overview

The Device Bridge v2 supports browser-based integrations through its REST API and WebSocket connections. This document provides comprehensive security guidance for CORS (Cross-Origin Resource Sharing) configuration and browser client implementation.

### Security Principles

1. **Least Privilege:** Only allow necessary origins and methods
2. **Defense in Depth:** Multiple layers of security
3. **Zero Trust:** Validate all requests, even from allowed origins
4. **Audit Everything:** Log all cross-origin requests

---

## CORS Security Model

### What is CORS?

CORS is a browser security feature that controls how web pages from one domain can access resources from another domain.

### CORS Headers Reference

| Header | Purpose | Security Impact |
|--------|---------|-----------------|
| `Access-Control-Allow-Origin` | Specifies allowed origins | **CRITICAL** - Controls who can access |
| `Access-Control-Allow-Methods` | Specifies allowed HTTP methods | **HIGH** - Limits operations |
| `Access-Control-Allow-Headers` | Specifies allowed request headers | **MEDIUM** - Controls metadata |
| `Access-Control-Allow-Credentials` | Allow cookies/auth headers | **CRITICAL** - Authentication bypass risk |
| `Access-Control-Max-Age` | Preflight cache duration | **LOW** - Performance vs security tradeoff |
| `Access-Control-Expose-Headers` | Headers visible to browser | **LOW** - Information disclosure |

---

## Security Best Practices

### 1. Origin Whitelisting

**❌ NEVER do this in production:**
```yaml
cors:
  allowed_origins:
    - "*"  # Allows ANY website to access your API
```

**✅ ALWAYS whitelist specific origins:**
```yaml
cors:
  allowed_origins:
    - "https://pos.example.com"
    - "https://admin.example.com"
    - "https://app.example.com"
```

**Development-only wildcard:**
```yaml
# config.development.yaml
cors:
  allowed_origins:
    - "http://localhost:3000"
    - "http://localhost:8080"
    - "http://127.0.0.1:*"  # Allow any port on localhost

# config.production.yaml
cors:
  allowed_origins:
    - "https://pos.example.com"  # Exact match only
```

### 2. Method Restriction

Only allow necessary HTTP methods:

```yaml
cors:
  allowed_methods:
    - "GET"
    - "POST"
    # DO NOT allow:
    # - "PUT"
    # - "DELETE"
    # - "PATCH"
  unless specifically needed
```

### 3. Credentials Handling

**⚠️ WARNING:** Enabling credentials with wildcard origin is a critical vulnerability:

```yaml
# ❌ NEVER do this
cors:
  allowed_origins:
    - "*"
  allow_credentials: true  # CRITICAL VULNERABILITY

# ✅ Correct approach
cors:
  allowed_origins:
    - "https://pos.example.com"
  allow_credentials: true  # Safe with specific origin
```

### 4. Header Whitelisting

Only allow necessary headers:

```yaml
cors:
  allowed_headers:
    - "Content-Type"
    - "Authorization"
    - "X-Device-ID"
    - "X-Request-ID"
  # Avoid allowing all headers:
  # - "*"
```

### 5. Preflight Caching

Balance security and performance:

```yaml
cors:
  max_age: 3600  # 1 hour (good balance)
  # Too short: Performance impact
  # Too long: Delayed security updates
```

---

## Browser Client Implementation

### Secure JavaScript Client

```javascript
// browser-client.js
class DeviceBridgeClient {
    constructor(config) {
        this.baseURL = config.baseURL;  // e.g., 'https://localhost:8080'
        this.apiKey = config.apiKey;
        this.deviceID = config.deviceID;
    }

    // Generic API request with security headers
    async request(method, endpoint, data = null) {
        const options = {
            method: method,
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': this.apiKey,
                'X-Device-ID': this.deviceID,
                'X-Request-ID': this.generateRequestID(),
            },
            credentials: 'include',  // Send cookies if needed
        };

        if (data) {
            options.body = JSON.stringify(data);
        }

        try {
            const response = await fetch(`${this.baseURL}${endpoint}`, options);

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            return await response.json();
        } catch (error) {
            console.error('API Request failed:', error);
            throw error;
        }
    }

    // Payment transaction
    async processPayment(amount, currency, type = 'sale') {
        return await this.request('POST', '/api/devices/payment-001/transaction', {
            type: type,
            amount: amount,
            currency: currency,
            timestamp: new Date().toISOString(),
        });
    }

    // Get device status
    async getDeviceStatus(deviceID) {
        return await this.request('GET', `/api/devices/${deviceID}/status`);
    }

    // List devices
    async listDevices() {
        return await this.request('GET', '/api/devices');
    }

    // Generate unique request ID
    generateRequestID() {
        return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    }
}

// Usage
const client = new DeviceBridgeClient({
    baseURL: 'https://localhost:8080',
    apiKey: 'your-api-key-here',
    deviceID: 'payment-001',
});

// Process a payment
client.processPayment(10000, 'SAR', 'sale')
    .then(response => {
        console.log('Payment successful:', response);
    })
    .catch(error => {
        console.error('Payment failed:', error);
    });
```

### HTML Integration Example

```html
<!DOCTYPE html>
<html lang="en" dir="ltr">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Device Bridge - Payment Terminal</title>

    <!-- Security Headers -->
    <meta http-equiv="Content-Security-Policy"
          content="default-src 'self';
                   connect-src 'self' https://localhost:8080 wss://localhost:8080;
                   script-src 'self' 'unsafe-inline';">
</head>
<body>
    <div id="app">
        <h1>Payment Terminal</h1>

        <form id="paymentForm">
            <label>Amount (SAR):</label>
            <input type="number" id="amount" min="1" step="0.01" required>

            <label>Currency:</label>
            <select id="currency">
                <option value="SAR">SAR - ريال سعودي</option>
                <option value="KWD">KWD - دينار كويتي</option>
                <option value="AED">AED - درهم إماراتي</option>
            </select>

            <button type="submit">Process Payment</button>
        </form>

        <div id="result"></div>
    </div>

    <script src="browser-client.js"></script>
    <script>
        // Initialize client
        const client = new DeviceBridgeClient({
            baseURL: 'https://localhost:8080',
            apiKey: 'dev-api-key-12345',
            deviceID: 'payment-001',
        });

        // Handle form submission
        document.getElementById('paymentForm').addEventListener('submit', async (e) => {
            e.preventDefault();

            const amount = parseFloat(document.getElementById('amount').value) * 100;
            const currency = document.getElementById('currency').value;

            try {
                const response = await client.processPayment(amount, currency, 'sale');

                document.getElementById('result').innerHTML = `
                    <div class="success">
                        ✅ Payment Approved<br>
                        Auth Code: ${response.auth_code}<br>
                        Transaction ID: ${response.transaction_id}
                    </div>
                `;
            } catch (error) {
                document.getElementById('result').innerHTML = `
                    <div class="error">
                        ❌ Payment Failed: ${error.message}
                    </div>
                `;
            }
        });
    </script>
</body>
</html>
```

---

## WebSocket Security

### Secure WebSocket Connection

```javascript
class SecureWebSocketClient {
    constructor(config) {
        this.url = config.url;  // wss://localhost:8080/ws
        this.apiKey = config.apiKey;
        this.deviceID = config.deviceID;
        this.ws = null;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = 5;
    }

    connect() {
        // Use secure WebSocket (wss://)
        this.ws = new WebSocket(`${this.url}?api_key=${this.apiKey}&device_id=${this.deviceID}`);

        this.ws.onopen = () => {
            console.log('WebSocket connected');
            this.reconnectAttempts = 0;
        };

        this.ws.onmessage = (event) => {
            const message = JSON.parse(event.data);
            this.handleMessage(message);
        };

        this.ws.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        this.ws.onclose = () => {
            console.log('WebSocket closed');
            this.reconnect();
        };
    }

    handleMessage(message) {
        switch (message.type) {
            case 'device_event':
                console.log('Device event:', message.data);
                break;
            case 'transaction_update':
                console.log('Transaction update:', message.data);
                break;
            default:
                console.warn('Unknown message type:', message.type);
        }
    }

    reconnect() {
        if (this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnectAttempts++;
            const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000);
            console.log(`Reconnecting in ${delay}ms...`);
            setTimeout(() => this.connect(), delay);
        } else {
            console.error('Max reconnect attempts reached');
        }
    }

    send(message) {
        if (this.ws && this.ws.readyState === WebSocket.OPEN) {
            this.ws.send(JSON.stringify(message));
        } else {
            console.error('WebSocket not connected');
        }
    }

    close() {
        if (this.ws) {
            this.ws.close();
        }
    }
}

// Usage
const wsClient = new SecureWebSocketClient({
    url: 'wss://localhost:8080/ws',
    apiKey: 'your-api-key',
    deviceID: 'payment-001',
});

wsClient.connect();
```

---

## Common Security Pitfalls

### 1. Wildcard Origin in Production

**❌ Vulnerability:**
```go
// DO NOT USE IN PRODUCTION
corsMiddleware := cors.New(cors.Options{
    AllowedOrigins:   []string{"*"},
    AllowCredentials: true,
})
```

**Impact:** Any website can access your API and steal user credentials.

**✅ Fix:**
```go
corsMiddleware := cors.New(cors.Options{
    AllowedOrigins: []string{
        "https://pos.example.com",
        "https://admin.example.com",
    },
    AllowCredentials: true,
})
```

### 2. Missing Origin Validation

**❌ Vulnerability:**
```go
// Accepting any Origin header value
w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
```

**Impact:** Attacker can bypass CORS by setting custom Origin.

**✅ Fix:**
```go
allowedOrigins := map[string]bool{
    "https://pos.example.com":   true,
    "https://admin.example.com": true,
}

origin := r.Header.Get("Origin")
if allowedOrigins[origin] {
    w.Header().Set("Access-Control-Allow-Origin", origin)
}
```

### 3. Overly Permissive Methods

**❌ Vulnerability:**
```yaml
cors:
  allowed_methods:
    - "*"  # Allows any HTTP method
```

**✅ Fix:**
```yaml
cors:
  allowed_methods:
    - "GET"
    - "POST"
    # Only methods actually needed
```

### 4. Information Disclosure via CORS

**❌ Vulnerability:**
```yaml
cors:
  expose_headers:
    - "X-Internal-Server-ID"
    - "X-Database-Query-Time"
    - "X-Admin-Email"
```

**Impact:** Leaks internal implementation details.

**✅ Fix:**
```yaml
cors:
  expose_headers:
    - "X-Request-ID"  # Safe to expose
    - "X-RateLimit-Remaining"
    # Only expose necessary headers
```

### 5. Missing Rate Limiting

**❌ Vulnerability:**
No rate limiting on CORS endpoints allows abuse.

**✅ Fix:**
```yaml
security:
  rate_limiting:
    enabled: true
    requests_per_second: 10
    burst: 20
    per_origin: true  # Rate limit per origin
```

---

## Testing & Validation

### Manual CORS Testing

```bash
# Test preflight request
curl -X OPTIONS https://localhost:8080/api/devices \
  -H "Origin: https://pos.example.com" \
  -H "Access-Control-Request-Method: POST" \
  -v

# Expected response headers:
# Access-Control-Allow-Origin: https://pos.example.com
# Access-Control-Allow-Methods: GET, POST
# Access-Control-Max-Age: 3600

# Test actual request
curl -X POST https://localhost:8080/api/devices/payment-001/transaction \
  -H "Origin: https://pos.example.com" \
  -H "Content-Type: application/json" \
  -d '{"type":"sale","amount":10000,"currency":"SAR"}' \
  -v

# Test unauthorized origin (should be rejected)
curl -X POST https://localhost:8080/api/devices/payment-001/transaction \
  -H "Origin: https://evil.com" \
  -H "Content-Type: application/json" \
  -d '{"type":"sale","amount":10000,"currency":"SAR"}' \
  -v
```

### Automated Security Tests

```javascript
// cors-security-test.js
const assert = require('assert');

async function testCORSSecurity() {
    const tests = [
        {
            name: 'Allowed origin should work',
            origin: 'https://pos.example.com',
            expectSuccess: true,
        },
        {
            name: 'Disallowed origin should fail',
            origin: 'https://evil.com',
            expectSuccess: false,
        },
        {
            name: 'Wildcard origin should fail',
            origin: '*',
            expectSuccess: false,
        },
    ];

    for (const test of tests) {
        const response = await fetch('https://localhost:8080/api/devices', {
            method: 'GET',
            headers: {
                'Origin': test.origin,
            },
        });

        const hasAllowOrigin = response.headers.has('access-control-allow-origin');

        if (test.expectSuccess) {
            assert(hasAllowOrigin, `${test.name}: Expected CORS headers`);
        } else {
            assert(!hasAllowOrigin, `${test.name}: Should not have CORS headers`);
        }

        console.log(`✅ ${test.name}`);
    }
}

testCORSSecurity().catch(console.error);
```

---

## Production Configuration

### Recommended Production CORS Config

```yaml
# config.production.yaml
cors:
  # Strict origin whitelisting
  allowed_origins:
    - "https://pos.example.com"
    - "https://admin.example.com"

  # Only necessary methods
  allowed_methods:
    - "GET"
    - "POST"

  # Only required headers
  allowed_headers:
    - "Content-Type"
    - "Authorization"
    - "X-Device-ID"
    - "X-Request-ID"

  # Enable credentials (with specific origins only)
  allow_credentials: true

  # Reasonable preflight cache
  max_age: 3600  # 1 hour

  # Only expose safe headers
  expose_headers:
    - "X-Request-ID"
    - "X-RateLimit-Remaining"

# Additional security
security:
  # Rate limiting per origin
  rate_limiting:
    enabled: true
    per_origin: true
    requests_per_second: 10
    burst: 20

  # HTTPS only
  force_https: true

  # Security headers
  headers:
    x_frame_options: "DENY"
    x_content_type_options: "nosniff"
    x_xss_protection: "1; mode=block"
    strict_transport_security: "max-age=31536000; includeSubDomains"
```

### Content Security Policy

```html
<!-- Add to your web application -->
<meta http-equiv="Content-Security-Policy"
      content="default-src 'self';
               connect-src 'self' https://localhost:8080 wss://localhost:8080;
               script-src 'self';
               style-src 'self' 'unsafe-inline';">
```

---

## Security Checklist

Before deploying to production:

- [ ] Wildcard (`*`) origin removed from all environments
- [ ] Specific origins whitelisted (HTTPS only)
- [ ] Only necessary HTTP methods allowed
- [ ] Only required headers allowed
- [ ] Credentials only enabled with specific origins
- [ ] Rate limiting enabled per origin
- [ ] Security headers configured
- [ ] HTTPS enforced (no HTTP in production)
- [ ] WebSocket connections use WSS (not WS)
- [ ] Content Security Policy configured
- [ ] CORS configuration tested
- [ ] Automated security tests in CI/CD
- [ ] Monitoring and alerting configured
- [ ] Incident response plan documented

---

## References

- [MDN CORS Documentation](https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS)
- [OWASP CORS Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross-Origin_Resource_Sharing_Cheat_Sheet.html)
- [Device Bridge Security Guide](PAYMENT_SECURITY_GUIDE.md)
- [Device Bridge Monitoring Guide](PAYMENT_MONITORING_GUIDE.md)

---

**Document Version:** 1.0
**Last Updated:** 2025-11-11
**Security Rating:** Production Ready ✅
