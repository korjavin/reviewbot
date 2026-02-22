---
id: owasp-mobile-m03-insecure-authentication
title: OWASP Mobile M03:2024 — Insecure Authentication/Authorization
severity: high
tags: [owasp-mobile-top10, authentication, biometrics, token-storage, mobile]
taxonomy: security/mobile/authentication
references:
  - https://owasp.org/www-project-mobile-top-10/2023-risks/m3-insecure-authentication-authorization.html
---

# OWASP Mobile M03:2024 — Insecure Authentication/Authorization

## Description

Mobile apps often implement weak authentication: relying solely on client-side checks, storing auth tokens insecurely, skipping re-authentication for sensitive actions, or implementing biometrics incorrectly (checking result locally without server validation).

## Vulnerable Pattern

```swift
// BAD — iOS: client-side only auth check (easily bypassed with Frida/debugger)
func isPremiumUser() -> Bool {
    return UserDefaults.standard.bool(forKey: "isPremium")  // attacker sets to true
}

// BAD — biometric auth without server validation
func authenticateWithBiometric(completion: @escaping (Bool) -> Void) {
    let context = LAContext()
    context.evaluatePolicy(.deviceOwnerAuthenticationWithBiometrics, ...) { success, _ in
        completion(success)  // success=true means app unlocks — no server check
    }
}
```

```kotlin
// BAD — Android: storing JWT in SharedPreferences (not encrypted)
val prefs = getSharedPreferences("auth", MODE_PRIVATE)
prefs.edit().putString("jwt_token", token).apply()  // accessible on rooted device
```

## Secure Pattern

```swift
// GOOD — iOS: token stored in Keychain with biometric protection
import Security

func storeToken(_ token: String) {
    let query: [String: Any] = [
        kSecClass as String: kSecClassGenericPassword,
        kSecAttrAccessible as String: kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
        kSecAttrAccessControl as String: SecAccessControlCreateWithFlags(
            nil, kSecAttrAccessibleWhenUnlockedThisDeviceOnly,
            .biometryCurrentSet, nil)!,
        kSecValueData as String: token.data(using: .utf8)!
    ]
    SecItemAdd(query as CFDictionary, nil)
}
// Server validates token on every request — biometric only unlocks keychain access
```

## Checks to Generate

- Flag `UserDefaults` / `SharedPreferences` storing `token`, `jwt`, `auth`, `session` keys.
- Grep for client-side premium/admin/role checks stored in local storage — server must authorize.
- Flag biometric authentication that only checks local result without sending signed challenge to server.
- Grep for `MODE_PRIVATE` SharedPreferences — use `EncryptedSharedPreferences` for sensitive data.
- Flag authentication tokens with no expiry or refresh mechanism.
- Check for missing certificate pinning in network calls (allows MITM interception of tokens).
