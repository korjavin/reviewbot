---
id: owasp-mobile-m08-security-misconfiguration-mobile
title: OWASP Mobile M08:2024 — Security Misconfiguration (Mobile)
severity: medium
tags: [owasp-mobile-top10, misconfiguration, android, ios, mobile]
taxonomy: security/mobile/misconfiguration
references:
  - https://owasp.org/www-project-mobile-top-10/2023-risks/m8-security-misconfiguration.html
---

# OWASP Mobile M08:2024 — Security Misconfiguration (Mobile)

## Description

Mobile-specific misconfigurations include exported Android components (activities, services, receivers) accessible by other apps without permission, iOS entitlement over-provisioning, debug flags in production builds, and backup-enabled manifests exposing app data.

## Vulnerable Pattern

```xml
<!-- BAD — Android: Activity exported without permission — any app can start it -->
<activity
    android:name=".AdminActivity"
    android:exported="true">  <!-- accessible by any installed app! -->
</activity>

<!-- BAD — BroadcastReceiver exported — any app can send intents to it -->
<receiver
    android:name=".PaymentReceiver"
    android:exported="true">
    <intent-filter>
        <action android:name="com.app.PAYMENT_COMPLETE"/>
    </intent-filter>
</receiver>

<!-- BAD — backup enabled exposes DB/files to adb backup -->
<application android:allowBackup="true" android:debuggable="true">
```

```xml
<!-- BAD — iOS: over-provisioned entitlements -->
<key>com.apple.security.app-sandbox</key>
<false/>  <!-- no sandbox -->
<key>keychain-access-groups</key>
<array>
    <string>*</string>  <!-- access to ALL keychain groups -->
</array>
```

## Secure Pattern

```xml
<!-- GOOD — Android: restrict exported components -->
<activity
    android:name=".AdminActivity"
    android:exported="false">  <!-- internal only -->
</activity>

<receiver
    android:name=".PaymentReceiver"
    android:exported="true"
    android:permission="com.app.PAYMENT_PERMISSION">  <!-- requires permission to send -->
</receiver>

<application
    android:allowBackup="false"
    android:debuggable="false">
```

## Checks to Generate

- Scan `AndroidManifest.xml` for `android:exported="true"` on activities/services/receivers not intended to be public.
- Grep for `android:debuggable="true"` — must be false in release builds.
- Flag `android:allowBackup="true"` for apps handling sensitive data.
- Check iOS `*.entitlements` for `com.apple.security.app-sandbox: false`.
- Flag `get-task-allow: true` entitlement in iOS (enables debugger attachment) — must be false in production.
- Grep for `android:networkSecurityConfig` absence — default allows cleartext in older API levels.
