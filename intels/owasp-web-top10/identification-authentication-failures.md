---
id: owasp-web-a07-identification-authentication-failures
title: OWASP A07:2021 — Identification and Authentication Failures
severity: high
tags: [owasp-top10, authentication, session, mfa, brute-force]
taxonomy: security/web/authentication
references:
  - https://owasp.org/Top10/A07_2021-Identification_and_Authentication_Failures/
  - https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
  - https://cheatsheetseries.owasp.org/cheatsheets/Session_Management_Cheat_Sheet.html
---

# OWASP A07:2021 — Identification and Authentication Failures

## Description

Authentication failures allow attackers to compromise passwords, keys, or session tokens, or to exploit implementation flaws to assume other users' identities. Common issues: no brute-force protection, weak passwords permitted, insecure session IDs, missing multi-factor authentication.

## Vulnerable Pattern

```python
# BAD — no rate limiting or lockout on login
@app.post("/login")
def login(username: str, password: str):
    user = db.query(User).filter(User.username == username).first()
    if user and user.password == password:  # also: plaintext password comparison!
        session["user_id"] = user.id
        return {"token": create_token(user)}
    return {"error": "Invalid credentials"}

# BAD — session ID in URL (logged by proxies/servers)
# GET /dashboard?session_id=abc123
```

```python
# BAD — no password complexity requirements
@app.post("/register")
def register(username: str, password: str):
    if len(password) < 4:  # trivially weak policy
        raise HTTPException(400, "Password too short")
    create_user(username, hash(password))
```

## Secure Pattern

```python
from slowapi import Limiter
limiter = Limiter(key_func=get_remote_address)

# GOOD — rate limited login, constant-time comparison, secure session
@app.post("/login")
@limiter.limit("5/minute")
def login(request: Request, credentials: LoginRequest):
    user = db.query(User).filter(User.username == credentials.username).first()
    if not user or not bcrypt.checkpw(credentials.password.encode(), user.password_hash):
        raise HTTPException(401, "Invalid credentials")
    if not verify_totp(user, credentials.totp_code):  # MFA check
        raise HTTPException(401, "Invalid MFA code")
    token = create_signed_jwt(user.id, expires_in=3600)
    response = JSONResponse({"message": "OK"})
    response.set_cookie("session", token, httponly=True, secure=True, samesite="strict")
    return response
```

## Checks to Generate

- Flag login endpoints missing rate-limiting decorator or middleware.
- Grep for `user.password == password` — plaintext password comparison (should use bcrypt/argon2 check).
- Flag session tokens or auth tokens in URL parameters (`?session=`, `?token=`, `?auth=`).
- Flag cookies without `HttpOnly`, `Secure`, and `SameSite` flags.
- Grep for `len(password) <` with values below 8 — weak password policy.
- Check for absence of MFA / TOTP verification in authentication flow.
- Flag `session.clear()` missing on logout — incomplete session invalidation.
