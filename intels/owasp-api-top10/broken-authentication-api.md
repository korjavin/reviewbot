---
id: owasp-api-a02-broken-authentication
title: OWASP API A02:2023 — Broken Authentication
severity: critical
tags: [owasp-api-top10, authentication, jwt, api-keys, token]
taxonomy: security/api/authentication
references:
  - https://owasp.org/API-Security/editions/2023/en/0xa2-broken-authentication/
  - https://cheatsheetseries.owasp.org/cheatsheets/JSON_Web_Token_for_Java_Cheat_Sheet.html
---

# OWASP API A02:2023 — Broken Authentication

## Description

Poorly implemented API authentication mechanisms allow attackers to compromise authentication tokens, impersonate users, and perform unauthorized actions. Common issues include: weak API key generation, JWT algorithm confusion (`alg: none`), missing token expiry, tokens in URLs, and no refresh token rotation.

## Vulnerable Pattern

```python
# BAD — JWT alg:none accepted (authentication bypass)
import jwt
def verify_token(token: str):
    return jwt.decode(token, options={"verify_signature": False})  # accepts alg:none!

# BAD — API key stored in URL / logs
GET /api/data?api_key=secret123  # appears in server logs, browser history, referer headers

# BAD — predictable token generation
import random
def generate_api_key():
    return str(random.randint(100000, 999999))  # 6 digits — easily brute-forced
```

```javascript
// BAD — JWT secret is weak / default
const token = jwt.sign({ userId: id }, "secret", { expiresIn: "never" });
```

## Secure Pattern

```python
# GOOD — JWT with algorithm allowlist, expiry, and audience validation
from jose import jwt, JWTError

SECRET_KEY = secrets.token_hex(32)
ALGORITHM = "HS256"
ACCESS_TOKEN_EXPIRE_MINUTES = 15

def verify_token(token: str) -> dict:
    try:
        payload = jwt.decode(
            token,
            SECRET_KEY,
            algorithms=[ALGORITHM],  # explicit allowlist — no alg:none
            options={"require": ["exp", "iat", "sub"]}
        )
        return payload
    except JWTError:
        raise HTTPException(401, "Invalid token")

# GOOD — API key in Authorization header, generated securely
import secrets
def generate_api_key() -> str:
    return secrets.token_urlsafe(32)  # 256 bits of randomness

# Authorization: Bearer <api_key>
```

## Checks to Generate

- Grep for `verify_signature": False` or `algorithms=["none"]` in JWT decode calls.
- Flag JWT tokens with `expiresIn: "never"` or missing expiry (`exp` claim not set).
- Grep for API keys or tokens in URL query parameters (`?token=`, `?api_key=`, `?key=`).
- Flag `random.randint` or `Math.random()` used for security token generation — must use `secrets` / `crypto.randomBytes`.
- Grep for hardcoded JWT secrets: `jwt.sign(..., "secret"`, `jwt.encode(..., "password"`.
- Check for missing token revocation mechanism on logout.
