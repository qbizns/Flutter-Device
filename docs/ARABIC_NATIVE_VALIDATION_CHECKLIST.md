# Arabic Native Speaker Validation Checklist

**Phase 5: Native Arabic Speaker Validation**

This checklist ensures that Arabic text rendering meets the expectations and readability standards of native Arabic speakers.

## Purpose

Technical testing (Phase 4) validates that characters print correctly, but only native speakers can confirm:
- **Natural text flow** and reading experience
- **Proper glyph shaping** (contextual forms)
- **Cultural appropriateness** of phrases
- **Real-world usability** in Middle East markets

## Validation Team

### Required Validators

| Role | Requirements | Quantity |
|------|--------------|----------|
| **Native Arabic Speaker (Gulf dialect)** | Saudi Arabia, UAE, Kuwait, Qatar, Bahrain | 1-2 |
| **Native Arabic Speaker (Levantine dialect)** | Jordan, Lebanon, Syria, Palestine | 1 |
| **Native Arabic Speaker (Egyptian dialect)** | Egypt, Sudan | 1 |
| **Arabic Typography Expert** | Knowledge of Arabic calligraphy and digital typography | 1 (optional) |

**Total**: 3-5 validators minimum

### Validator Information Form

```markdown
**Name**: [Full name]
**Native Dialect**: [Gulf / Levantine / Egyptian / Maghrebi / Other]
**Country of Origin**: [Country]
**Years of Arabic Education**: [Number]
**Typography Experience**: [Yes / No]
**POS System Experience**: [Yes / No]
**Contact**: [Email]
**Date**: [YYYY-MM-DD]
```

---

## Validation Scenarios

### Scenario 1: Receipt Readability

**Objective**: Ensure printed receipts are natural and readable

#### Sample Receipt 1: Restaurant

```
مطعم البرج
Tower Restaurant
شارع الملك فهد، الرياض
King Fahd Road, Riyadh

رقم الطلب: ١٢٣٤
Order #: 1234

التاريخ: ٢٠٢٤/٠١/١٥  الوقت: ١٤:٣٠
Date: 2024/01/15  Time: 14:30

المنتج                    الكمية    السعر
Product                  Qty       Price
─────────────────────────────────────────
Big Mac بيج ماك           ٢        ٥٠ ر.س
بطاطس كبيرة               ١        ١٥ ر.س
Large Fries
كوكاكولا                  ٢        ١٠ ر.س
Coca-Cola
─────────────────────────────────────────
المجموع الفرعي:                    ٧٥ ر.س
Subtotal:                          75 SAR

الضريبة (١٥%):                    ١١.٢٥ ر.س
Tax (15%):                        11.25 SAR

الإجمالي:                         ٨٦.٢٥ ر.س
TOTAL:                            86.25 SAR

طريقة الدفع: بطاقة ائتمان
Payment: Credit Card

شكراً لزيارتكم!
Thank you for visiting!

للاستفسارات: ٩٢٠٠١٢٣٤٥
Inquiries: 920012345
```

**Validation Questions:**

1. **Overall Readability** (Scale: 1-5, 5 = Perfect)
   - [ ] 5 - Perfectly readable, natural flow
   - [ ] 4 - Mostly readable, minor issues
   - [ ] 3 - Readable with effort
   - [ ] 2 - Difficult to read
   - [ ] 1 - Unreadable

2. **Text Direction**
   - [ ] Arabic text flows naturally right-to-left
   - [ ] English text flows naturally left-to-right
   - [ ] Mixed text transitions smoothly
   - **Issues**: ___________________________

3. **Character Shaping**
   - [ ] Letters connect properly (e.g., م ر ح ب ا → مرحبا)
   - [ ] Isolated, initial, medial, final forms correct
   - [ ] No broken or disconnected characters
   - **Issues**: ___________________________

4. **Number Formatting**
   - [ ] Eastern Arabic numerals (٠-٩) readable
   - [ ] Western numerals (0-9) readable
   - [ ] Price formatting natural (٥٠ ر.س or 50 SAR)
   - **Preference**: [ ] Eastern [ ] Western [ ] Both

