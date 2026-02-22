---
id: owasp-mobile-m02-inadequate-supply-chain
title: OWASP Mobile M02:2024 — Inadequate Supply Chain Security
severity: high
tags: [owasp-mobile-top10, supply-chain, third-party-sdk, mobile, dependencies]
taxonomy: security/mobile/supply-chain
references:
  - https://owasp.org/www-project-mobile-top-10/2023-risks/m2-inadequate-supply-chain-security.html
---

# OWASP Mobile M02:2024 — Inadequate Supply Chain Security

## Description

Mobile apps integrate dozens of third-party SDKs (analytics, ads, social login, crash reporting) that run with the same privileges as the host app. Malicious or compromised SDKs can exfiltrate user data, inject ads, track users, or execute arbitrary code. SDK updates can introduce backdoors without the app developer's knowledge.

## Vulnerable Pattern

```gradle
// BAD — no version pinning for third-party SDKs
dependencies {
    implementation "com.facebook.android:facebook-android-sdk:+"  // latest version always
    implementation "com.google.firebase:firebase-analytics"        // no version lock
    // SDK update may introduce data collection or malicious code
}
```

```kotlin
// BAD — third-party analytics SDK with extensive permissions
// SDK initialized with access to:
// - full contact list
// - device identifiers
// - clipboard content
// - keyboard input (some SDKs)
Analytics.init(context, API_KEY, Analytics.Config().enableAllFeatures())
```

```ruby
# iOS — Podfile without version constraint
pod 'AnalyticsSDK'           # pulls latest — breaking changes + security risks
pod 'AdNetwork', '~> 3.0'    # allows minor updates that may contain changes
```

## Secure Pattern

```gradle
// GOOD — pinned versions + integrity verification
dependencies {
    implementation "com.example.sdk:analytics:2.3.1"  // exact version
}
```

```kotlin
// GOOD — initialize analytics with minimal data collection
Analytics.init(
    context, API_KEY,
    Analytics.Config()
        .disableAdvertisingId()
        .disableDeviceMetrics()
        .setAnonymousMode(true)
)
```

```yaml
# GOOD — CI pipeline scans mobile dependencies
- name: Scan Android dependencies
  run: ./gradlew dependencyCheckAnalyze --info
  # Or: use MobSF for SDK analysis
```

## Checks to Generate

- Grep `build.gradle` / `Podfile` for `+` or no version pinning in SDK dependencies.
- Flag analytics/advertising SDKs initialized without minimal data collection configuration.
- Check for SDKs requesting permissions beyond the host app's declared permissions.
- Grep for `enable*` flags on analytics SDKs that enable extensive tracking.
- Flag absence of SDK software composition analysis (SCA) in CI pipeline.
- Check `Podfile.lock` and `build.gradle` lock files committed to version control.
