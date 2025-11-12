package payment_tcp

import (
	"fmt"
	"sync"
)

// Language represents a supported language
type Language string

const (
	// LanguageEnglish represents English language
	LanguageEnglish Language = "en"

	// LanguageArabic represents Arabic language
	LanguageArabic Language = "ar"
)

// TranslationKey represents a translatable string key
type TranslationKey string

// Translation keys for common strings
const (
	// Transaction types
	KeyTransactionSale       TranslationKey = "transaction.sale"
	KeyTransactionVoid       TranslationKey = "transaction.void"
	KeyTransactionRefund     TranslationKey = "transaction.refund"
	KeyTransactionPreAuth    TranslationKey = "transaction.preauth"
	KeyTransactionCompletion TranslationKey = "transaction.completion"
	KeyTransactionSettlement TranslationKey = "transaction.settlement"

	// Transaction status
	KeyStatusApproved TranslationKey = "status.approved"
	KeyStatusDeclined TranslationKey = "status.declined"
	KeyStatusError    TranslationKey = "status.error"

	// Receipt fields
	KeyReceiptMerchantName   TranslationKey = "receipt.merchant_name"
	KeyReceiptMerchantID     TranslationKey = "receipt.merchant_id"
	KeyReceiptTerminalID     TranslationKey = "receipt.terminal_id"
	KeyReceiptDate           TranslationKey = "receipt.date"
	KeyReceiptTime           TranslationKey = "receipt.time"
	KeyReceiptCardNumber     TranslationKey = "receipt.card_number"
	KeyReceiptCardType       TranslationKey = "receipt.card_type"
	KeyReceiptTransactionID  TranslationKey = "receipt.transaction_id"
	KeyReceiptAuthCode       TranslationKey = "receipt.auth_code"
	KeyReceiptAmount         TranslationKey = "receipt.amount"
	KeyReceiptResponse       TranslationKey = "receipt.response"
	KeyReceiptSignature      TranslationKey = "receipt.signature"
	KeyReceiptCustomerCopy   TranslationKey = "receipt.customer_copy"
	KeyReceiptMerchantCopy   TranslationKey = "receipt.merchant_copy"
	KeyReceiptThankYou       TranslationKey = "receipt.thank_you"

	// Error messages
	KeyErrorConnectionFailed    TranslationKey = "error.connection_failed"
	KeyErrorTimeout             TranslationKey = "error.timeout"
	KeyErrorInvalidAmount       TranslationKey = "error.invalid_amount"
	KeyErrorInvalidCurrency     TranslationKey = "error.invalid_currency"
	KeyErrorInvalidCard         TranslationKey = "error.invalid_card"
	KeyErrorTransactionDeclined TranslationKey = "error.transaction_declined"
	KeyErrorInsufficientFunds   TranslationKey = "error.insufficient_funds"
	KeyErrorExpiredCard         TranslationKey = "error.expired_card"
	KeyErrorIncorrectPIN        TranslationKey = "error.incorrect_pin"
	KeyErrorIssuerUnavailable   TranslationKey = "error.issuer_unavailable"
	KeyErrorSystemError         TranslationKey = "error.system_error"

	// Card types
	KeyCardVisa       TranslationKey = "card.visa"
	KeyCardMastercard TranslationKey = "card.mastercard"
	KeyCardMada       TranslationKey = "card.mada"
	KeyCardKNET       TranslationKey = "card.knet"
	KeyCardAmex       TranslationKey = "card.amex"

	// Settlement
	KeySettlementBatchNumber TranslationKey = "settlement.batch_number"
	KeySettlementTotalCount  TranslationKey = "settlement.total_count"
	KeySettlementTotalAmount TranslationKey = "settlement.total_amount"
	KeySettlementSuccessful  TranslationKey = "settlement.successful"

	// Common
	KeyYes    TranslationKey = "common.yes"
	KeyNo     TranslationKey = "common.no"
	KeyTotal  TranslationKey = "common.total"
	KeyDate   TranslationKey = "common.date"
	KeyTime   TranslationKey = "common.time"
	KeyStatus TranslationKey = "common.status"
)

// Translator provides translation services
type Translator struct {
	mu           sync.RWMutex
	translations map[Language]map[TranslationKey]string
}

// NewTranslator creates a new translator with default translations
func NewTranslator() *Translator {
	t := &Translator{
		translations: make(map[Language]map[TranslationKey]string),
	}
	t.loadDefaultTranslations()
	return t
}

// loadDefaultTranslations loads English and Arabic translations
func (t *Translator) loadDefaultTranslations() {
	// English translations
	t.translations[LanguageEnglish] = englishTranslations()

	// Arabic translations
	t.translations[LanguageArabic] = arabicTranslations()
}

// Translate returns the translation for a given key and language
func (t *Translator) Translate(key TranslationKey, lang Language) string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if langMap, ok := t.translations[lang]; ok {
		if translation, ok := langMap[key]; ok {
			return translation
		}
	}

	// Fallback to English
	if lang != LanguageEnglish {
		if langMap, ok := t.translations[LanguageEnglish]; ok {
			if translation, ok := langMap[key]; ok {
				return translation
			}
		}
	}

	// If no translation found, return the key
	return string(key)
}

