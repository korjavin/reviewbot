---
id: owasp-mobile-m09-insecure-data-storage
title: OWASP Mobile M09:2024 — Insecure Data Storage
severity: high
tags: [owasp-mobile-top10, data-storage, sqlite, keychain, logs, mobile]
taxonomy: security/mobile/storage
references:
  - https://owasp.org/www-project-mobile-top-10/2023-risks/m9-insecure-data.html
  - https://mobile-security.gitbook.io/mobile-security-testing-guide/general-mobile-app-testing-guide/0x05d-testing-data-storage
---

# OWASP Mobile M09:2024 — Insecure Data Storage

## Description

Sensitive data (PII, tokens, financial data, health records) stored insecurely on mobile devices is accessible to other apps, device backups, or physical device access. Common mistakes: plaintext SQLite databases, unprotected files in external storage, PII in application logs, and caching sensitive data in browser/WebView.

## Vulnerable Pattern

```kotlin
// BAD — Android: writing sensitive data to external storage (world-readable pre-API 29)
val file = File(Environment.getExternalStorageDirectory(), "user_data.txt")
file.writeText("SSN: 123-45-6789, CardNumber: 4111111111111111")

// BAD — logging PII (visible in logcat, crash reports)
Log.d("UserProfile", "Loaded user: email=${user.email}, dob=${user.dob}, ssn=${user.ssn}")

// BAD — SQLite DB not encrypted
val db = openOrCreateDatabase("app.db", MODE_PRIVATE, null)
db.execSQL("INSERT INTO users VALUES ('${user.email}', '${user.password}')")
```

```swift
// BAD — iOS: storing sensitive data with wrong accessibility
let status = SecItemAdd([
    kSecClass: kSecClassGenericPassword,
    kSecAttrAccessible: kSecAttrAccessibleAlways,  // accessible even when locked/after reboot
    kSecValueData: tokenData
] as CFDictionary, nil)
```

## Secure Pattern

```kotlin
// GOOD — Android: use EncryptedFile for sensitive data
val masterKey = MasterKey.Builder(context)
    .setKeyScheme(MasterKey.KeyScheme.AES256_GCM)
    .build()
val encryptedFile = EncryptedFile.Builder(
    context, File(context.filesDir, "sensitive.dat"),
    masterKey, EncryptedFile.FileEncryptionScheme.AES256_GCM_HKDF_4KB
).build()

// GOOD — use SQLCipher for encrypted SQLite
val db = SQLiteDatabase.openOrCreateDatabase(path, SQLCipherUtils.getEncryptionKey(context), null)
```

## Checks to Generate

- Grep for `Log.d`, `Log.i`, `NSLog`, `print` containing `email`, `password`, `ssn`, `card`, `token`.
- Flag `getExternalStorageDirectory()` for sensitive file writes.
- Grep for `openOrCreateDatabase` without SQLCipher / Room with encryption.
- Flag `kSecAttrAccessibleAlways` or `kSecAttrAccessibleAfterFirstUnlock` for highly sensitive data — prefer `kSecAttrAccessibleWhenUnlockedThisDeviceOnly`.
- Check for `android:allowBackup="true"` in manifest — device backups include app data.
- Flag WebView cache paths storing authentication responses.
