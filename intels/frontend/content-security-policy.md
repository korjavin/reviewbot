---
id: frontend-content-security-policy
title: Missing or Weak Content Security Policy (CSP)
severity: medium
tags: [csp, frontend, xss, browser-security, headers]
taxonomy: security/frontend/csp
references:
  - https://owasp.org/www-community/controls/Content_Security_Policy
  - https://cheatsheetseries.owasp.org/cheatsheets/Content_Security_Policy_Cheat_Sheet.html
  - https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP
---

# Missing or Weak Content Security Policy (CSP)

## Description

Content Security Policy is a browser security mechanism that restricts sources of scripts, styles, images, and other resources. A strong CSP significantly mitigates XSS impact by preventing inline scripts and restricting external script loading. Missing or overly permissive CSP (`unsafe-inline`, `unsafe-eval`, wildcards) negates this protection.

## Vulnerable Pattern

```python
# BAD — no CSP header at all
@app.get("/")
def homepage():
    return render_template("index.html")  # no security headers

# BAD — CSP with unsafe-inline and unsafe-eval (common but dangerous)
response.headers["Content-Security-Policy"] = (
    "default-src 'self' 'unsafe-inline' 'unsafe-eval' *"
)
# unsafe-inline allows <script>evil()</script>; unsafe-eval allows eval()
```

```html
<!-- BAD — inline event handlers (require unsafe-inline CSP) -->
<button onclick="login()">Login</button>
<a href="javascript:void(0)" onclick="doSomething()">Click</a>
```

## Secure Pattern

```python
# GOOD — strict CSP with nonce-based script allowance
import secrets

@app.middleware("http")
async def add_security_headers(request: Request, call_next):
    response = await call_next(request)
    nonce = secrets.token_urlsafe(16)
    response.headers["Content-Security-Policy"] = (
        f"default-src 'self'; "
        f"script-src 'self' 'nonce-{nonce}'; "  # nonce instead of unsafe-inline
        f"style-src 'self' 'unsafe-inline'; "   # inline styles often needed
        f"img-src 'self' data: https:; "
        f"font-src 'self' https://fonts.gstatic.com; "
        f"connect-src 'self' https://api.example.com; "
        f"frame-ancestors 'none'; "             # clickjacking protection
        f"base-uri 'self'; "
        f"form-action 'self';"
    )
    return response
```

## Checks to Generate

- Flag missing `Content-Security-Policy` header on HTML responses.
- Flag CSP containing `'unsafe-inline'` in `script-src` — allows inline XSS.
- Flag CSP containing `'unsafe-eval'` in `script-src` — allows `eval()`.
- Flag CSP with wildcard `*` in `script-src` or `default-src`.
- Grep for inline event handlers (`onclick=`, `onload=`, `onerror=`) — incompatible with strict CSP.
- Check for `Content-Security-Policy-Report-Only` header in production — enforce, don't just report.
- Flag `frame-ancestors` missing — clickjacking vulnerability.
