---
id: headers-subresource-integrity
title: Missing Subresource Integrity (SRI) on External Assets
severity: medium
tags: [headers, sri, cdn, supply-chain, browser-security, integrity]
taxonomy: security/headers/sri
references:
  - https://developer.mozilla.org/en-US/docs/Web/Security/Subresource_Integrity
  - https://owasp.org/www-project-secure-headers/
---

# Missing Subresource Integrity (SRI) on External Assets

## Description

Subresource Integrity (SRI) allows browsers to verify that externally loaded scripts and stylesheets haven't been tampered with. Without SRI, a compromised CDN can serve malicious JavaScript to every visitor. The Polyfill.io supply chain attack (2024) compromised 380,000+ websites because SRI wasn't used.

## Vulnerable Pattern

```html
<!-- BAD — CDN-hosted jQuery without SRI -->
<script src="https://code.jquery.com/jquery-3.7.1.min.js"></script>
<!-- BAD — Bootstrap CSS from CDN without integrity check -->
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css">
<!-- If CDN serves modified files, your users run malicious code -->
```

```html
<!-- BAD — polyfill.io (compromised in 2024) without SRI -->
<script src="https://polyfill.io/v3/polyfill.min.js"></script>
```

## Secure Pattern

```html
<!-- GOOD — SRI with integrity hash and CORS attribute -->
<script
  src="https://code.jquery.com/jquery-3.7.1.min.js"
  integrity="sha384-1H217gwSVyLSIfaLxHbE7dRb3v4mYCKbpQvzx0cegeju1MVsGrX5xXxAvs/HgeFs"
  crossorigin="anonymous">
</script>

<link
  rel="stylesheet"
  href="https://cdn.jsdelivr.net/npm/bootstrap@5.3.0/dist/css/bootstrap.min.css"
  integrity="sha384-gH2yIJqKdNHPEq0n4Mqa/HGKIhSkIHeL5AyhkYV8i59U5AR6csBvApHHNl/vI1Bx"
  crossorigin="anonymous">
```

```bash
# GOOD — generate SRI hash for any external resource
curl -s "https://code.jquery.com/jquery-3.7.1.min.js" | \
  openssl dgst -sha384 -binary | \
  openssl base64 -A
# Output prefix with: sha384-<output>
```

```html
<!-- ALTERNATIVE — self-host all external assets, remove CDN dependency -->
<script src="/static/vendor/jquery-3.7.1.min.js"></script>
```

## Checks to Generate

- Grep HTML templates for `<script src="https://` without `integrity=` attribute.
- Grep for `<link rel="stylesheet" href="https://` without `integrity=` attribute.
- Flag polyfill.io usage — service was compromised in 2024; self-host polyfills instead.
- Check that `crossorigin="anonymous"` is set alongside `integrity` attribute.
- Grep for CDN-hosted fonts (Google Fonts) without SRI — lower risk but still supply chain.
- Flag `integrity` hashes using SHA-256 — recommend SHA-384 or SHA-512 for higher security.
