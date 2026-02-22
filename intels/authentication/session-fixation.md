---
id: auth-session-fixation
title: Session Fixation and Session Hijacking
severity: high
tags: [authentication, session, fixation, hijacking, cookies]
taxonomy: security/authentication/session
references:
  - https://owasp.org/www-community/attacks/Session_fixation
  - https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
---

# Session Fixation and Session Hijacking

## Description

Session fixation allows an attacker to set a known session ID before the victim authenticates. After login, the server keeps the same session ID — which the attacker already knows. Session hijacking is stealing an existing session token via XSS, network interception, or log leakage.

## Vulnerable Pattern

```python
# BAD — session ID preserved across privilege level change (fixation)
@app.post("/login")
def login(username: str, password: str):
    user = authenticate(username, password)
    if not user:
        raise HTTPException(401)
    # BAD: existing session_id kept — attacker pre-set it
    session["user_id"] = user.id
    return {"message": "Logged in"}

# BAD — session ID in URL (hijackable via referrer/logs)
# GET /dashboard?JSESSIONID=abc123

# BAD — session cookie without Secure/HttpOnly
response.set_cookie("session", session_token)  # no flags set
```

## Secure Pattern

```python
# GOOD — regenerate session ID on authentication (prevents fixation)
from flask import session
import secrets

@app.post("/login")
def login(username: str, password: str):
    user = authenticate(username, password)
    if not user:
        raise HTTPException(401)
    # Regenerate session to prevent fixation
    session.clear()
    session["user_id"] = user.id
    session["session_id"] = secrets.token_hex(32)  # new ID

# GOOD — secure cookie attributes
response.set_cookie(
    "session",
    session_token,
    httponly=True,    # not accessible to JavaScript
    secure=True,      # HTTPS only
    samesite="Strict", # CSRF protection
    max_age=3600      # 1 hour expiry
)
```

## Checks to Generate

- Flag login handlers that set `session["user_id"]` without clearing/regenerating the session first.
- Grep for session tokens in URL parameters (`?session_id=`, `?JSESSIONID=`, `?sid=`).
- Flag `set_cookie(` calls missing `httponly=True`, `secure=True`, or `samesite=`.
- Check for session invalidation on logout (`session.clear()`, `session.invalidate()`).
- Flag session tokens with predictable values or insufficient entropy.
- Grep for session tokens stored in `localStorage` — accessible to XSS attacks.
