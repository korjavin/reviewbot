---
id: frontend-csrf
title: Cross-Site Request Forgery (CSRF)
severity: high
tags: [csrf, frontend, authentication, session, browser-security]
taxonomy: security/frontend/csrf
references:
  - https://owasp.org/www-community/attacks/csrf
  - https://cheatsheetseries.owasp.org/cheatsheets/Cross-Site_Request_Forgery_Prevention_Cheat_Sheet.html
---

# Cross-Site Request Forgery (CSRF)

## Description

CSRF tricks authenticated users into submitting unwanted requests. A malicious website causes the victim's browser to send authenticated requests (using session cookies) to the target application. Can result in fund transfers, password/email changes, account deletion, or any state-changing operation.

## Vulnerable Pattern

```html
<!-- Attacker's page: triggers fund transfer on victim's bank -->
<img src="https://bank.example.com/transfer?to=attacker&amount=10000" />
<!-- Browser automatically includes victim's session cookie -->

<!-- Form-based CSRF -->
<form action="https://app.example.com/change-email" method="POST">
    <input name="email" value="attacker@evil.com">
</form>
<script>document.forms[0].submit();</script>
```

```python
# BAD — state-changing endpoint with no CSRF protection, relies only on cookies
@app.post("/change-password")
def change_password(new_password: str, user=Depends(get_current_user)):
    user.password_hash = hash_password(new_password)
    db.commit()  # CSRF: any site can trigger this with victim's session cookie
```

## Secure Pattern

```python
# GOOD — Synchronizer Token Pattern
import secrets

@app.get("/settings")
def settings_page(user=Depends(get_current_user)):
    csrf_token = secrets.token_hex(32)
    session["csrf_token"] = csrf_token
    return render_template("settings.html", csrf_token=csrf_token)

@app.post("/change-password")
def change_password(
    new_password: str,
    csrf_token: str = Form(...),
    user=Depends(get_current_user)
):
    if csrf_token != session.get("csrf_token"):
        raise HTTPException(403, "Invalid CSRF token")
    user.password_hash = hash_password(new_password)
    db.commit()
```

```python
# GOOD — SameSite=Strict cookie attribute (defense-in-depth)
response.set_cookie("session", token, samesite="Strict", secure=True, httponly=True)
# Modern defense: SameSite=Lax prevents most CSRF, Strict prevents all cross-site requests
```

## Checks to Generate

- Flag state-changing endpoints (`POST`, `PUT`, `DELETE`, `PATCH`) without CSRF token validation.
- Check cookie settings — missing `SameSite=Strict` or `SameSite=Lax` on session cookies.
- Grep for CSRF token implementations using `GET` request (token should be in request body or header).
- Flag APIs using cookie-based auth without `SameSite` cookies and without CSRF tokens.
- Check for absent CSRF middleware: Django's `CsrfViewMiddleware`, Flask-WTF, etc.
- Flag `SameSite=None` cookies without CSRF token — cross-origin requests allowed.
