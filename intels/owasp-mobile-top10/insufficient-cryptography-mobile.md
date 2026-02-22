---
id: owasp-mobile-m10-insufficient-cryptography
title: OWASP Mobile M10:2024 — Insufficient Cryptography
severity: high
tags: [owasp-mobile-top10, cryptography, mobile, encryption, tls]
taxonomy: security/mobile/cryptography
references:
  - https://owasp.org/www-project-mobile-top-10/2023-risks/m10-insufficient-cryptography.html
  - https://mobile-security.gitbook.io/mobile-security-testing-guide/general-mobile-app-testing-guide/0x04g-testing-cryptography
---

# OWASP Mobile M10:2024 — Insufficient Cryptography

## Description

Mobile apps frequently misuse cryptography: using deprecated algorithms (MD5, SHA1, DES), ECB mode (deterministic, reveals patterns), hardcoded IV/salt values, or using encryption libraries incorrectly. Mobile apps often store cryptographic keys insecurely or derive weak keys from user passwords.

## Vulnerable Pattern

```kotlin
// BAD — Android: AES in ECB mode with hardcoded key
val cipher = Cipher.getInstance("AES/ECB/PKCS5Padding")  // ECB leaks patterns!
val keySpec = SecretKeySpec("hardcoded_key_1".toByteArray(), "AES")  // fixed key!
cipher.init(Cipher.ENCRYPT_MODE, keySpec)
val encrypted = cipher.doFinal(plaintext)
```

```swift
// BAD — iOS: MD5 for data integrity
import CommonCrypto
func md5(data: Data) -> String {
    var digest = [UInt8](repeating: 0, count: Int(CC_MD5_DIGEST_LENGTH))
    data.withUnsafeBytes { CC_MD5($0.baseAddress, CC_LONG(data.count), &digest) }
    return digest.map { String(format: "%02x", $0) }.joined()
    // MD5 is collision-broken — not suitable for integrity checks
}
```

```kotlin
// BAD — weak key derivation from PIN
val pin = "1234"
val key = pin.toByteArray().copyOf(32)  // just pads PIN to 32 bytes — catastrophically weak
```

## Secure Pattern

```kotlin
// GOOD — Android: AES-GCM with Android Keystore
val keyGenerator = KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES, "AndroidKeyStore")
keyGenerator.init(
    KeyGenParameterSpec.Builder("my_key_alias",
        KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT)
        .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
        .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
        .setKeySize(256)
        .build()
)
val secretKey = keyGenerator.generateKey()
val cipher = Cipher.getInstance("AES/GCM/NoPadding")
cipher.init(Cipher.ENCRYPT_MODE, secretKey)
// IV is auto-generated and embedded: cipher.iv
```

```swift
// GOOD — iOS: SHA-256 for integrity, CryptoKit for encryption
import CryptoKit
let hash = SHA256.hash(data: data)  // use SHA-256 or SHA-512
let symmetricKey = SymmetricKey(size: .bits256)
let sealedBox = try AES.GCM.seal(plaintext, using: symmetricKey)  // authenticated encryption
```

## Checks to Generate

- Grep for `AES/ECB/` cipher transformation — ECB mode is insecure, use GCM.
- Grep for `Cipher.getInstance("DES`, `Cipher.getInstance("RC4` — deprecated algorithms.
- Grep for `CC_MD5`, `MD5(`, `SHA1(` in security contexts in iOS code.
- Flag hardcoded IVs: `IvParameterSpec("fixed_iv_123".toByteArray())`.
- Grep for `SecretKeySpec(string.toByteArray()` — key derived from short string.
- Flag key derivation from PIN without PBKDF2/Argon2 with salt and high iteration count.
- Grep for `SecureRandom` seeded with `setSeed(System.currentTimeMillis())` — predictable seed.
