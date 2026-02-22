---
id: owasp-mobile-m05-insecure-communication
title: OWASP Mobile M05:2024 — Insecure Communication
severity: high
tags: [owasp-mobile-top10, tls, certificate-pinning, mitm, mobile]
taxonomy: security/mobile/communication
references:
  - https://owasp.org/www-project-mobile-top-10/2023-risks/m5-insecure-communication.html
---

# OWASP Mobile M05:2024 — Insecure Communication

## Description

Mobile apps transmitting data over insecure channels or with improper TLS configuration expose users to man-in-the-middle (MITM) attacks. Common failures: using HTTP instead of HTTPS, accepting invalid/self-signed certificates, missing certificate pinning for sensitive apps, and logging network traffic in production.

## Vulnerable Pattern

```kotlin
// BAD — Android: trust all certificates (common "fix" for dev that ships to prod)
val trustAllCerts = arrayOf<TrustManager>(object : X509TrustManager {
    override fun checkClientTrusted(chain: Array<X509Certificate>, authType: String) {}
    override fun checkServerTrusted(chain: Array<X509Certificate>, authType: String) {}
    override fun getAcceptedIssuers(): Array<X509Certificate> = arrayOf()
})
val sslContext = SSLContext.getInstance("SSL")
sslContext.init(null, trustAllCerts, java.security.SecureRandom())
```

```swift
// BAD — iOS: disabling ATS (App Transport Security) in Info.plist
// Allows HTTP connections — all traffic unencrypted
<key>NSAppTransportSecurity</key>
<dict>
    <key>NSAllowsArbitraryLoads</key>
    <true/>
</dict>
```

```javascript
// BAD — React Native: cleartext HTTP in network request
fetch("http://api.example.com/users")  // no TLS
```

## Secure Pattern

```kotlin
// GOOD — Android: certificate pinning with OkHttp
val certificatePinner = CertificatePinner.Builder()
    .add("api.example.com", "sha256/AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=")
    .build()
val client = OkHttpClient.Builder()
    .certificatePinner(certificatePinner)
    .build()
```

```xml
<!-- GOOD — Android: network security config restricts cleartext -->
<!-- res/xml/network_security_config.xml -->
<network-security-config>
    <base-config cleartextTrafficPermitted="false">
        <trust-anchors>
            <certificates src="system"/>
        </trust-anchors>
    </base-config>
</network-security-config>
```

## Checks to Generate

- Grep for `TrustAllCerts`, `trustAllCerts`, `checkServerTrusted` returning nothing — MITM vulnerability.
- Flag `NSAllowsArbitraryLoads: true` in iOS `Info.plist`.
- Grep for `http://` (not `https://`) in API base URLs in mobile code.
- Flag `rejectUnauthorized: false` in Node.js HTTPS options.
- Check for missing `network_security_config.xml` in Android manifest.
- Flag production builds with Charles/mitmproxy trust certificates embedded.
