# Browser Payment Client Examples

Complete browser-based payment terminal integration with Arabic/English support.

## Files

- **payment-client.js** - JavaScript client library for Device Bridge API
- **index.html** - Complete demo application with bilingual UI (Arabic/English)
- **README.md** - This file

## Features

✅ **Bilingual Support**
- Arabic (RTL) and English (LTR) interface
- Language switcher
- Currency names in both languages

✅ **Secure Communication**
- HTTPS/TLS support
- API key authentication
- Request ID tracking
- Timeout handling

✅ **Complete Payment Operations**
- Sale transactions
- Void transactions
- Refund transactions
- Settlement operations
- Device status checking

✅ **User Experience**
- Responsive design
- Loading indicators
- Error handling
- Receipt display

## Quick Start

### 1. Configure the Client

Edit `index.html` and update the configuration:

```javascript
const config = {
    baseURL: 'https://localhost:8080',  // Your Device Bridge URL
    apiKey: 'your-api-key-here',        // Your API key
    deviceID: 'payment-001',            // Your payment device ID
    language: 'en',                     // Default language ('en' or 'ar')
};
```

### 2. Serve the Files

You can use any web server. Examples:

**Python:**
```bash
# Python 3
python3 -m http.server 8000

# Open browser to http://localhost:8000
```

**Node.js:**
```bash
npx http-server -p 8000

# Open browser to http://localhost:8000
```

**VS Code:**
Install "Live Server" extension and click "Go Live"

### 3. Open in Browser

Navigate to:
```
http://localhost:8000/index.html
```

## Using the JavaScript Client

### Basic Usage

```javascript
// Initialize client
const client = new PaymentClient({
    baseURL: 'https://localhost:8080',
    apiKey: 'your-api-key',
    deviceID: 'payment-001',
    language: 'en',  // or 'ar'
});

// Process a sale
const response = await client.sale(
    10000,  // 100.00 in minor units
    'SAR',  // Currency
    'INV-001'  // Optional reference
);

console.log('Auth Code:', response.auth_code);
console.log('Transaction ID:', response.transaction_id);
```

### Advanced Features

```javascript
// Void a transaction
await client.void(
    'original-transaction-id',
    10000,
    'SAR'
);

// Refund a transaction
await client.refund(
    'original-transaction-id',
    5000,  // Partial refund
    'SAR'
);

// End-of-day settlement
await client.settlement();

// Get device status
const status = await client.getStatus();
console.log('Connected:', status.connected);

// Format amount for display
const formatted = client.formatAmount(10000, 'SAR');
console.log(formatted);  // "SAR 100.00" or "100.00 ريال سعودي"

// Parse user input
const minorUnits = client.parseAmount('100.50', 'SAR');
console.log(minorUnits);  // 10050
```

## Currency Support

The client supports all Middle East currencies with correct decimal handling:

| Currency | Name (EN) | Name (AR) | Decimals |
|----------|-----------|-----------|----------|
| SAR | Saudi Riyal | ريال سعودي | 2 |
| KWD | Kuwaiti Dinar | دينار كويتي | 3 |
| AED | UAE Dirham | درهم إماراتي | 2 |
| BHD | Bahraini Dinar | دينار بحريني | 3 |
| OMR | Omani Rial | ريال عماني | 3 |
| QAR | Qatari Riyal | ريال قطري | 2 |
| JOD | Jordanian Dinar | دينار أردني | 3 |
| EGP | Egyptian Pound | جنيه مصري | 2 |

## Arabic/RTL Support

The demo application includes full Arabic support:

### Language Switching

```javascript
// Switch to Arabic
client.setLanguage('ar');

// Switch to English
client.setLanguage('en');
```

### RTL Layout

The HTML automatically switches between LTR and RTL:

```html
<html lang="ar" dir="rtl">  <!-- Arabic -->
<html lang="en" dir="ltr">  <!-- English -->
```

### Arabic Currency Formatting

```javascript
// Get currency info
const info = client.getCurrencyInfo('SAR');
console.log(info);
// {
//   nameEn: 'Saudi Riyal',
//   nameAr: 'ريال سعودي',
//   symbol: 'SAR',
//   decimals: 2
// }
```

## Error Handling

```javascript
try {
    const response = await client.sale(10000, 'SAR');
    console.log('Success:', response);
} catch (error) {
    if (error.message === 'Request timeout') {
        console.error('Transaction timed out');
    } else if (error.message.includes('HTTP 401')) {
        console.error('Invalid API key');
    } else {
        console.error('Transaction failed:', error.message);
    }
}
```

## Security Considerations

### CORS Configuration

Ensure your Device Bridge CORS settings allow your domain:

```yaml
# config.production.yaml
cors:
  allowed_origins:
    - "https://yourwebsite.com"
  allowed_methods:
    - "GET"
    - "POST"
  allow_credentials: true
```

### Content Security Policy

The demo includes a strict CSP. Adjust as needed:

```html
<meta http-equiv="Content-Security-Policy"
      content="default-src 'self';
               connect-src 'self' https://localhost:8080;
               script-src 'self' 'unsafe-inline';
               style-src 'self' 'unsafe-inline';">
```

### HTTPS Only

Always use HTTPS in production:

```javascript
const client = new PaymentClient({
    baseURL: 'https://your-device-bridge.com',  // HTTPS, not HTTP
    // ...
});
```

## Browser Compatibility

Tested and supported:

- ✅ Chrome 90+
- ✅ Firefox 88+
- ✅ Safari 14+
- ✅ Edge 90+

Requires:
- ES6+ support
- Fetch API
- Promises/async-await

## Debugging

Enable detailed logging in browser console:

```javascript
// Check connection
const status = await client.getStatus();
console.log('Device status:', status);

// Monitor network requests
// Open DevTools > Network tab and filter by "device"

// Check CORS headers
// Network tab > select request > Headers tab
// Look for "Access-Control-Allow-Origin"
```

## Production Checklist

Before deploying to production:

- [ ] Replace API key with production key
- [ ] Update `baseURL` to production URL
- [ ] Configure CORS for production domain
- [ ] Enable HTTPS/TLS
- [ ] Test with actual payment terminal
- [ ] Review security headers
- [ ] Implement error logging
- [ ] Add transaction history
- [ ] Test both English and Arabic
- [ ] Test on all target browsers
- [ ] Load test your integration

## Support

For issues or questions:

- **Documentation:** [BROWSER_CORS_SECURITY.md](../../docs/BROWSER_CORS_SECURITY.md)
- **API Reference:** [PAYMENT_TERMINAL_SETUP.md](../../docs/PAYMENT_TERMINAL_SETUP.md)
- **Arabic Guide:** [PAYMENT_TERMINAL_SETUP_AR.md](../../docs/PAYMENT_TERMINAL_SETUP_AR.md)

## License

MIT License - See [LICENSE](../../LICENSE)

---

**Version:** 2.0
**Last Updated:** 2025-11-11
**Status:** ✅ Production Ready
