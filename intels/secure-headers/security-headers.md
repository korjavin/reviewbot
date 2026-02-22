---
id: headers-security-headers
title: Missing HTTP Security Headers
severity: medium
tags: [headers, security-headers, browser-security, xss, clickjacking, mime-sniffing]
taxonomy: security/headers/http-security
references:
  - https://owasp.org/www-project-secure-headers/
  - https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html
  - https://securityheaders.com/
---

# Missing HTTP Security Headers

## Description

HTTP security headers instruct browsers to enforce security policies. Missing headers enable clickjacking, MIME-type sniffing attacks, cross-origin data leakage, and XSS. A comprehensive header policy is a defense-in-depth layer that reduces attack surface.

## Vulnerable Pattern

```python
# BAD — no security headers on responses
@app.get("/")
def homepage():
    return render_template("index.html")
    # Missing: X-Frame-Options, X-Content-Type-Options, HSTS, Referrer-Policy, Permissions-Policy
```

```nginx
# BAD — nginx with no security headers
server {
    listen 443 ssl;
    location / {
        proxy_pass http://app:8080;
        # No add_header directives for security
    }
}
```

## Secure Pattern

```python
# GOOD — FastAPI middleware adding all security headers
from fastapi import FastAPI, Request

@app.middleware("http")
async def add_security_headers(request: Request, call_next):
    response = await call_next(request)
    response.headers["X-Content-Type-Options"] = "nosniff"
    response.headers["X-Frame-Options"] = "DENY"
    response.headers["X-XSS-Protection"] = "0"  # disable legacy XSS filter (causes issues)
    response.headers["Referrer-Policy"] = "strict-origin-when-cross-origin"
    response.headers["Permissions-Policy"] = "camera=(), microphone=(), geolocation=()"
    response.headers["Strict-Transport-Security"] = "max-age=63072000; includeSubDomains; preload"
    response.headers["Content-Security-Policy"] = (
        "default-src 'self'; "
        "script-src 'self'; "
        "style-src 'self' 'unsafe-inline'; "
        "img-src 'self' data: https:; "
        "frame-ancestors 'none';"
    )
    return response
```

```nginx
# GOOD — nginx security header block
add_header X-Content-Type-Options "nosniff" always;
add_header X-Frame-Options "DENY" always;
add_header Referrer-Policy "strict-origin-when-cross-origin" always;
add_header Permissions-Policy "camera=(), microphone=(), geolocation=()";
add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
add_header Content-Security-Policy "default-src 'self'; frame-ancestors 'none';" always;
```

## Checks to Generate

- Check HTTP responses for missing `X-Content-Type-Options: nosniff`.
- Check for missing `X-Frame-Options: DENY` or `SAMEORIGIN` (clickjacking protection).
- Flag missing `Strict-Transport-Security` header on HTTPS responses.
- Flag missing `Referrer-Policy` — default `no-referrer-when-downgrade` leaks URLs.
- Check for missing or permissive `Content-Security-Policy` header.
- Flag `X-Powered-By` or `Server` response headers revealing tech stack — should be removed.
- Check for `X-XSS-Protection: 1; mode=block` — recommend setting to `0` (legacy filter causes issues).
