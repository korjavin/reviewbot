---
id: owasp-mobile-m01-improper-credential-usage
title: OWASP Mobile M01:2024 — Improper Credential Usage
severity: critical
tags: [owasp-mobile-top10, credentials, hardcoded-secrets, api-keys, mobile]
taxonomy: security/mobile/credentials
references:
  - https://owasp.org/www-project-mobile-top-10/2023-risks/m1-improper-credential-usage.html
---

# OWASP Mobile M01:2024 — Improper Credential Usage

## Description

Mobile apps frequently contain hardcoded API keys, credentials, and secrets in source code, resource files, or build artifacts. Attackers decompile the app (APK/IPA) to extract these credentials. Additionally, credentials stored insecurely (SharedPreferences, UserDefaults, local SQLite without encryption) are accessible on rooted/jailbroken devices.

## Vulnerable Pattern

```java
// BAD — Android: hardcoded API key in source
public class ApiClient {
    private static final String API_KEY = "sk-prod-abc123xyz789";  // decompilable!
    private static final String DB_PASSWORD = "SuperSecret123!";
}
```

```javascript
// BAD — React Native: credentials in JS bundle (shipped to device)
const config = {
    apiKey: "AIzaSyD-abc123",        // Firebase key
    stripeSecretKey: "sk_live_abc",  // CRITICAL: secret key in client!
    adminPassword: "admin123",
};
```

```xml
<!-- BAD — Android strings.xml (included in APK) -->
<string name="aws_access_key">AKIAIOSFODNN7EXAMPLE</string>
<string name="aws_secret_key">wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY</string>
```

## Secure Pattern

```java
// GOOD — Android: retrieve secrets from secure server at runtime, never hardcode
public class ApiClient {
    private String apiKey;

    public void initialize(Context context) {
        // Fetch from backend with device attestation
        apiKey = fetchApiKeyFromServer(getDeviceAttestation(context));
    }
}

// GOOD — Store user credentials in Android Keystore
KeyStore keyStore = KeyStore.getInstance("AndroidKeyStore");
keyStore.load(null);
// Use KeyGenerator with AndroidKeyStoreProvider
```

## Checks to Generate

- Grep source for patterns: `API_KEY = "`, `SECRET = "`, `PASSWORD = "`, `TOKEN = "` as string literals.
- Scan `strings.xml`, `*.plist`, `*.xcconfig`, `google-services.json` for embedded credentials.
- Grep JS/TS bundles for `sk_live_`, `AKIA`, `ghp_`, `glpat-` (known secret prefixes).
- Flag `SharedPreferences` storing tokens in plaintext — should use `EncryptedSharedPreferences`.
- Check for secrets in `.env` files committed to source control (`.env` not in `.gitignore`).
- Scan decompiled APK `classes.dex` strings for credential patterns.