5. **Spacing and Alignment**
   - [ ] Column alignment proper
   - [ ] Word spacing natural
   - [ ] Line spacing comfortable
   - **Issues**: ___________________________

6. **Cultural Appropriateness**
   - [ ] Phrases sound natural
   - [ ] No awkward translations
   - [ ] Professional tone appropriate
   - **Suggestions**: ___________________________

**Overall Rating**: ☐ Excellent ☐ Good ☐ Acceptable ☐ Needs Improvement ☐ Unacceptable

**Notes**:
```
[Detailed feedback]
```

---

### Scenario 2: Retail Store Receipt

#### Sample Receipt 2: Grocery Store

```
سوبر ماركت الأسرة
Family Supermarket

الفرع: الرياض - حي العليا
Branch: Riyadh - Al Olaya

الفاتورة رقم: INV-2024-001234
Invoice No: INV-2024-001234

الصراف: أحمد محمد
Cashier: Ahmed Mohamed

التاريخ: ٢٠٢٤/٠١/١٥  الساعة: ١٠:٣٥ صباحاً
Date: 2024/01/15  Time: 10:35 AM

المنتج                الوحدة    الكمية    السعر
─────────────────────────────────────────────
حليب نادك كامل الدسم
Nadec Full Cream Milk
كرتون                  ١         ٣٥ ر.س

أرز أبيض ٥ كجم
White Rice 5kg
كيس                    ٢         ٦٠ ر.س

دجاج طازج
Fresh Chicken
كجم                   ١.٥        ٤٥ ر.س

خبز عربي طازج
Fresh Arabic Bread
كيس                    ٣         ٩ ر.س
─────────────────────────────────────────────
إجمالي المشتريات (٧ قطع):      ١٤٩ ر.س
Total Items (7 pcs):            149 SAR

خصم العميل الدائم (٥%):        ٧.٤٥ ر.س-
Loyalty Discount (5%):          7.45 SAR-

الضريبة المضافة (١٥%):         ٢١.٢٣ ر.س
VAT (15%):                     21.23 SAR

المبلغ المستحق:                ١٦٢.٧٨ ر.س
Amount Due:                    162.78 SAR

طريقة الدفع: نقداً
Payment Method: Cash

المدفوع:                       ٢٠٠ ر.س
Paid:                         200 SAR

المتبقي:                       ٣٧.٢٢ ر.س
Change:                       37.22 SAR

نقاط المكافآت المكتسبة: ١٦ نقطة
Rewards Points Earned: 16 pts

شكراً لتسوقكم معنا!
Thank you for shopping with us!

نرجو الاحتفاظ بالفاتورة
Please keep your receipt

للشكاوى والاقتراحات: ٩٢٠٠٣٣٣٣٣
Complaints & Suggestions: 920033333

زورونا على:
Visit us at: www.familymarket.sa
```

**Validation Questions:**

1. **Product Names**
   - [ ] Arabic product names natural
   - [ ] Translations accurate
   - [ ] No confusion in mixed-language names
   - **Issues**: ___________________________

2. **Units and Quantities**
   - [ ] Unit labels clear (كجم، كرتون، كيس)
   - [ ] Quantity formats natural
   - [ ] Decimal points clear (١.٥)
   - **Issues**: ___________________________

3. **Financial Information**
   - [ ] Discount wording clear
   - [ ] Tax label appropriate
   - [ ] Total calculations easy to follow
   - **Issues**: ___________________________

4. **Professional Terminology**
   - [ ] "Invoice" (فاتورة) used correctly
   - [ ] "Cashier" (صراف) appropriate
   - [ ] "Loyalty" (العميل الدائم) natural
   - **Suggestions**: ___________________________

**Overall Rating**: ☐ Excellent ☐ Good ☐ Acceptable ☐ Needs Improvement ☐ Unacceptable

**Notes**:
```
[Detailed feedback]
```

---

### Scenario 3: Pharmacy Receipt

#### Sample Receipt 3: Medical/Pharmacy

