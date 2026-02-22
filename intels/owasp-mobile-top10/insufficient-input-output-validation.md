---
id: owasp-mobile-m04-insufficient-input-output-validation
title: OWASP Mobile M04:2024 — Insufficient Input/Output Validation
severity: high
tags: [owasp-mobile-top10, input-validation, webview, xss, mobile]
taxonomy: security/mobile/validation
references:
  - https://owasp.org/www-project-mobile-top-10/2023-risks/m4-insufficient-input-output-validation.html
---

# OWASP Mobile M04:2024 — Insufficient Input/Output Validation

## Description

Mobile apps receiving data from servers, other apps (deep links, IPC), or user input without proper validation are vulnerable to injection, XSS via WebView, path traversal, and intent-based attacks. Particularly dangerous when WebView executes JavaScript from untrusted sources.

## Vulnerable Pattern

```kotlin
// BAD — Android: WebView with JavaScript enabled loading external URL
val webView = WebView(context)
webView.settings.javaScriptEnabled = true
webView.settings.allowFileAccessFromFileURLs = true  // dangerous
val url = intent.getStringExtra("url")  // from deep link — attacker controlled!
webView.loadUrl(url)  // loads attacker's page with JS — accesses local files

// BAD — deep link parameter used in SQL without validation
val userId = intent.data?.getQueryParameter("user_id")
db.rawQuery("SELECT * FROM users WHERE id = $userId", null)  // SQL injection via deep link
```

```swift
// BAD — iOS: WKWebView loading unvalidated URL from URL scheme
func application(_ app: UIApplication, open url: URL, ...) -> Bool {
    let target = url.queryParameters["redirect"]!
    webView.load(URLRequest(url: URL(string: target)!))  // open redirect / JS injection
    return true
}
```

## Secure Pattern

```kotlin
// GOOD — validate deep link URLs against allowlist
val allowedHosts = setOf("help.example.com", "support.example.com")
val url = intent.getStringExtra("url")
val parsed = Uri.parse(url)
if (parsed.scheme == "https" && parsed.host in allowedHosts) {
    webView.loadUrl(url)
} else {
    showError("Invalid URL")
}

// GOOD — WebView with minimal permissions
webView.settings.javaScriptEnabled = false  // only enable if truly needed
webView.settings.allowFileAccessFromFileURLs = false
webView.settings.allowUniversalAccessFromFileURLs = false
```

## Checks to Generate

- Flag `WebView` with `javaScriptEnabled = true` loading URLs from intents/deep links.
- Flag `allowFileAccessFromFileURLs = true` or `allowUniversalAccessFromFileURLs = true`.
- Grep for `intent.data?.getQueryParameter(` used in SQL queries without parameterization.
- Flag URL scheme handlers that redirect to user-supplied URLs without validation.
- Grep for `addJavascriptInterface` — exposes Java methods to JavaScript, potential RCE pre-API 17.
- Check deep link handlers for path traversal: `../` in filename/path parameters.
