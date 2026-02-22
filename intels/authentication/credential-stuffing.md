---
id: auth-credential-stuffing
title: Credential Stuffing and Brute Force Protection
severity: high
tags: [authentication, credential-stuffing, brute-force, rate-limiting, account-lockout]
taxonomy: security/authentication/credential-stuffing
references:
  - https://owasp.org/www-community/attacks/Credential_stuffing
  - https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
---

# Credential Stuffing and Brute Force Protection

## Description

Credential stuffing uses leaked username/password databases from other breaches to test credentials against your application. Brute force tries many password combinations. Both rely on automated, high-volume login attempts. Without rate limiting, account lockout, or bot detection, attackers can compromise accounts at scale.

The "Have I Been Pwned" database contains 12+ billion breached credentials — widely used in credential stuffing attacks.

## Vulnerable Pattern

```python
# BAD — no rate limiting, no lockout, returns too much info
@app.post("/login")
def login(username: str, password: str):
    user = db.query(User).filter(User.username == username).first()
    if not user:
        return {"error": "User not found"}  # reveals whether username exists
    if not verify_password(password, user.password_hash):
        return {"error": "Wrong password"}  # confirms username is valid!
    return create_session(user)
    # Bot can try 10,000 passwords/second with no consequences
```

## Secure Pattern

```python
from slowapi import Limiter
from slowapi.util import get_remote_address
import time

limiter = Limiter(key_func=get_remote_address)

# GOOD — rate limiting + account lockout + generic error
@app.post("/login")
@limiter.limit("5/minute;20/hour")  # per-IP rate limit
def login(request: Request, credentials: LoginRequest, db=Depends(get_db)):
    # Constant-time lookup (prevent timing attacks for username enumeration)
    user = db.query(User).filter(User.username == credentials.username).first()

    # Check account lockout
    if user and user.failed_attempts >= 10:
        lockout_until = user.last_failed_at + timedelta(minutes=15)
        if datetime.utcnow() < lockout_until:
            raise HTTPException(429, "Account temporarily locked")

    # Verify credentials — use dummy hash if user not found (prevent timing enumeration)
    password_valid = verify_password(
        credentials.password,
        user.password_hash if user else DUMMY_HASH
    )

    if not user or not password_valid:
        if user:
            user.failed_attempts += 1
            user.last_failed_at = datetime.utcnow()
            db.commit()
        time.sleep(0.5)  # slow down automated attacks
        raise HTTPException(401, "Invalid credentials")  # generic message

    # Success: reset lockout counter
    user.failed_attempts = 0
    db.commit()
    return create_session(user)
```

## Checks to Generate

- Flag login endpoints missing rate-limiting decorator or middleware.
- Flag different error messages for invalid username vs invalid password — enables enumeration.
- Grep for `user.failed_attempts` tracking — verify lockout threshold exists.
- Flag absence of CAPTCHA or device fingerprinting on login after N failures.
- Check for missing `time.sleep` or constant-time comparison in auth flows — timing attacks.
- Flag no integration with HIBP (Have I Been Pwned) API for password breach checking at registration.
