---
id: http-response-splitting
title: HTTP Response Splitting (CRLF Injection)
severity: high
tags: [http, response-splitting, crlf-injection, header-injection, xss, cache-poisoning]
taxonomy: security/http/response-splitting
references:
  - https://owasp.org/www-community/attacks/HTTP_Response_Splitting
  - https://cheatsheetseries.owasp.org/cheatsheets/HTTP_Headers_Cheat_Sheet.html
---

# HTTP Response Splitting (CRLF Injection)

## Description

HTTP Response Splitting occurs when user-controlled data is embedded in HTTP response headers without stripping CRLF (`\r\n`) sequences. Attackers inject newlines to add arbitrary headers or inject a complete second HTTP response body, enabling: XSS, cache poisoning, session hijacking via Set-Cookie injection, and phishing via response body control.

The `\r\n` sequence terminates a header; `\r\n\r\n` ends headers and begins the body.

Payload: `value\r\nSet-Cookie: session=attacker_value` injects a rogue Set-Cookie header.

## Vulnerable Pattern

```python
# BAD — user input reflected into response header without CRLF stripping
from flask import redirect, request

@app.get("/redirect")
def do_redirect():
    next_url = request.args.get("next", "/")
    response = redirect(next_url)
    # BAD: setting header with user-controlled language param
    lang = request.args.get("lang", "en")
    response.headers["Content-Language"] = lang
    # Payload: lang=en\r\nSet-Cookie: admin=true
    return response
```

```python
# BAD — cookie value set from user input
@app.post("/set-preference")
def set_preference():
    theme = request.form.get("theme", "light")
    response = make_response("OK")
    response.set_cookie("theme", theme)
    # Payload: theme=light\r\nSet-Cookie: session=stolen_value
    # (raw set_cookie vulnerable to CRLF in some frameworks)
    return response
```

```javascript
// BAD — Express: setting header from user input
app.get("/track", (req, res) => {
    const ref = req.query.ref;
    res.setHeader("X-Referral", ref);  // CRLF injection possible
    res.send("Tracked");
});
```

## Secure Pattern

```python
# GOOD — strip CRLF from any value going into headers
import re

def sanitize_header_value(value: str) -> str:
    # Strip carriage return and newline characters
    return re.sub(r"[\r\n]", "", value)

@app.get("/redirect")
def do_redirect():
    next_url = request.args.get("next", "/")
    # Validate URL is safe (allowlist)
    if not is_safe_redirect(next_url):
        next_url = "/"
    lang = sanitize_header_value(request.args.get("lang", "en"))
    response = redirect(next_url)
    response.headers["Content-Language"] = lang
    return response
```

```javascript
// GOOD — Node.js: sanitize before setting header
function sanitizeHeaderValue(value) {
    return String(value).replace(/[\r\n]/g, "");
}

app.get("/track", (req, res) => {
    const ref = sanitizeHeaderValue(req.query.ref || "");
    res.setHeader("X-Referral", ref);
    res.send("Tracked");
});
```

## Checks to Generate

- Grep for `response.headers[key] = request.` — user input directly into response header value.
- Grep for `res.setHeader(`, `response.set(` with values derived from query params, body, or path params.
- Grep for `set_cookie(`, `response.set_cookie(` with cookie value from user input without sanitization.
- Flag redirect responses where `Location` header value comes from user input without CRLF stripping.
- Grep for logging headers written back into HTTP responses without newline stripping.
- Check for raw `\r\n` in user-controlled strings inserted into multipart responses.