```
صيدلية النهدي
Al Nahdi Pharmacy

العنوان: طريق الملك عبدالله، جدة
Address: King Abdullah Road, Jeddah

رقم الوصفة الطبية: RX-2024-5678
Prescription No: RX-2024-5678

اسم المريض: [سري]
Patient Name: [Confidential]

التاريخ: ٢٠٢٤/٠١/١٥
Date: 2024/01/15

الدواء                   الجرعة    الكمية    السعر
──────────────────────────────────────────────
بنادول إكسترا
Panadol Extra
حبة ٥٠٠ ملغم             علبة ١     ١٥ ر.س

مضاد حيوي أموكسيسيلين
Amoxicillin Antibiotic
كبسولة ٥٠٠ ملغم          علبة ١     ٤٥ ر.س

فيتامين د٣
Vitamin D3
قرص ١٠٠٠ وحدة دولية     علبة ١     ٣٥ ر.س
──────────────────────────────────────────────
المجموع:                           ٩٥ ر.س
Total:                            95 SAR

تغطية التأمين (٨٠%):              ٧٦ ر.س
Insurance Coverage (80%):         76 SAR

المبلغ المطلوب:                   ١٩ ر.س
Amount Payable:                  19 SAR

⚠️ تعليمات مهمة / Important Instructions:
───────────────────────────────────────────
• تناول المضاد الحيوي لمدة ٧ أيام كاملة
  Take antibiotic for full 7 days

• لا تتوقف عن تناول الدواء دون استشارة الطبيب
  Don't stop medication without consulting doctor

• يُحفظ في مكان بارد وجاف
  Store in cool, dry place

الصيدلي: د. خالد أحمد - رخصة: ١٢٣٤٥
Pharmacist: Dr. Khalid Ahmed - License: 12345

للطوارئ: ٩٢٠٠٠٠٧١١
Emergency: 920000711

صحتكم تهمنا!
Your health matters!
```

**Validation Questions:**

1. **Medical Terminology**
   - [ ] Drug names correct (بنادول، أموكسيسيلين)
   - [ ] Dosage terms clear (ملغم، وحدة دولية)
   - [ ] Instructions understandable
   - **Issues**: ___________________________

2. **Safety Information**
   - [ ] Warning symbol (⚠️) appropriate
   - [ ] Instructions clear and unambiguous
   - [ ] Tone appropriate for medical context
   - **Suggestions**: ___________________________

3. **Professional Standards**
   - [ ] Pharmacist title format correct (د. = دكتور)
   - [ ] License number presentation clear
   - [ ] Privacy respected ([سري] for patient name)
   - **Issues**: ___________________________

**Overall Rating**: ☐ Excellent ☐ Good ☐ Acceptable ☐ Needs Improvement ☐ Unacceptable

**Notes**:
```
[Detailed feedback]
```

---

### Scenario 4: Common Phrases

**Objective**: Validate frequently used POS phrases

Rate each phrase on naturalness (1-5, 5 = Perfect):

| # | Arabic | English | Rating | Issues/Suggestions |
|---|--------|---------|--------|-------------------|
| 1 | مرحباً بكم | Welcome | ☐1 ☐2 ☐3 ☐4 ☐5 | |
| 2 | شكراً لزيارتكم | Thank you for visiting | ☐1 ☐2 ☐3 ☐4 ☐5 | |
| 3 | المجموع الفرعي | Subtotal | ☐1 ☐2 ☐3 ☐4 ☐5 | |
| 4 | الإجمالي | Total | ☐1 ☐2 ☐3 ☐4 ☐5 | |
| 5 | الضريبة | Tax | ☐1 ☐2 ☐3 ☐4 ☐5 | |
| 6 | رقم الفاتورة | Invoice Number | ☐1 ☐2 ☐3 ☐4 ☐5 | |
| 7 | التاريخ | Date | ☐1 ☐2 ☐3 ☐4 ☐5 | |
| 8 | الوقت | Time | ☐1 ☐2 ☐3 ☐4 ☐5 | |
| 9 | الكمية | Quantity | ☐1 ☐2 ☐3 ☐4 ☐5 | |
| 10 | السعر | Price | ☐1 ☐2 ☐3 ☐4 ☐5 | |
| 11 | المنتج | Product | ☐1 ☐2 ☐3 ☐4 ☐5 | |
| 12 | الخصم | Discount | ☐1 ☐2 ☐3 ☐4 ☐5 | |
| 13 | طريقة الدفع | Payment Method | ☐1 ☐2 ☐3 ☐4 ☐5 | |
| 14 | نقداً | Cash | ☐1 ☐2 ☐3 ☐4 ☐5 | |
| 15 | بطاقة ائتمان | Credit Card | ☐1 ☐2 ☐3 ☐4 ☐5 | |

