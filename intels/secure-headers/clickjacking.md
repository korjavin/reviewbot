---
id: headers-clickjacking
title: Clickjacking (UI Redressing) Attack
severity: medium
tags: [headers, clickjacking, iframe, x-frame-options, csp]
taxonomy: security/headers/clickjacking
references:
  - https://owasp.org/www-community/attacks/Clickjacking
  - https://cheatsheetseries.owasp.org/cheatsheets/Clickjacking_Defense_Cheat_Sheet.html
---

# Clickjacking (UI Redressing) Attack

## Description

Clickjacking tricks users into clicking on invisible, overlaid content by embedding the target site in an iframe. Attackers overlay buttons (transfer funds, change email, authorize OAuth) beneath innocent-looking UI. Protection requires browser-level frame restriction via `X-Frame-Options` or CSP `frame-ancestors`.

## Vulnerable Pattern

```html
<!-- Attacker's page: overlays victim's bank transfer button -->
<style>
    #target { opacity: 0; position: absolute; z-index: 2; }
    #decoy  { position: absolute; z-index: 1; }
</style>
<iframe id="target" src="https://bank.example.com/transfer?to=attacker&amount=5000"></iframe>
<div id="decoy">Click here to win a prize!</div>
```

```python
# BAD — application missing X-Frame-Options header
# Any site can embed the app in an iframe
@app.get("/transfer")
def transfer_funds():
    return render_template("transfer.html")  # no frame protection
```

## Secure Pattern

```python
# GOOD — deny framing entirely (most secure for sensitive pages)
@app.middleware("http")
async def frame_protection(request: Request, call_next):
    response = await call_next(request)
    response.headers["X-Frame-Options"] = "DENY"
    # OR: allow same origin only
    # response.headers["X-Frame-Options"] = "SAMEORIGIN"
    return response

# GOOD — CSP frame-ancestors (supersedes X-Frame-Options)
response.headers["Content-Security-Policy"] = "frame-ancestors 'none';"
# OR for same-origin embeds:
# response.headers["Content-Security-Policy"] = "frame-ancestors 'self';"
```

```javascript
// LEGACY — JavaScript frame-busting (easily defeated, not recommended as sole defense)
if (window !== window.top) {
    window.top.location = window.location;  // attacker can disable with sandbox attribute
}
// Use X-Frame-Options or CSP instead
```

## Checks to Generate

- Check all HTML responses for missing `X-Frame-Options` header.
- Verify CSP includes `frame-ancestors` directive (overrides X-Frame-Options in modern browsers).
- Flag JavaScript-only frame-busting without HTTP header protection.
- Flag `X-Frame-Options: ALLOW-FROM uri` — deprecated, use CSP `frame-ancestors` instead.
- Check sensitive action pages (payment, settings, account operations) specifically for frame protection.
