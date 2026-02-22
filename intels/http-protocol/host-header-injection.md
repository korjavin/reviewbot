---
id: http-host-header-injection
title: Host Header Injection
severity: high
tags: [http, host-header, password-reset, cache-poisoning, ssrf, open-redirect]
taxonomy: security/http/host-header
references:
  - https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/07-Input_Validation_Testing/17-Testing_for_HTTP_Host_Header_Injections
  - https://portswigger.net/web-security/host-header
---

# Host Header Injection

## Description

Applications that trust the HTTP `Host` header without validation are vulnerable to Host Header Injection. Attackers supply a malicious `Host` value to:
1. **Poison password reset links** — reset email contains attacker-controlled URL
2. **Web cache poisoning** — poison CDN/proxy cache with malicious Host-dependent content
3. **SSRF** — backend URL constructed from Host header reaches internal services
4. **Open redirect** — redirect URL built from Host header

Common vectors: `Host: evil.com`, `X-Forwarded-Host: evil.com`, `X-Host: evil.com`

## Vulnerable Pattern

```python
# BAD — password reset URL built from Host header
from flask import request

@app.post("/forgot-password")
def forgot_password(email: str):
    user = get_user(email)
    token = generate_reset_token(user)
    # Host header controlled by attacker!
    reset_url = f"https://{request.host}/reset?token={token}"
    send_email(user.email, f"Reset your password: {reset_url}")
    # Attacker sends: Host: attacker.com
    # Victim clicks link → attacker.com/reset?token=VALID_TOKEN → token stolen
```

```python
# BAD — URL generated from X-Forwarded-Host (set by attacker)
host = request.headers.get("X-Forwarded-Host", request.host)
base_url = f"https://{host}"

# BAD — Django: sites framework using Host header without validation
# settings.py
ALLOWED_HOSTS = []  # empty = no validation (debug mode)
# Or:
ALLOWED_HOSTS = ["*"]  # allows any host
```

```javascript
// BAD — Express: constructing URLs from req.hostname
app.post("/reset-password", (req, res) => {
    const resetUrl = `${req.protocol}://${req.hostname}/reset?token=${token}`;
    sendResetEmail(email, resetUrl);  // req.hostname comes from Host header
});
```

## Secure Pattern

```python
# GOOD — use hardcoded base URL from config, never from request
from flask import current_app

@app.post("/forgot-password")
def forgot_password(email: str):
    user = get_user(email)
    token = generate_reset_token(user)
    # BASE_URL is set by server config, not request headers
    reset_url = f"{current_app.config['BASE_URL']}/reset?token={token}"
    send_email(user.email, f"Reset your password: {reset_url}")

# GOOD — Django: strict ALLOWED_HOSTS validation
# settings.py
ALLOWED_HOSTS = ["myapp.com", "www.myapp.com"]
USE_X_FORWARDED_HOST = False  # don't trust X-Forwarded-Host
SECURE_PROXY_SSL_HEADER = ("HTTP_X_FORWARDED_PROTO", "https")  # trust only specific headers
```

```javascript
// GOOD — hardcoded base URL from environment
const BASE_URL = process.env.BASE_URL;  // "https://myapp.com"

app.post("/reset-password", (req, res) => {
    const resetUrl = `${BASE_URL}/reset?token=${token}`;  // never from req
    sendResetEmail(email, resetUrl);
});
```

## Checks to Generate

- Grep for `request.host`, `req.hostname`, `request.headers["host"]` used in URL construction for emails, redirects, or links.
- Grep for `request.headers.get("X-Forwarded-Host"`, `request.headers.get("X-Host"` used in URL generation.
- Flag Django `ALLOWED_HOSTS = []` or `ALLOWED_HOSTS = ["*"]` — host validation disabled.
- Flag Django `USE_X_FORWARDED_HOST = True` — enables X-Forwarded-Host which attackers control.
- Grep for `${req.protocol}://${req.hostname}` or `f"https://{request.host}"` patterns.
- Grep for password reset, email verification, and magic link functions using request-derived base URL.
