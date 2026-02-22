---
id: owasp-mobile-m07-insufficient-binary-protections
title: OWASP Mobile M07:2024 — Insufficient Binary Protections
severity: medium
tags: [owasp-mobile-top10, reverse-engineering, obfuscation, tampering, mobile]
taxonomy: security/mobile/binary
references:
  - https://owasp.org/www-project-mobile-top-10/2023-risks/m7-insufficient-binary-protections.html
---

# OWASP Mobile M07:2024 — Insufficient Binary Protections

## Description

Mobile binaries without obfuscation, anti-tampering, or anti-debugging protections are easily reverse-engineered. Attackers can extract business logic, API keys, bypass license checks, or create modified versions. Particularly important for fintech, gaming (anti-cheat), and high-value apps.

## Vulnerable Pattern

```java
// BAD — Android: license/premium check in plaintext Java (decompilable)
public boolean isPremium() {
    return BuildConfig.BUILD_TYPE.equals("premium") || purchasedSkus.contains("premium_access");
    // Frida hook: Java.use("com.app.LicenseChecker").isPremium.implementation = () => true;
}

// BAD — no root detection (allows Frida/Magisk-based bypass)
public void startApp() {
    // No check for rooted device or debugger attached
    initializeApp();
}
```

```gradle
// BAD — no code obfuscation (ProGuard/R8 disabled)
buildTypes {
    release {
        minifyEnabled false   // class names fully readable in decompiler
        shrinkResources false
    }
}
```

## Secure Pattern

```gradle
// GOOD — enable R8 with custom rules
buildTypes {
    release {
        minifyEnabled true
        shrinkResources true
        proguardFiles getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro"
    }
}
```

```kotlin
// GOOD — runtime integrity checks (defense-in-depth)
fun performSecurityChecks(context: Context): Boolean {
    if (isEmulator()) return false
    if (isDebuggerConnected()) return false
    if (isAppTampered(context)) return false  // check APK signature
    return true
}

fun isAppTampered(context: Context): Boolean {
    val signatures = context.packageManager
        .getPackageInfo(context.packageName, PackageManager.GET_SIGNATURES).signatures
    val currentHash = hashSignature(signatures[0])
    return currentHash != EXPECTED_SIGNATURE_HASH
}
```

## Checks to Generate

- Check Gradle build files for `minifyEnabled false` in release builds.
- Grep for license/premium checks in plaintext without native code or remote validation.
- Flag absence of SafetyNet/Play Integrity API attestation for high-value operations.
- Check for `android:debuggable="true"` in production AndroidManifest.xml.
- Flag iOS apps without bitcode stripped and without symbol stripping in release.
- Check for missing `get-task-allow` entitlement removal in iOS production builds.
