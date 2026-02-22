---
id: supply-chain-third-party-scripts
title: Third-Party Script Risks (CDN, Analytics, Tag Managers)
severity: high
tags: [supply-chain, third-party, cdn, analytics, sri, csp]
taxonomy: security/supply-chain/third-party-scripts
references:
  - https://owasp.org/www-project-top-ten/2021/A06_2021-Vulnerable_and_Outdated_Components/
  - https://developer.mozilla.org/en-US/docs/Web/Security/Subresource_Integrity
---

# Third-Party Script Risks (CDN, Analytics, Tag Managers)

## Description

Third-party scripts loaded from CDNs (analytics, ads, tag managers, chat widgets) execute with full access to the page DOM, including form inputs, cookies, and payment fields. Supply chain attacks on CDN providers (Polyfill.io in 2024, Magecart) have compromised thousands of sites simultaneously via single script modifications.

## Vulnerable Pattern

```html
<!-- BAD — third-party scripts without integrity check -->
<script src="https://cdn.example.com/analytics.js"></script>
<!-- If CDN is compromised, attacker's code runs on your site with full DOM access -->

<!-- BAD — Google Tag Manager loads arbitrary scripts from external dashboard -->
<script>(function(w,d,s,l,i){...})(window,document,'script','dataLayer','GTM-XXXX');</script>
<!-- Tag manager can inject any script — security review bypassed -->

<!-- BAD — CDN-hosted payment scripts without version pinning -->
<script src="https://js.stripe.com/v3/"></script>  <!-- version may change -->
```

## Secure Pattern

```html
<!-- GOOD — Subresource Integrity (SRI) for external scripts -->
<script
  src="https://cdn.example.com/library@2.3.1/dist/lib.min.js"
  integrity="sha384-abc123xyzdef456..."
  crossorigin="anonymous">
</script>
<!-- Browser verifies hash — tampered script refused -->

<!-- GOOD — self-host critical scripts (payment, auth) -->
<script src="/static/vendor/stripe-v3.js"></script>  <!-- own hash control -->
```

```python
# GOOD — generate SRI hash for third-party scripts in CI
import hashlib, base64, requests

def generate_sri_hash(url: str) -> str:
    content = requests.get(url).content
    digest = hashlib.sha384(content).digest()
    return "sha384-" + base64.b64encode(digest).decode()
```

```
# GOOD — CSP restricts which external hosts can load scripts
Content-Security-Policy: script-src 'self' https://js.stripe.com;
# Only allow stripe's known domain, nothing else
```

## Checks to Generate

- Grep HTML templates for `<script src="https://` without `integrity=` attribute (SRI).
- Flag Google Tag Manager or similar script loaders — all scripts loaded via GTM bypass CSP and SRI.
- Grep for CDN-hosted JavaScript without version pinning in URL.
- Check CSP `script-src` for wildcards that allow arbitrary CDN subdomains (`*.cdnjs.com`).
- Flag payment form pages loading third-party scripts that have DOM access to card input fields.
- Grep for `crossorigin` attribute missing on SRI-protected scripts (required for cross-origin SRI).
