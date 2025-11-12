package payment_tcp

import "fmt"

// arabicTranslations returns Arabic translation map
// All translations are in Modern Standard Arabic (MSA) suitable for Middle East markets
func arabicTranslations() map[TranslationKey]string {
	return map[TranslationKey]string{
		// Transaction types - أنواع المعاملات
		KeyTransactionSale:       "بيع",
		KeyTransactionVoid:       "إلغاء",
		KeyTransactionRefund:     "استرجاع",
		KeyTransactionPreAuth:    "تفويض مسبق",
		KeyTransactionCompletion: "إتمام",
		KeyTransactionSettlement: "تسوية",

		// Transaction status - حالة المعاملة
		KeyStatusApproved: "موافق عليه",
		KeyStatusDeclined: "مرفوض",
		KeyStatusError:    "خطأ",

		// Receipt fields - حقول الإيصال
		KeyReceiptMerchantName:   "اسم التاجر",
		KeyReceiptMerchantID:     "رقم التاجر",
		KeyReceiptTerminalID:     "رقم الجهاز",
		KeyReceiptDate:           "التاريخ",
		KeyReceiptTime:           "الوقت",
		KeyReceiptCardNumber:     "رقم البطاقة",
		KeyReceiptCardType:       "نوع البطاقة",
		KeyReceiptTransactionID:  "رقم المعاملة",
		KeyReceiptAuthCode:       "رمز التفويض",
		KeyReceiptAmount:         "المبلغ",
		KeyReceiptResponse:       "الرد",
		KeyReceiptSignature:      "التوقيع",
		KeyReceiptCustomerCopy:   "نسخة العميل",
		KeyReceiptMerchantCopy:   "نسخة التاجر",
		KeyReceiptThankYou:       "شكراً لك",

		// Error messages - رسائل الخطأ
		KeyErrorConnectionFailed:    "فشل الاتصال",
		KeyErrorTimeout:             "انتهت مهلة المعاملة",
		KeyErrorInvalidAmount:       "مبلغ غير صالح",
		KeyErrorInvalidCurrency:     "عملة غير صالحة",
		KeyErrorInvalidCard:         "بطاقة غير صالحة",
		KeyErrorTransactionDeclined: "المعاملة مرفوضة",
		KeyErrorInsufficientFunds:   "رصيد غير كافٍ",
		KeyErrorExpiredCard:         "بطاقة منتهية الصلاحية",
		KeyErrorIncorrectPIN:        "رقم سري غير صحيح",
		KeyErrorIssuerUnavailable:   "جهة الإصدار غير متاحة",
		KeyErrorSystemError:         "خطأ في النظام",

		// Card types - أنواع البطاقات
		KeyCardVisa:       "فيزا",
		KeyCardMastercard: "ماستركارد",
		KeyCardMada:       "مدى",
		KeyCardKNET:       "كي نت",
		KeyCardAmex:       "أمريكان إكسبريس",

		// Settlement - التسوية
		KeySettlementBatchNumber: "رقم الدفعة",
		KeySettlementTotalCount:  "إجمالي المعاملات",
		KeySettlementTotalAmount: "المبلغ الإجمالي",
		KeySettlementSuccessful:  "تمت التسوية بنجاح",

		// Common - عام
		KeyYes:    "نعم",
		KeyNo:     "لا",
		KeyTotal:  "الإجمالي",
		KeyDate:   "التاريخ",
		KeyTime:   "الوقت",
		KeyStatus: "الحالة",
	}
}

// ArabicResponseMessages maps ISO 8583 response codes to Arabic messages
var ArabicResponseMessages = map[string]string{
	"00": "موافق عليه",                           // Approved
	"01": "يرجى الاتصال بجهة إصدار البطاقة",      // Refer to card issuer
	"03": "رقم تاجر غير صالح",                    // Invalid merchant
	"04": "احتفظ بالبطاقة",                        // Pick up card
	"05": "لا تكرم",                              // Do not honor
	"12": "معاملة غير صالحة",                    // Invalid transaction
	"13": "مبلغ غير صالح",                        // Invalid amount
	"14": "رقم بطاقة غير صالح",                   // Invalid card number
	"30": "خطأ في التنسيق",                       // Format error
	"41": "بطاقة مفقودة",                         // Lost card
	"43": "بطاقة مسروقة",                         // Stolen card
	"51": "رصيد غير كافٍ",                        // Insufficient funds
	"54": "بطاقة منتهية الصلاحية",                // Expired card
	"55": "رقم سري غير صحيح",                     // Incorrect PIN
	"57": "معاملة غير مسموح بها لحامل البطاقة",  // Transaction not permitted to cardholder
	"58": "معاملة غير مسموح بها للجهاز",          // Transaction not permitted to terminal
	"61": "تجاوز حد المبلغ",                      // Exceeds amount limit
	"62": "بطاقة محظورة",                         // Restricted card
	"65": "تجاوز حد عدد المعاملات",               // Exceeds transaction limit
	"75": "عدد محاولات الرقم السري تجاوز الحد",   // PIN tries exceeded
	"76": "رقم حساب غير صالح",                    // Invalid account
	"91": "جهة الإصدار غير متاحة",                // Issuer unavailable
	"92": "لا يمكن توجيه المعاملة",                // Cannot route transaction
	"94": "معاملة مكررة",                         // Duplicate transaction
	"96": "خلل في النظام",                        // System malfunction
}

