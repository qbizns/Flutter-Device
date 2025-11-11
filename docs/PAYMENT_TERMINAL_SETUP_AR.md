# إعداد محطة الدفع - دليل شامل
# Payment Terminal Setup - Complete Guide

**النسخة:** 2.0
**التاريخ:** 11 نوفمبر 2025
**اللغات:** العربية | English

---

## جدول المحتويات
## Table of Contents

1. [نظرة عامة](#نظرة-عامة) | [Overview](#نظرة-عامة)
2. [الشبكات المدعومة](#الشبكات-المدعومة) | [Supported Networks](#الشبكات-المدعومة)
3. [التثبيت](#التثبيت) | [Installation](#التثبيت)
4. [الإعداد السريع](#الإعداد-السريع) | [Quick Setup](#الإعداد-السريع)
5. [أمثلة الاستخدام](#أمثلة-الاستخدام) | [Usage Examples](#أمثلة-الاستخدام)
6. [الإيصالات ثنائية اللغة](#الإيصالات-ثنائية-اللغة) | [Bilingual Receipts](#الإيصالات-ثنائية-اللغة)
7. [استكشاف الأخطاء](#استكشاف-الأخطاء) | [Troubleshooting](#استكشاف-الأخطاء)
8. [الأمان والامتثال](#الأمان-والامتثال) | [Security & Compliance](#الأمان-والامتثال)

---

## نظرة عامة

برنامج تشغيل محطة الدفع هو حل شامل لمعالجة مدفوعات البطاقات في منطقة الشرق الأوسط. يدعم البرنامج الشبكات المحلية والدولية مع امتثال كامل لمعايير PCI DSS.

**Driver Features:**
- ✅ **Mada Support** - دعم شبكة مدى السعودية
- ✅ **KNET Support** - دعم شبكة كي نت الكويتية
- ✅ **Bilingual** - إيصالات ورسائل بالعربية والإنجليزية
- ✅ **PCI DSS Compliant** - امتثال كامل لمعايير أمان البطاقات
- ✅ **Offline Testing** - محاكي للاختبار بدون أجهزة
- ✅ **Comprehensive Audit** - سجل تدقيق شامل

---

## الشبكات المدعومة

### شبكة مدى (السعودية)
### Mada Network (Saudi Arabia)

**حول مدى:**
شبكة مدى هي الشبكة الوطنية للمدفوعات في المملكة العربية السعودية، تديرها شركة الشرق الأوسط للتقنيات المالية.

**التفاصيل التقنية:**
- **العملة:** ريال سعودي (SAR)
- **الحد الأدنى:** 1.00 ريال
- **الحد الأقصى:** 100,000.00 ريال
- **نطاقات BIN:** أكثر من 23 نطاق BIN

**أنواع المعاملات المدعومة:**
```
✅ بيع (Sale)
✅ إلغاء (Void)
✅ استرجاع (Refund)
✅ تفويض مسبق (Pre-Authorization)
✅ إتمام (Completion)
✅ تسوية (Settlement)
```

### شبكة كي نت (الكويت)
### KNET Network (Kuwait)

**حول كي نت:**
شبكة كي نت هي شبكة المدفوعات الوطنية في الكويت، تديرها شركة كي نت للمدفوعات الإلكترونية.

**التفاصيل التقنية:**
- **العملة:** دينار كويتي (KWD)
- **الحد الأدنى:** 0.100 دينار
- **الحد الأقصى:** 5,000.000 دينار
- **نطاقات BIN:** أكثر من 12 نطاق BIN

**أنواع المعاملات المدعومة:**
```
✅ بيع (Sale)
✅ إلغاء (Void)
✅ استرجاع (Refund)
✅ تسوية (Settlement)
⛔ التفويض المسبق غير مدعوم
```

---

## التثبيت

### المتطلبات

```bash
# Go 1.21 أو أحدث
go version

# مكتبات النظام (Linux)
sudo apt-get install build-essential
```

### التثبيت عبر Go

```bash
# تثبيت الحزمة
go get github.com/Macber-eg/Flutter-Device/internal/drivers/payment_tcp

# استيراد في الكود
import "github.com/Macber-eg/Flutter-Device/internal/drivers/payment_tcp"
```

---

## الإعداد السريع

### 1. المحاكي (للاختبار)

استخدم المحاكي للاختبار بدون أجهزة حقيقية:

```go
package main

import (
    "context"
    "fmt"
    "github.com/Macber-eg/Flutter-Device/internal/drivers/payment_tcp"
)

func main() {
    // إنشاء محطة محاكاة
    driver := payment_tcp.NewSimulatorDriver("test-001", "متجر الاختبار")

    ctx := context.Background()

    // الاتصال
    if err := driver.Connect(ctx); err != nil {
        fmt.Printf("خطأ في الاتصال: %v\n", err)
        return
    }
    defer driver.Disconnect(ctx)

    // إجراء عملية بيع
    req := payment_tcp.TransactionRequest{
        Type:     payment_tcp.TransactionSale,
        Amount:   10000, // 100.00 ريال
        Currency: "SAR",
    }

    resp, err := driver.ProcessTransaction(ctx, req)
    if err != nil {
        fmt.Printf("خطأ في المعاملة: %v\n", err)
        return
    }

    if resp.Success {
        fmt.Println("✅ المعاملة موافق عليها")
        fmt.Printf("رقم التفويض: %s\n", resp.AuthCode)
        fmt.Printf("رقم المعاملة: %s\n", resp.TransactionID)
    } else {
        fmt.Printf("❌ المعاملة مرفوضة: %s\n", resp.ResponseMessage)
    }
}
```

### 2. شبكة مدى (الإنتاج)

```go
package main

import (
    "context"
    "github.com/Macber-eg/Flutter-Device/internal/drivers/payment_tcp"
)

func main() {
    // إعداد الاتصال
    config := payment_tcp.ConnectionConfig{
        Host:       "mada.terminal.example.com",
        Port:       3000,
        TerminalID: "MADA_TERMINAL_001",
        MerchantID: "MERCHANT_12345",
        Provider:   "mada",
        UseTLS:     true,
    }

    // إنشاء برنامج التشغيل
    driver := payment_tcp.NewDriver("mada-001", "متجر الرياض", config)

    // إنشاء مزود مدى
    madaProvider := payment_tcp.NewMadaProvider(driver)

    ctx := context.Background()

    // الاتصال بالمحطة
    if err := madaProvider.Connect(ctx); err != nil {
        panic(err)
    }
    defer madaProvider.Disconnect(ctx)

    // معاملة البيع
    req := payment_tcp.TransactionRequest{
        Type:     payment_tcp.TransactionSale,
        Amount:   50000, // 500.00 ريال
        Currency: "SAR",
        Reference: "INV-2025-001",
    }

    resp, err := madaProvider.ProcessTransaction(ctx, req)
    if err != nil {
        panic(err)
    }

    // طباعة الإيصال
    for _, line := range resp.ReceiptLines {
        fmt.Println(line)
    }
}
```

### 3. شبكة كي نت (الإنتاج)

```go
package main

import (
    "context"
    "github.com/Macber-eg/Flutter-Device/internal/drivers/payment_tcp"
)

func main() {
    // إعداد الاتصال
    config := payment_tcp.ConnectionConfig{
        Host:       "knet.terminal.example.com",
        Port:       3000,
        TerminalID: "KNET_TERMINAL_001",
        MerchantID: "MERCHANT_67890",
        Provider:   "knet",
        UseTLS:     true,
    }

    // إنشاء برنامج التشغيل
    driver := payment_tcp.NewDriver("knet-001", "متجر الكويت", config)

    // إنشاء مزود كي نت
    knetProvider := payment_tcp.NewKNETProvider(driver)

    ctx := context.Background()

    // الاتصال بالمحطة
    if err := knetProvider.Connect(ctx); err != nil {
        panic(err)
    }
    defer knetProvider.Disconnect(ctx)

    // معاملة البيع
    req := payment_tcp.TransactionRequest{
        Type:     payment_tcp.TransactionSale,
        Amount:   25000, // 25.000 دينار
        Currency: "KWD",
        Reference: "INV-2025-001",
    }

    resp, err := knetProvider.ProcessTransaction(ctx, req)
    if err != nil {
        panic(err)
    }

    // طباعة الإيصال ثنائي اللغة
    for _, line := range resp.ReceiptLines {
        fmt.Println(line)
    }
}
```

---

## أمثلة الاستخدام

### عملية بيع كاملة

```go
// 1. إنشاء طلب البيع
saleReq := payment_tcp.TransactionRequest{
    Type:      payment_tcp.TransactionSale,
    Amount:    10000,  // 100.00 ريال
    Currency:  "SAR",
    Reference: "SALE-001",
}

// 2. معالجة المعاملة
saleResp, err := driver.ProcessTransaction(ctx, saleReq)
if err != nil {
    log.Printf("خطأ: %v", err)
    return
}

// 3. التحقق من النجاح
if saleResp.Success {
    fmt.Println("✅ تمت الموافقة على المعاملة")
    fmt.Printf("رمز التفويض: %s\n", saleResp.AuthCode)
    fmt.Printf("رقم المعاملة: %s\n", saleResp.TransactionID)
} else {
    fmt.Printf("❌ المعاملة مرفوضة: %s\n", saleResp.ResponseMessage)
}
```

### إلغاء معاملة

```go
// إلغاء معاملة موجودة
voidReq := payment_tcp.TransactionRequest{
    Type:              payment_tcp.TransactionVoid,
    Amount:            10000,
    Currency:          "SAR",
    OriginalReference: saleResp.TransactionID, // من المعاملة الأصلية
}

voidResp, err := driver.ProcessTransaction(ctx, voidReq)
if err != nil {
    log.Printf("خطأ في الإلغاء: %v", err)
    return
}

if voidResp.Success {
    fmt.Println("✅ تم إلغاء المعاملة بنجاح")
}
```

### استرجاع مبلغ

```go
// استرجاع معاملة سابقة
refundReq := payment_tcp.TransactionRequest{
    Type:              payment_tcp.TransactionRefund,
    Amount:            10000,
    Currency:          "SAR",
    OriginalReference: saleResp.TransactionID,
}

refundResp, err := driver.ProcessTransaction(ctx, refundReq)
if err != nil {
    log.Printf("خطأ في الاسترجاع: %v", err)
    return
}

if refundResp.Success {
    fmt.Println("✅ تم الاسترجاع بنجاح")
}
```

### التسوية اليومية

```go
// إجراء تسوية نهاية اليوم
settlementReq := payment_tcp.SettlementRequest{
    BatchNumber: time.Now().Format("20060102"), // YYYYMMDD
}

settlementResp, err := driver.Settlement(ctx, settlementReq)
if err != nil {
    log.Printf("خطأ في التسوية: %v", err)
    return
}

if settlementResp.Success {
    fmt.Println("✅ تمت التسوية بنجاح")
    fmt.Printf("إجمالي المعاملات: %d\n", settlementResp.TotalCount)
    fmt.Printf("المبلغ الإجمالي: %d\n", settlementResp.TotalAmount)
}
```

---

## الإيصالات ثنائية اللغة

يدعم البرنامج إيصالات ثنائية اللغة (عربي/إنجليزي) تلقائياً.

### مثال على إيصال مدى

```
========================================
           متجر الرياض
        Riyadh Store
========================================

نوع المعاملة      Transaction Type
بيع               Sale

رقم التاجر        Merchant ID
MERCHANT_12345    MERCHANT_12345

رقم الجهاز         Terminal ID
MADA_TERMINAL_001 MADA_TERMINAL_001

التاريخ           Date
2025-11-11        2025-11-11

الوقت             Time
14:30:45          14:30:45

رقم البطاقة       Card Number
****1234          ****1234

نوع البطاقة       Card Type
مدى               mada

رقم المعاملة      Transaction ID
TX20251111143045  TX20251111143045

رمز التفويض       Auth Code
AUTH123456        AUTH123456

المبلغ            Amount
500.00 ريال سعودي
500.00 SAR

الرد              Response
موافق عليه        Approved

شكراً لك          Thank You

نسخة العميل      Customer Copy
========================================
```

### استخدام الترجمة في الكود

```go
import "github.com/Macber-eg/Flutter-Device/internal/drivers/payment_tcp"

// الحصول على مترجم
translator := payment_tcp.GetTranslator()

// ترجمة مفتاح واحد
arabic := translator.Translate(payment_tcp.KeyTransactionSale, payment_tcp.LanguageArabic)
// النتيجة: "بيع"

english := translator.Translate(payment_tcp.KeyTransactionSale, payment_tcp.LanguageEnglish)
// النتيجة: "Sale"

// ترجمة ثنائية اللغة
ar, en := translator.TranslateBilingual(payment_tcp.KeyStatusApproved)
// ar = "موافق عليه", en = "Approved"

// تنسيق ثنائي اللغة
bilingual := translator.FormatBilingual(payment_tcp.KeyTransactionSale)
// النتيجة: "بيع\nSale"

// تنسيق سطر واحد
line := translator.FormatBilingualLine(payment_tcp.KeyTransactionSale)
// النتيجة: "بيع - Sale"
```

### رسائل الخطأ بالعربية

```go
// الحصول على رسالة خطأ بالعربية
errorMsg := payment_tcp.GetArabicResponseMessage("51")
// النتيجة: "رصيد غير كافٍ"

errorMsg = payment_tcp.GetArabicResponseMessage("54")
// النتيجة: "بطاقة منتهية الصلاحية"

errorMsg = payment_tcp.GetArabicResponseMessage("55")
// النتيجة: "رقم سري غير صحيح"
```

### تنسيق المبالغ بالعربية

```go
// تنسيق مبلغ بالريال السعودي
amount := payment_tcp.FormatArabicAmount(10000, "SAR")
// النتيجة: "100.00 ريال سعودي"

// تنسيق مبلغ بالدينار الكويتي
amount = payment_tcp.FormatArabicAmount(25000, "KWD")
// النتيجة: "25.00 دينار كويتي"

// تنسيق مبلغ بالدرهم الإماراتي
amount = payment_tcp.FormatArabicAmount(50000, "AED")
// النتيجة: "500.00 درهم إماراتي"
```

---

## استكشاف الأخطاء

### الأخطاء الشائعة وحلولها

#### 1. فشل الاتصال
```
خطأ: connection_failed
الحل: تحقق من عنوان IP والمنفذ
```

#### 2. مهلة المعاملة
```
خطأ: transaction_timeout
الحل: زيادة مهلة الانتظار في الإعدادات
```

#### 3. عملة غير صالحة
```
خطأ: invalid_currency
الحل: استخدم SAR لمدى، KWD لكي نت
```

#### 4. مبلغ خارج النطاق
```
خطأ: invalid_amount
مدى: 1.00 - 100,000.00 ريال
كي نت: 0.100 - 5,000.000 دينار
```

### تمكين سجلات التصحيح

```go
// تمكين السجلات التفصيلية
import "log"

log.SetFlags(log.LstdFlags | log.Lshortfile)

// سيتم طباعة جميع عمليات المعاملات
```

---

## الأمان والامتثال

### امتثال PCI DSS

✅ **إخفاء أرقام البطاقات**
- يتم إخفاء جميع أرقام البطاقات (****1234)
- لا يتم تخزين أرقام البطاقات الكاملة

✅ **التدقيق الشامل**
- سجل تدقيق لجميع المعاملات
- تتبع التاريخ والوقت
- تسجيل الأحداث الأمنية

✅ **التشفير**
- دعم TLS/SSL
- اتصال آمن بالمحطات
- لا يتم تخزين البيانات الحساسة

### أفضل الممارسات الأمنية

1. **استخدام TLS في الإنتاج**
```go
config := payment_tcp.ConnectionConfig{
    UseTLS:    true,
    TLSConfig: &tls.Config{
        MinVersion: tls.VersionTLS13,
    },
}
```

2. **التحقق من المعاملات**
```go
// تحقق دائماً من حالة النجاح
if !resp.Success {
    // معالجة الفشل
    log.Printf("فشلت المعاملة: %s", resp.ResponseMessage)
}
```

3. **التسوية اليومية**
```go
// قم بإجراء التسوية في نهاية كل يوم عمل
err := performDailySettlement()
```

---

## الدعم الفني

للحصول على الدعم الفني:

- **الوثائق:** راجع [PAYMENT_TERMINAL_SETUP.md](PAYMENT_TERMINAL_SETUP.md)
- **الأمان:** راجع [PAYMENT_SECURITY_GUIDE.md](PAYMENT_SECURITY_GUIDE.md)
- **المراقبة:** راجع [PAYMENT_MONITORING_GUIDE.md](PAYMENT_MONITORING_GUIDE.md)
- **البريد الإلكتروني:** support@example.com
- **الهاتف:** +966-XX-XXXXXXX (السعودية)
- **الهاتف:** +965-XXXX-XXXX (الكويت)

---

## رخصة الاستخدام

هذا البرنامج مرخص بموجب رخصة MIT. راجع ملف [LICENSE](../LICENSE) للتفاصيل.

---

**تم التحديث:** 11 نوفمبر 2025
**النسخة:** 2.0
**الحالة:** ✅ جاهز للإنتاج