// TranslateBilingual returns both Arabic and English translations
func (t *Translator) TranslateBilingual(key TranslationKey) (arabic, english string) {
	arabic = t.Translate(key, LanguageArabic)
	english = t.Translate(key, LanguageEnglish)
	return
}

// FormatBilingual formats a bilingual string (Arabic on top, English below)
func (t *Translator) FormatBilingual(key TranslationKey) string {
	ar, en := t.TranslateBilingual(key)
	return fmt.Sprintf("%s\n%s", ar, en)
}

// FormatBilingualLine formats a bilingual line (Arabic - English)
func (t *Translator) FormatBilingualLine(key TranslationKey) string {
	ar, en := t.TranslateBilingual(key)
	return fmt.Sprintf("%s - %s", ar, en)
}

// AddTranslation adds or updates a translation
func (t *Translator) AddTranslation(lang Language, key TranslationKey, value string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if _, ok := t.translations[lang]; !ok {
		t.translations[lang] = make(map[TranslationKey]string)
	}
	t.translations[lang][key] = value
}

// GetSupportedLanguages returns all supported languages
func (t *Translator) GetSupportedLanguages() []Language {
	t.mu.RLock()
	defer t.mu.RUnlock()

	langs := make([]Language, 0, len(t.translations))
	for lang := range t.translations {
		langs = append(langs, lang)
	}
	return langs
}

// englishTranslations returns English translation map
func englishTranslations() map[TranslationKey]string {
	return map[TranslationKey]string{
		// Transaction types
		KeyTransactionSale:       "Sale",
		KeyTransactionVoid:       "Void",
		KeyTransactionRefund:     "Refund",
		KeyTransactionPreAuth:    "Pre-Authorization",
		KeyTransactionCompletion: "Completion",
		KeyTransactionSettlement: "Settlement",

		// Transaction status
		KeyStatusApproved: "Approved",
		KeyStatusDeclined: "Declined",
		KeyStatusError:    "Error",

		// Receipt fields
		KeyReceiptMerchantName:   "Merchant Name",
		KeyReceiptMerchantID:     "Merchant ID",
		KeyReceiptTerminalID:     "Terminal ID",
		KeyReceiptDate:           "Date",
		KeyReceiptTime:           "Time",
		KeyReceiptCardNumber:     "Card Number",
		KeyReceiptCardType:       "Card Type",
		KeyReceiptTransactionID:  "Transaction ID",
		KeyReceiptAuthCode:       "Auth Code",
		KeyReceiptAmount:         "Amount",
		KeyReceiptResponse:       "Response",
		KeyReceiptSignature:      "Signature",
		KeyReceiptCustomerCopy:   "Customer Copy",
		KeyReceiptMerchantCopy:   "Merchant Copy",
		KeyReceiptThankYou:       "Thank You",

		// Error messages
		KeyErrorConnectionFailed:    "Connection failed",
		KeyErrorTimeout:             "Transaction timeout",
		KeyErrorInvalidAmount:       "Invalid amount",
		KeyErrorInvalidCurrency:     "Invalid currency",
		KeyErrorInvalidCard:         "Invalid card",
		KeyErrorTransactionDeclined: "Transaction declined",
		KeyErrorInsufficientFunds:   "Insufficient funds",
		KeyErrorExpiredCard:         "Card expired",
		KeyErrorIncorrectPIN:        "Incorrect PIN",
		KeyErrorIssuerUnavailable:   "Issuer unavailable",
		KeyErrorSystemError:         "System error",

		// Card types
		KeyCardVisa:       "Visa",
		KeyCardMastercard: "Mastercard",
		KeyCardMada:       "mada",
		KeyCardKNET:       "KNET",
		KeyCardAmex:       "American Express",

		// Settlement
		KeySettlementBatchNumber: "Batch Number",
		KeySettlementTotalCount:  "Total Transactions",
		KeySettlementTotalAmount: "Total Amount",
		KeySettlementSuccessful:  "Settlement Successful",

		// Common
		KeyYes:    "Yes",
		KeyNo:     "No",
		KeyTotal:  "Total",
		KeyDate:   "Date",
		KeyTime:   "Time",
		KeyStatus: "Status",
	}
}

// Global translator instance
var globalTranslator *Translator
var translatorOnce sync.Once

// GetTranslator returns the global translator instance
func GetTranslator() *Translator {
	translatorOnce.Do(func() {
		globalTranslator = NewTranslator()
	})
	return globalTranslator
}

// T is a shorthand for translation
func T(key TranslationKey, lang Language) string {
	return GetTranslator().Translate(key, lang)
}

// TB is a shorthand for bilingual translation
func TB(key TranslationKey) (arabic, english string) {
	return GetTranslator().TranslateBilingual(key)
}