// GetArabicResponseMessage returns the Arabic message for a response code
func GetArabicResponseMessage(code string) string {
	if msg, ok := ArabicResponseMessages[code]; ok {
		return msg
	}
	return "خطأ غير معروف" // Unknown error
}

// ArabicCardSchemes maps card scheme names to Arabic
var ArabicCardSchemes = map[string]string{
	"visa":       "فيزا",
	"mastercard": "ماستركارد",
	"mada":       "مدى",
	"knet":       "كي نت",
	"amex":       "أمريكان إكسبريس",
	"discover":   "ديسكفر",
	"jcb":        "جي سي بي",
	"unionpay":   "يونيون باي",
}

// GetArabicCardScheme returns the Arabic name for a card scheme
func GetArabicCardScheme(scheme string) string {
	if arabic, ok := ArabicCardSchemes[scheme]; ok {
		return arabic
	}
	return scheme
}

// ArabicCurrencies maps currency codes to Arabic names
var ArabicCurrencies = map[string]string{
	"SAR": "ريال سعودي",      // Saudi Riyal
	"KWD": "دينار كويتي",     // Kuwaiti Dinar
	"AED": "درهم إماراتي",    // UAE Dirham
	"BHD": "دينار بحريني",    // Bahraini Dinar
	"OMR": "ريال عماني",      // Omani Rial
	"QAR": "ريال قطري",       // Qatari Riyal
	"JOD": "دينار أردني",     // Jordanian Dinar
	"EGP": "جنيه مصري",       // Egyptian Pound
	"LBP": "ليرة لبنانية",    // Lebanese Pound
	"IQD": "دينار عراقي",     // Iraqi Dinar
	"USD": "دولار أمريكي",    // US Dollar
	"EUR": "يورو",            // Euro
}

// GetArabicCurrency returns the Arabic name for a currency code
func GetArabicCurrency(code string) string {
	if arabic, ok := ArabicCurrencies[code]; ok {
		return arabic
	}
	return code
}

// FormatArabicAmount formats an amount with Arabic currency name
// amount is in minor units (e.g., 10000 = 100.00)
func FormatArabicAmount(amount int64, currency string) string {
	currencyArabic := GetArabicCurrency(currency)

	// Different decimal places for different currencies
	var formattedAmount string
	switch currency {
	case "KWD", "BHD", "OMR", "JOD": // 3 decimal places
		formattedAmount = formatArabicNumber(float64(amount) / 1000.0)
	default: // 2 decimal places (SAR, AED, etc.)
		formattedAmount = formatArabicNumber(float64(amount) / 100.0)
	}

	return formattedAmount + " " + currencyArabic
}

// formatArabicNumber formats a number with Arabic-Indic numerals
// Note: While Arabic-Indic numerals (٠١٢٣٤٥٦٧٨٩) are traditional,
// modern receipts in Gulf countries often use Western Arabic numerals (0-9)
// This function uses Western numerals for better compatibility
func formatArabicNumber(num float64) string {
	// Use Western Arabic numerals for receipt compatibility
	return fmt.Sprintf("%.2f", num)
}

// ArabicDaysOfWeek maps day names to Arabic
var ArabicDaysOfWeek = map[string]string{
	"Sunday":    "الأحد",
	"Monday":    "الاثنين",
	"Tuesday":   "الثلاثاء",
	"Wednesday": "الأربعاء",
	"Thursday":  "الخميس",
	"Friday":    "الجمعة",
	"Saturday":  "السبت",
}

// ArabicMonths maps month names to Arabic
var ArabicMonths = map[string]string{
	"January":   "يناير",
	"February":  "فبراير",
	"March":     "مارس",
	"April":     "أبريل",
	"May":       "مايو",
	"June":      "يونيو",
	"July":      "يوليو",
	"August":    "أغسطس",
	"September": "سبتمبر",
	"October":   "أكتوبر",
	"November":  "نوفمبر",
	"December":  "ديسمبر",
}

// ArabicNumbers maps Western numerals to Arabic-Indic numerals
// Note: This is provided for completeness, but Western numerals are
// more commonly used in modern Middle Eastern payment receipts
var ArabicNumbers = map[rune]rune{
	'0': '٠',
	'1': '١',
	'2': '٢',
	'3': '٣',
	'4': '٤',
	'5': '٥',
	'6': '٦',
	'7': '٧',
	'8': '٨',
	'9': '٩',
}

// ConvertToArabicIndic converts Western numerals to Arabic-Indic numerals
func ConvertToArabicIndic(s string) string {
	result := make([]rune, 0, len(s))
	for _, r := range s {
		if arabic, ok := ArabicNumbers[r]; ok {
			result = append(result, arabic)
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

// RTLMark is the Unicode Right-to-Left Mark
const RTLMark = "\u200F"

// LRMark is the Unicode Left-to-Right Mark
const LRMark = "\u200E"

// WrapRTL wraps text with RTL marks for proper rendering
func WrapRTL(text string) string {
	return RTLMark + text + RTLMark
}

// WrapLTR wraps text with LTR marks for proper rendering
func WrapLTR(text string) string {
	return LRMark + text + LRMark
}
