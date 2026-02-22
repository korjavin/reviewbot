---
id: owasp-mobile-m06-inadequate-privacy-controls
title: OWASP Mobile M06:2024 — Inadequate Privacy Controls
severity: high
tags: [owasp-mobile-top10, privacy, pii, gdpr, permissions, mobile]
taxonomy: security/mobile/privacy
references:
  - https://owasp.org/www-project-mobile-top-10/2023-risks/m6-inadequate-privacy-controls.html
---

# OWASP Mobile M06:2024 — Inadequate Privacy Controls

## Description

Mobile apps collecting more data than needed, sharing PII with third-party SDKs without user consent, or failing to honor deletion requests violate privacy regulations (GDPR, CCPA, PDPA). Third-party analytics and advertising SDKs are a major source of inadvertent data sharing.

## Vulnerable Pattern

```kotlin
// BAD — requesting unnecessary permissions
// AndroidManifest.xml
<uses-permission android:name="android.permission.READ_CONTACTS"/>  // not needed for core feature
<uses-permission android:name="android.permission.ACCESS_FINE_LOCATION"/>  // approximate sufficient?
<uses-permission android:name="android.permission.READ_CALL_LOG"/>  // irrelevant to app function

// BAD — passing PII to third-party analytics without scrubbing
FirebaseAnalytics.getInstance(this).logEvent("purchase", Bundle().apply {
    putString("user_email", user.email)  // PII sent to Google
    putString("user_ssn", user.ssn)      // CRITICAL PII breach
    putDouble("amount", purchase.amount)
})
```

```swift
// BAD — iOS: tracking user across apps without ATT permission
let idfa = ASIdentifierManager.shared().advertisingIdentifier  // requires ATT consent since iOS 14
// Using without checking ATT authorization status
```

## Secure Pattern

```kotlin
// GOOD — minimal permissions, explain purpose
// Only request what is strictly needed; use coarse location when fine is not required
<uses-permission android:name="android.permission.ACCESS_COARSE_LOCATION"/>

// GOOD — strip PII before analytics
FirebaseAnalytics.getInstance(this).logEvent("purchase", Bundle().apply {
    putString("user_tier", user.tier)  // non-PII attribute
    putDouble("amount", purchase.amount)
    // email/SSN not included
})

// GOOD — honor data deletion
fun deleteUserData(userId: String) {
    db.query(User).filter(User.id == userId).delete()
    analyticsProvider.deleteUser(userId)  // delete from analytics too
    cloudBackup.deleteUserData(userId)
}
```

## Checks to Generate

- Scan `AndroidManifest.xml` for sensitive permissions (`READ_CONTACTS`, `READ_CALL_LOG`, `CAMERA`, `RECORD_AUDIO`) — verify they are necessary.
- Grep for PII fields (`email`, `ssn`, `dob`, `phone`) passed to analytics SDK calls.
- Check for ATT (App Tracking Transparency) implementation in iOS apps using IDFA.
- Flag third-party SDK initialization without privacy consent check.
- Check for absence of data deletion API/mechanism (GDPR right to erasure).
- Grep for device identifiers (`IMEI`, `IMSI`, `MAC address`) collected and transmitted.