**Alternative Suggestions**:
```
[List any alternative phrasings that would be more natural]
```

---

### Scenario 5: Currency Formats

**Objective**: Validate currency display preferences

#### Option A: Eastern Arabic Numerals
```
السعر: ٢٥٫٥٠ ر.س
السعر: ٧٥ ر.س
المجموع: ١٬٢٣٤٫٥٦ ر.س
```

#### Option B: Western Numerals
```
السعر: 25.50 ر.س
السعر: 75 ر.س
المجموع: 1,234.56 ر.س
```

#### Option C: Mixed (Label Arabic, Numbers Western)
```
السعر: 25.50 ر.س
Price: 25.50 SAR
```

**Questions**:

1. **Which format do you prefer?**
   - [ ] Option A (Eastern Arabic numerals)
   - [ ] Option B (Western numerals)
   - [ ] Option C (Mixed)
   - [ ] Other: ___________________

2. **Which is most common in your region?**
   - [ ] Option A
   - [ ] Option B
   - [ ] Option C

3. **For business receipts specifically:**
   - [ ] Option A
   - [ ] Option B
   - [ ] Option C

4. **Decimal separator preference:**
   - [ ] Comma (٫) - Arabic style
   - [ ] Period (.) - Western style
   - [ ] Either is fine

5. **Thousands separator preference:**
   - [ ] Comma (،) - Arabic style
   - [ ] Comma (,) - Western style
   - [ ] Space ( )
   - [ ] None

**Reasoning**:
```
[Explain preference]
```

---

### Scenario 6: Typography and Aesthetics

**Objective**: Assess visual quality of Arabic text

Print 5 sample lines with different characteristics:

**Sample 1**: Plain Arabic text
```
مرحباً بكم في مطعمنا
```

**Sample 2**: Bold Arabic text
```
**المجموع: ١٠٠ ر.س**
```

**Sample 3**: Mixed script
```
Coca-Cola كوكاكولا
```

**Sample 4**: Long text (wrapping)
```
محل البقالة الكبير للمواد الغذائية والمشروبات
والمنتجات الطازجة والمجمدة
```

**Sample 5**: With diacritics
```
مَرْحَباً بِكُمْ
```

**Evaluation Questions**:

1. **Letter Forms**
   - [ ] All letters clearly distinguishable
   - [ ] Proper contextual forms (initial/medial/final)
   - [ ] No broken connections
   - **Issues**: ___________________________

2. **Letter Spacing**
   - [ ] Spacing natural between letters
   - [ ] No excessive gaps
   - [ ] No overlapping characters
   - **Issues**: ___________________________

3. **Word Spacing**
   - [ ] Clear separation between words
   - [ ] Spacing consistent
   - [ ] Comfortable reading rhythm
   - **Issues**: ___________________________

4. **Line Spacing**
   - [ ] Adequate space between lines
   - [ ] Diacritics don't overlap previous line
   - [ ] Comfortable multi-line reading
   - **Issues**: ___________________________

5. **Aesthetics**
   - [ ] Professional appearance
   - [ ] Comparable to commercial receipts
   - [ ] Acceptable for business use
   - **Rating** (1-5): ☐1 ☐2 ☐3 ☐4 ☐5

**Comparison to Commercial Receipts**:
```
How does this compare to receipts from:
- McDonald's: ☐ Better ☐ Same ☐ Worse
- Local grocery stores: ☐ Better ☐ Same ☐ Worse
- Pharmacies: ☐ Better ☐ Same ☐ Worse
```

---

## Aggregated Feedback

### Summary Questions

1. **Would you trust a receipt printed with this system?**
   - [ ] Yes, completely
   - [ ] Yes, with minor reservations
   - [ ] Maybe, needs improvement
   - [ ] No, significant issues
   - **Explain**: ___________________________

