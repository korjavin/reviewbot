---
id: infra-open-redirect
title: Open Redirect
severity: medium
tags: [infrastructure, open-redirect, phishing, url-validation]
taxonomy: security/web/open-redirect
references:
  - https://owasp.org/www-community/attacks/Unvalidated_Redirects_and_Forwards_Cheat_Sheet
  - https://cheatsheetseries.owasp.org/cheatsheets/Unvalidated_Redirects_and_Forwards_Cheat_Sheet.html
---

# Open Redirect

## Description

Open redirects allow attackers to craft trusted-looking URLs that redirect users to malicious sites. Used in phishing ("Login at myapp.com, then get redirected to evil.com"), bypassing URL filters, and as SSRF intermediaries. Particularly dangerous in OAuth flows where redirect URIs are validated by string matching.

## Vulnerable Pattern

```python
# BAD — redirect to user-supplied URL without validation
@app.get("/logout")
def logout(next: str = "/"):
    session.clear()
    return redirect(next)  # attacker: /logout?next=https://evil.com

# BAD — substring validation (easily bypassed)
@app.get("/go")
def go(url: str):
    if "myapp.com" in url:  # bypassed with: https://myapp.com.evil.com
        return redirect(url)
    raise HTTPException(400)

# BAD — relative path bypass via protocol-relative URL
# /go?url=//evil.com/phishing → redirects to https://evil.com/phishing
```

```javascript
// BAD — Express: redirect based on Referer header
app.get("/success", (req, res) => {
    const returnUrl = req.query.returnUrl || req.headers.referer;
    res.redirect(returnUrl);  // attacker controls returnUrl parameter
});
```

## Secure Pattern

```python
from urllib.parse import urlparse

ALLOWED_DOMAINS = {"myapp.com", "api.myapp.com"}

def is_safe_redirect(url: str) -> bool:
    if not url or url.startswith("//"):  # block protocol-relative
        return False
    if url.startswith("/"):
        return True  # relative path always safe
    parsed = urlparse(url)
    return parsed.hostname in ALLOWED_DOMAINS and parsed.scheme in ("http", "https")

@app.get("/logout")
def logout(next: str = "/"):
    session.clear()
    safe_next = next if is_safe_redirect(next) else "/"
    return redirect(safe_next)
```

## Checks to Generate

- Grep for `redirect(request.args.get(`, `redirect(request.query.get(` — user-controlled redirect target.
- Flag `res.redirect(req.query.`, `response.redirect(req.params.` in Node.js.
- Grep for redirect URL validation using `in` or `startswith` — must use exact hostname comparison.
- Flag protocol-relative URLs (`//`) in redirect logic — should be blocked.
- Check OAuth/OIDC `redirect_uri` validation for open redirect vulnerability.
- Grep for `window.location = ` assigned from URL parameters in JavaScript.
