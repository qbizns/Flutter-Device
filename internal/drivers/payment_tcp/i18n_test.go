package payment_tcp

import (
	"strings"
	"testing"
)

// TestTranslatorEnglish tests English translations
func TestTranslatorEnglish(t *testing.T) {
	tr := NewTranslator()

	tests := []struct {
		key      TranslationKey
		expected string
	}{
		{KeyTransactionSale, "Sale"},
		{KeyTransactionVoid, "Void"},
		{KeyStatusApproved, "Approved"},
		{KeyReceiptMerchantName, "Merchant Name"},
		{KeyErrorInsufficientFunds, "Insufficient funds"},
		{KeyCardMada, "mada"},
	}

	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			result := tr.Translate(tt.key, LanguageEnglish)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestTranslatorArabic tests Arabic translations
func TestTranslatorArabic(t *testing.T) {
	tr := NewTranslator()

	tests := []struct {
		key      TranslationKey
		expected string
	}{
		{KeyTransactionSale, "بيع"},
		{KeyTransactionVoid, "إلغاء"},
		{KeyStatusApproved, "موافق عليه"},
		{KeyReceiptMerchantName, "اسم التاجر"},
		{KeyErrorInsufficientFunds, "رصيد غير كافٍ"},
		{KeyCardMada, "مدى"},
	}

	for _, tt := range tests {
		t.Run(string(tt.key), func(t *testing.T) {
			result := tr.Translate(tt.key, LanguageArabic)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestTranslatorBilingual tests bilingual translations
func TestTranslatorBilingual(t *testing.T) {
	tr := NewTranslator()

	ar, en := tr.TranslateBilingual(KeyTransactionSale)
	if ar != "بيع" {
		t.Errorf("Expected Arabic 'بيع', got %q", ar)
	}
	if en != "Sale" {
		t.Errorf("Expected English 'Sale', got %q", en)
	}
}

// TestFormatBilingual tests bilingual formatting
func TestFormatBilingual(t *testing.T) {
	tr := NewTranslator()

	result := tr.FormatBilingual(KeyTransactionSale)
	expected := "بيع\nSale"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestFormatBilingualLine tests bilingual line formatting
func TestFormatBilingualLine(t *testing.T) {
	tr := NewTranslator()

	result := tr.FormatBilingualLine(KeyTransactionSale)
	expected := "بيع - Sale"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

// TestTranslatorFallback tests fallback to English
func TestTranslatorFallback(t *testing.T) {
	tr := NewTranslator()

	// Non-existent key should return the key itself
	result := tr.Translate("non.existent.key", LanguageEnglish)
	if result != "non.existent.key" {
		t.Errorf("Expected key as fallback, got %q", result)
	}
}

// TestAddTranslation tests adding custom translations
func TestAddTranslation(t *testing.T) {
	tr := NewTranslator()

	customKey := TranslationKey("custom.test")
	customValue := "Custom Translation"

	tr.AddTranslation(LanguageEnglish, customKey, customValue)

	result := tr.Translate(customKey, LanguageEnglish)
	if result != customValue {
		t.Errorf("Expected %q, got %q", customValue, result)
	}
}

// TestGetSupportedLanguages tests getting supported languages
func TestGetSupportedLanguages(t *testing.T) {
	tr := NewTranslator()

	langs := tr.GetSupportedLanguages()
	if len(langs) != 2 {
		t.Errorf("Expected 2 languages, got %d", len(langs))
	}

	// Check both English and Arabic are supported
	hasEnglish := false
	hasArabic := false
	for _, lang := range langs {
		if lang == LanguageEnglish {
			hasEnglish = true
		}
		if lang == LanguageArabic {
			hasArabic = true
		}
	}

	if !hasEnglish {
		t.Error("English language not found")
	}
	if !hasArabic {
		t.Error("Arabic language not found")
	}
}

// TestGlobalTranslator tests global translator instance
func TestGlobalTranslator(t *testing.T) {
	tr1 := GetTranslator()
	tr2 := GetTranslator()

	// Should return the same instance
	if tr1 != tr2 {
		t.Error("Global translator should return the same instance")
	}
}

// TestShorthandFunctions tests T and TB shorthand functions
func TestShorthandFunctions(t *testing.T) {
	result := T(KeyTransactionSale, LanguageEnglish)
	if result != "Sale" {
		t.Errorf("Expected 'Sale', got %q", result)
	}

	ar, en := TB(KeyTransactionSale)
	if ar != "بيع" || en != "Sale" {
		t.Errorf("Expected 'بيع' and 'Sale', got %q and %q", ar, en)
	}
}

// TestArabicResponseMessages tests Arabic response code messages
func TestArabicResponseMessages(t *testing.T) {
	tests := []struct {
		code     string
		contains string
	}{
		{"00", "موافق"},
		{"51", "رصيد"},
		{"54", "منتهية"},
		{"55", "رقم سري"},
		{"91", "غير متاحة"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			msg := GetArabicResponseMessage(tt.code)
			if !strings.Contains(msg, tt.contains) {
				t.Errorf("Expected message to contain %q, got %q", tt.contains, msg)
			}
		})
	}
}

// TestGetArabicResponseMessageUnknown tests unknown response code
func TestGetArabicResponseMessageUnknown(t *testing.T) {
	msg := GetArabicResponseMessage("XX")
	if !strings.Contains(msg, "خطأ") {
		t.Errorf("Expected error message, got %q", msg)
	}
}

// TestArabicCardSchemes tests Arabic card scheme names
func TestArabicCardSchemes(t *testing.T) {
	tests := []struct {
		scheme   string
		expected string
	}{
		{"visa", "فيزا"},
		{"mastercard", "ماستركارد"},
		{"mada", "مدى"},
		{"knet", "كي نت"},
		{"amex", "أمريكان إكسبريس"},
	}

	for _, tt := range tests {
		t.Run(tt.scheme, func(t *testing.T) {
			result := GetArabicCardScheme(tt.scheme)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestGetArabicCardSchemeUnknown tests unknown card scheme
func TestGetArabicCardSchemeUnknown(t *testing.T) {
	scheme := "unknown"
	result := GetArabicCardScheme(scheme)
	if result != scheme {
		t.Errorf("Expected %q, got %q", scheme, result)
	}
}

// TestArabicCurrencies tests Arabic currency names
func TestArabicCurrencies(t *testing.T) {
	tests := []struct {
		code     string
		expected string
	}{
		{"SAR", "ريال سعودي"},
		{"KWD", "دينار كويتي"},
		{"AED", "درهم إماراتي"},
		{"USD", "دولار أمريكي"},
		{"EUR", "يورو"},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			result := GetArabicCurrency(tt.code)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestFormatArabicAmount tests Arabic amount formatting
func TestFormatArabicAmount(t *testing.T) {
	tests := []struct {
		name     string
		amount   int64
		currency string
		contains []string
	}{
		{
			name:     "SAR 100.00",
			amount:   10000,
			currency: "SAR",
			contains: []string{"100.00", "ريال"},
		},
		{
			name:     "KWD 1.000",
			amount:   1000,
			currency: "KWD",
			contains: []string{"1.00", "دينار"},
		},
		{
			name:     "AED 50.00",
			amount:   5000,
			currency: "AED",
			contains: []string{"50.00", "درهم"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatArabicAmount(tt.amount, tt.currency)
			for _, contains := range tt.contains {
				if !strings.Contains(result, contains) {
					t.Errorf("Expected result to contain %q, got %q", contains, result)
				}
			}
		})
	}
}

// TestConvertToArabicIndic tests Arabic-Indic numeral conversion
func TestConvertToArabicIndic(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0", "٠"},
		{"123", "١٢٣"},
		{"456789", "٤٥٦٧٨٩"},
		{"100.00", "١٠٠.٠٠"},
		{"ABC123", "ABC١٢٣"}, // Mixed text
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ConvertToArabicIndic(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestRTLMarks tests RTL and LTR mark functions
func TestRTLMarks(t *testing.T) {
	text := "مرحبا"

	rtl := WrapRTL(text)
	if !strings.HasPrefix(rtl, RTLMark) || !strings.HasSuffix(rtl, RTLMark) {
		t.Errorf("WrapRTL should wrap with RTL marks")
	}

	ltr := WrapLTR("Hello")
	if !strings.HasPrefix(ltr, LRMark) || !strings.HasSuffix(ltr, LRMark) {
		t.Errorf("WrapLTR should wrap with LTR marks")
	}
}

// TestAllTranslationKeysHaveEnglish verifies all keys have English translations
func TestAllTranslationKeysHaveEnglish(t *testing.T) {
	tr := NewTranslator()

	// All defined keys
	keys := []TranslationKey{
		KeyTransactionSale, KeyTransactionVoid, KeyTransactionRefund,
		KeyTransactionPreAuth, KeyTransactionCompletion, KeyTransactionSettlement,
		KeyStatusApproved, KeyStatusDeclined, KeyStatusError,
		KeyReceiptMerchantName, KeyReceiptMerchantID, KeyReceiptTerminalID,
		KeyReceiptDate, KeyReceiptTime, KeyReceiptCardNumber, KeyReceiptCardType,
		KeyReceiptTransactionID, KeyReceiptAuthCode, KeyReceiptAmount,
		KeyReceiptResponse, KeyReceiptSignature, KeyReceiptCustomerCopy,
		KeyReceiptMerchantCopy, KeyReceiptThankYou,
		KeyErrorConnectionFailed, KeyErrorTimeout, KeyErrorInvalidAmount,
		KeyErrorInvalidCurrency, KeyErrorInvalidCard, KeyErrorTransactionDeclined,
		KeyErrorInsufficientFunds, KeyErrorExpiredCard, KeyErrorIncorrectPIN,
		KeyErrorIssuerUnavailable, KeyErrorSystemError,
		KeyCardVisa, KeyCardMastercard, KeyCardMada, KeyCardKNET, KeyCardAmex,
		KeySettlementBatchNumber, KeySettlementTotalCount, KeySettlementTotalAmount,
		KeySettlementSuccessful,
		KeyYes, KeyNo, KeyTotal, KeyDate, KeyTime, KeyStatus,
	}

	for _, key := range keys {
		result := tr.Translate(key, LanguageEnglish)
		if result == string(key) {
			t.Errorf("Key %q has no English translation", key)
		}
	}
}

// TestAllTranslationKeysHaveArabic verifies all keys have Arabic translations
func TestAllTranslationKeysHaveArabic(t *testing.T) {
	tr := NewTranslator()

	// All defined keys
	keys := []TranslationKey{
		KeyTransactionSale, KeyTransactionVoid, KeyTransactionRefund,
		KeyTransactionPreAuth, KeyTransactionCompletion, KeyTransactionSettlement,
		KeyStatusApproved, KeyStatusDeclined, KeyStatusError,
		KeyReceiptMerchantName, KeyReceiptMerchantID, KeyReceiptTerminalID,
		KeyReceiptDate, KeyReceiptTime, KeyReceiptCardNumber, KeyReceiptCardType,
		KeyReceiptTransactionID, KeyReceiptAuthCode, KeyReceiptAmount,
		KeyReceiptResponse, KeyReceiptSignature, KeyReceiptCustomerCopy,
		KeyReceiptMerchantCopy, KeyReceiptThankYou,
		KeyErrorConnectionFailed, KeyErrorTimeout, KeyErrorInvalidAmount,
		KeyErrorInvalidCurrency, KeyErrorInvalidCard, KeyErrorTransactionDeclined,
		KeyErrorInsufficientFunds, KeyErrorExpiredCard, KeyErrorIncorrectPIN,
		KeyErrorIssuerUnavailable, KeyErrorSystemError,
		KeyCardVisa, KeyCardMastercard, KeyCardMada, KeyCardKNET, KeyCardAmex,
		KeySettlementBatchNumber, KeySettlementTotalCount, KeySettlementTotalAmount,
		KeySettlementSuccessful,
		KeyYes, KeyNo, KeyTotal, KeyDate, KeyTime, KeyStatus,
	}

	for _, key := range keys {
		result := tr.Translate(key, LanguageArabic)
		if result == string(key) {
			t.Errorf("Key %q has no Arabic translation", key)
		}
		// Also verify it contains Arabic characters
		hasArabic := false
		for _, r := range result {
			if r >= 0x0600 && r <= 0x06FF {
				hasArabic = true
				break
			}
		}
		if !hasArabic && key != KeyCardVisa && key != KeyCardMastercard {
			// Some technical terms might use Latin script
			t.Logf("Warning: Key %q Arabic translation %q has no Arabic characters", key, result)
		}
	}
}

// TestConcurrentAccess tests thread-safe access to translator
func TestConcurrentAccess(t *testing.T) {
	tr := NewTranslator()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				_ = tr.Translate(KeyTransactionSale, LanguageEnglish)
				_ = tr.Translate(KeyTransactionSale, LanguageArabic)
				_, _ = tr.TranslateBilingual(KeyStatusApproved)
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

// BenchmarkTranslate benchmarks translation performance
func BenchmarkTranslate(b *testing.B) {
	tr := NewTranslator()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = tr.Translate(KeyTransactionSale, LanguageEnglish)
	}
}

// BenchmarkTranslateBilingual benchmarks bilingual translation
func BenchmarkTranslateBilingual(b *testing.B) {
	tr := NewTranslator()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = tr.TranslateBilingual(KeyTransactionSale)
	}
}

// BenchmarkFormatArabicAmount benchmarks Arabic amount formatting
func BenchmarkFormatArabicAmount(b *testing.B) {
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = FormatArabicAmount(10000, "SAR")
	}
}