2. **Would you recommend this system for use in your country?**
   - [ ] Yes, ready for production
   - [ ] Yes, after addressing minor issues
   - [ ] No, needs major improvements
   - [ ] No, fundamental problems exist
   - **Explain**: ___________________________

3. **Top 3 Issues** (if any):
   ```
   1. ___________________________
   2. ___________________________
   3. ___________________________
   ```

4. **Top 3 Strengths**:
   ```
   1. ___________________________
   2. ___________________________
   3. ___________________________
   ```

5. **Overall Quality Rating** (1-10, 10 = Perfect):
   ```
   Score: ___/10
   ```

6. **Readiness for Production**:
   - [ ] Ready now
   - [ ] Ready in 1-2 weeks with minor fixes
   - [ ] Ready in 1 month with moderate fixes
   - [ ] Needs major overhaul (2+ months)

---

## Additional Feedback

### Free-Form Comments

```
[Please provide any additional observations, suggestions, or concerns that weren't covered in the above questions.]
```

### Comparison to Existing Systems

**Have you used POS systems with Arabic support before?**
- [ ] Yes
- [ ] No

**If yes, which systems?**
```
[List systems]
```

**How does this compare?**
```
[Comparison feedback]
```

### Regional Variations

**Are there dialect-specific issues?**
```
[Describe any dialect-specific observations]
```

**Should the system support multiple Arabic variants?**
- [ ] Yes, essential
- [ ] Yes, nice to have
- [ ] No, standard Arabic is sufficient

---

## Validation Sign-Off

**Validator Information:**
- **Name**: ___________________________
- **Dialect**: ___________________________
- **Date**: ___________________________
- **Overall Approval**: ☐ Approved ☐ Approved with conditions ☐ Not approved

**Signature**: ___________________________

---

## Next Steps After Validation

1. **Compile Results**
   - Aggregate feedback from all validators
   - Identify common issues
   - Prioritize fixes

2. **Create Action Items**
   - Log critical issues in GitHub
   - Document phrase alternatives
   - Update translation dictionary

3. **Iterate**
   - Implement fixes
   - Revalidate with same validators
   - Obtain final approval

4. **Document**
   - Update GAPS.md with remaining issues
   - Document regional variations
   - Create best practices guide

5. **Production Release**
   - Final sign-off from validators
   - Release notes with Arabic support details
   - Monitor production feedback

---

## Resources for Validators

### Background Reading

- **Project Overview**: `README.md`
- **Arabic Integration Design**: `docs/ARABIC_INTEGRATION_DESIGN.md`
- **Known Gaps**: `GAPS.md`
- **Hardware Testing**: `docs/ARABIC_PRINTER_TESTING_GUIDE.md`

### Support Contacts

- **Technical Lead**: [Email]
- **Project Manager**: [Email]
- **Validation Coordinator**: [Email]

### Validation Session

- **Duration**: 60-90 minutes
- **Location**: [Office/Remote]
- **Equipment**: Sample receipts, rating forms, camera for documentation
- **Compensation**: [If applicable]

---

## Appendix: Arabic Typography Basics

### Contextual Forms

Arabic letters change shape based on position:

| Letter | Isolated | Initial | Medial | Final |
|--------|----------|---------|--------|-------|
| ب (Beh) | ب | بـ | ـبـ | ـب |
| م (Meem) | م | مـ | ـمـ | ـم |
| ر (Reh) | ر | رـ | ـرـ | ـر |

### Common Issues to Watch For

1. **Broken Connections**: Letters that should connect appear separated
2. **Wrong Forms**: Medial form used where final form should be
3. **Reversed Direction**: Text reads left-to-right instead of right-to-left
4. **Missing Diacritics**: Vowel marks don't appear
5. **Character Substitution**: Wrong character used (e.g., Latin 'o' for Arabic zero ٠)

### Quality Benchmarks

**Commercial Standard**: Text should be comparable to receipts from:
- McDonald's / Burger King (international chains)
- Carrefour / Danube (major retailers)
- Saudi Post / Aramex (shipping labels)

---

**End of Validation Checklist**
