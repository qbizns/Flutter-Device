/**
 * Device Bridge Browser Client - Payment Terminal
 * Version: 2.0
 *
 * Secure JavaScript client for browser-based payment terminal integration.
 * Supports both English and Arabic (RTL) interfaces.
 */

class PaymentClient {
    /**
     * Initialize the payment client
     * @param {Object} config - Configuration options
     * @param {string} config.baseURL - Device Bridge API URL
     * @param {string} config.apiKey - API authentication key
     * @param {string} config.deviceID - Payment device ID
     * @param {string} config.language - UI language ('en' or 'ar')
     */
    constructor(config) {
        this.baseURL = config.baseURL || 'https://localhost:8080';
        this.apiKey = config.apiKey;
        this.deviceID = config.deviceID;
        this.language = config.language || 'en';
        this.requestTimeout = config.timeout || 30000;  // 30 seconds

        // Validate configuration
        if (!this.apiKey) {
            throw new Error('API key is required');
        }
        if (!this.deviceID) {
            throw new Error('Device ID is required');
        }
    }

    /**
     * Make an authenticated API request
     * @private
     */
    async _request(method, endpoint, data = null) {
        const url = `${this.baseURL}${endpoint}`;

        const options = {
            method: method,
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': this.apiKey,
                'X-Device-ID': this.deviceID,
                'X-Request-ID': this._generateRequestID(),
                'Accept-Language': this.language,
            },
            credentials: 'include',
        };

        if (data) {
            options.body = JSON.stringify(data);
        }

        // Add timeout
        const controller = new AbortController();
        const timeoutId = setTimeout(() => controller.abort(), this.requestTimeout);
        options.signal = controller.signal;

