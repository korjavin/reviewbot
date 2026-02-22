---
id: headers-referrer-policy
title: Missing or Weak Referrer-Policy Header
severity: low
tags: [headers, referrer-policy, privacy, information-disclosure, url-leakage]
taxonomy: security/headers/referrer-policy
references:
  - https://owasp.org/www-project-secure-headers/
  - https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Referrer-Policy
  - https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html
---

# Missing or Weak Referrer-Policy Header

## Description

Without a Referrer-Policy header, browsers send the full page URL as the `Referer` header when navigating to external sites or loading external resources. URLs often contain sensitive information: session tokens (anti-pattern but common), user IDs, search terms, password reset tokens, and internal path structures. This leaks to third-party analytics, CDNs, and linked external sites.

## Vulnerable Pattern

```python
# BAD — no Referrer-Policy header
@app.get("/password-reset")
def password_reset_page():
    # URL: https://app.com/password-reset?token=abc123&email=user@example.com
    return render_template("reset.html")
    # User clicks external link on this page → Referer: https://app.com/password-reset?token=abc123
    # Token leaked to external site's server logs!

# BAD — page with sensitive data in URL contains external links
# https://app.com/admin/users?search=john.doe@company.com
# Clicking external ad → leaks search query to advertiser
```

## Secure Pattern

```python
# GOOD — strict referrer policy
@app.middleware("http")
async def add_security_headers(request: Request, call_next):
    response = await call_next(request)
    # Only send origin (no path/query) on cross-origin; full URL for same-origin
    response.headers["Referrer-Policy"] = "strict-origin-when-cross-origin"
    return response

# For highest privacy (no referrer sent at all to cross-origin):
# response.headers["Referrer-Policy"] = "no-referrer"
```

```html
<!-- GOOD — per-link referrer policy for sensitive pages -->
<a href="https://external.com" referrerpolicy="no-referrer">External Link</a>

<!-- GOOD — meta tag for page-level policy -->
<meta name="referrer" content="strict-origin-when-cross-origin">
```

## Checks to Generate

- Check HTTP responses for missing `Referrer-Policy` header on HTML pages.
- Flag `Referrer-Policy: unsafe-url` — sends full URL including path and query to all destinations.
- Flag `Referrer-Policy: no-referrer-when-downgrade` (browser default) — sends full URL on HTTPS→HTTPS.
- Grep for sensitive data in URL patterns: `?token=`, `?reset=`, `?email=`, `?key=` on pages with external links.
- Check for `Referrer-Policy` in nginx/Apache config vs application middleware — ensure coverage.