        try {
            const response = await fetch(url, options);
            clearTimeout(timeoutId);

            if (!response.ok) {
                const error = await response.json().catch(() => ({
                    error: `HTTP ${response.status}: ${response.statusText}`
                }));
                throw new Error(error.error || error.message);
            }

            return await response.json();
        } catch (error) {
            clearTimeout(timeoutId);

            if (error.name === 'AbortError') {
                throw new Error('Request timeout');
            }

            throw error;
        }
    }

    /**
     * Process a payment transaction
     * @param {Object} params - Transaction parameters
     * @param {number} params.amount - Amount in minor currency units (e.g., 10000 = 100.00)
     * @param {string} params.currency - Currency code (SAR, KWD, AED, etc.)
     * @param {string} params.type - Transaction type (sale, void, refund, preauth, completion)
     * @param {string} [params.reference] - Optional reference number
     * @param {string} [params.originalReference] - Required for void/refund
     * @returns {Promise<Object>} Transaction response
     */
    async processTransaction(params) {
        const { amount, currency, type = 'sale', reference, originalReference } = params;

        // Validate required parameters
        if (!amount || amount <= 0) {
            throw new Error('Invalid amount');
        }
        if (!currency) {
            throw new Error('Currency is required');
        }

        const payload = {
            type: type,
            amount: amount,
            currency: currency,
            timestamp: new Date().toISOString(),
        };

        if (reference) {
            payload.reference = reference;
        }

        if (originalReference) {
            payload.original_reference = originalReference;
        }

        return await this._request(
            'POST',
            `/api/devices/${this.deviceID}/transaction`,
            payload
        );
    }

    /**
     * Process a sale transaction
     * @param {number} amount - Amount in minor currency units
     * @param {string} currency - Currency code
     * @param {string} [reference] - Optional reference
     * @returns {Promise<Object>} Transaction response
     */
    async sale(amount, currency, reference) {
        return await this.processTransaction({
            amount,
            currency,
            type: 'sale',
            reference,
        });
    }

    /**
     * Void a previous transaction
     * @param {string} originalReference - Original transaction ID to void
     * @param {number} amount - Original amount
     * @param {string} currency - Original currency
     * @returns {Promise<Object>} Transaction response
     */
    async void(originalReference, amount, currency) {
        return await this.processTransaction({
            amount,
            currency,
            type: 'void',
            originalReference,
        });
    }

    /**
     * Refund a previous transaction
     * @param {string} originalReference - Original transaction ID to refund
     * @param {number} amount - Refund amount
     * @param {string} currency - Currency code
     * @returns {Promise<Object>} Transaction response
     */
    async refund(originalReference, amount, currency) {
        return await this.processTransaction({
            amount,
            currency,
            type: 'refund',
            originalReference,
        });
    }

    /**
     * Perform end-of-day settlement
     * @param {string} [batchNumber] - Optional batch number
     * @returns {Promise<Object>} Settlement response
     */
    async settlement(batchNumber) {
        const payload = {};
        if (batchNumber) {
            payload.batch_number = batchNumber;
        }

        return await this._request(
            'POST',
            `/api/devices/${this.deviceID}/settlement`,
            payload
        );
    }

    /**
     * Get device status
     * @returns {Promise<Object>} Device status
     */
    async getStatus() {
        return await this._request('GET', `/api/devices/${this.deviceID}/status`);
    }

    /**
     * List all devices
     * @returns {Promise<Array>} List of devices
     */
    async listDevices() {
        const response = await this._request('GET', '/api/devices');
        return response.devices || [];
    }

    /**
     * Format amount for display
     * @param {number} amount - Amount in minor units
     * @param {string} currency - Currency code
     * @returns {string} Formatted amount
     */
    formatAmount(amount, currency) {
        // Different currencies have different decimal places
        const decimalPlaces = ['KWD', 'BHD', 'OMR', 'JOD'].includes(currency) ? 3 : 2;
        const divisor = Math.pow(10, decimalPlaces);
        const value = amount / divisor;

        return new Intl.NumberFormat(this.language === 'ar' ? 'ar-SA' : 'en-US', {
            style: 'currency',
            currency: currency,
            minimumFractionDigits: decimalPlaces,
            maximumFractionDigits: decimalPlaces,
        }).format(value);
    }

    /**
     * Parse user input amount to minor units
     * @param {string|number} input - User input (e.g., "100.50")
     * @param {string} currency - Currency code
     * @returns {number} Amount in minor units
     */
    parseAmount(input, currency) {
        const value = parseFloat(input);
        if (isNaN(value) || value < 0) {
            throw new Error('Invalid amount');
        }

        const decimalPlaces = ['KWD', 'BHD', 'OMR', 'JOD'].includes(currency) ? 3 : 2;
        const multiplier = Math.pow(10, decimalPlaces);

        return Math.round(value * multiplier);
    }

    /**
     * Get currency info
     * @param {string} currency - Currency code
     * @returns {Object} Currency information
     */
    getCurrencyInfo(currency) {
        const currencies = {
            'SAR': { nameEn: 'Saudi Riyal', nameAr: 'ريال سعودي', symbol: 'SAR', decimals: 2 },
            'KWD': { nameEn: 'Kuwaiti Dinar', nameAr: 'دينار كويتي', symbol: 'KWD', decimals: 3 },
            'AED': { nameEn: 'UAE Dirham', nameAr: 'درهم إماراتي', symbol: 'AED', decimals: 2 },
            'BHD': { nameEn: 'Bahraini Dinar', nameAr: 'دينار بحريني', symbol: 'BHD', decimals: 3 },
            'OMR': { nameEn: 'Omani Rial', nameAr: 'ريال عماني', symbol: 'OMR', decimals: 3 },
            'QAR': { nameEn: 'Qatari Riyal', nameAr: 'ريال قطري', symbol: 'QAR', decimals: 2 },
            'JOD': { nameEn: 'Jordanian Dinar', nameAr: 'دينار أردني', symbol: 'JOD', decimals: 3 },
            'EGP': { nameEn: 'Egyptian Pound', nameAr: 'جنيه مصري', symbol: 'EGP', decimals: 2 },
        };

        return currencies[currency] || { nameEn: currency, nameAr: currency, symbol: currency, decimals: 2 };
    }

    /**
     * Generate unique request ID
     * @private
     */
    _generateRequestID() {
        const timestamp = Date.now();
        const random = Math.random().toString(36).substring(2, 11);
        return `req_${timestamp}_${random}`;
    }

    /**
     * Set language
     * @param {string} lang - Language code ('en' or 'ar')
     */
    setLanguage(lang) {
        if (lang !== 'en' && lang !== 'ar') {
            throw new Error('Unsupported language. Use "en" or "ar"');
        }
        this.language = lang;
    }
}

// Export for use in modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = PaymentClient;
}

// Export for browser global
if (typeof window !== 'undefined') {
    window.PaymentClient = PaymentClient;
}
