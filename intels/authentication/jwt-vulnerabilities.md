---
id: auth-jwt-vulnerabilities
title: JWT — Algorithm Confusion, None Algorithm, and Weak Secrets
severity: critical
tags: [authentication, jwt, algorithm-confusion, token, authorization]
taxonomy: security/authentication/jwt
references:
  - https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/06-Session_Management_Testing/10-Testing_JSON_Web_Tokens
  - https://portswigger.net/web-security/jwt
---

# JWT — Algorithm Confusion, None Algorithm, and Weak Secrets

## Description

JWTs have several well-known attack vectors:
1. **Algorithm None**: `alg: none` skips signature verification entirely
2. **Algorithm Confusion**: Switching RS256 to HS256 causes the server to verify with the public key as HMAC secret — easily obtained
3. **Weak Secret**: Brute-forceable HMAC secret (`secret`, `password`, `123456`)
4. **Missing Claims Validation**: `exp`, `iss`, `aud` not validated

## Vulnerable Pattern

```python
# BAD — PyJWT: accepts any algorithm including none
import jwt
payload = jwt.decode(token, key, algorithms=None)  # BAD: all algorithms accepted

# BAD — algorithm confusion vector
payload = jwt.decode(token, public_key, algorithms=["RS256", "HS256"])
# Attacker uses public key (known) as HMAC secret with HS256

# BAD — no claims validation
payload = jwt.decode(token, key, algorithms=["HS256"],
                     options={"verify_exp": False, "verify_aud": False})
```

```javascript
// BAD — jsonwebtoken: algorithm not specified on verification
jwt.verify(token, secret)  // accepts whatever algorithm is in header
```

## Secure Pattern

```python
# GOOD — strict algorithm allowlist, required claims
import jwt

SECRET = os.environ["JWT_SECRET"]  # strong random secret (32+ bytes)

def create_token(user_id: int) -> str:
    return jwt.encode({
        "sub": str(user_id),
        "iat": datetime.utcnow(),
        "exp": datetime.utcnow() + timedelta(minutes=15),
        "iss": "myapp",
        "aud": "myapp-api",
    }, SECRET, algorithm="HS256")

def verify_token(token: str) -> dict:
    return jwt.decode(
        token,
        SECRET,
        algorithms=["HS256"],  # single algorithm, explicit
        options={"require": ["exp", "iat", "sub", "iss", "aud"]},
        issuer="myapp",
        audience="myapp-api",
    )
```

## Checks to Generate

- Grep for `jwt.decode(` without explicit `algorithms=[...]` list or with `algorithms=None`.
- Grep for `algorithms=["RS256", "HS256"]` — confusion vulnerability.
- Flag `verify_exp: False`, `verify_signature: False`, `verify_aud: False` in JWT options.
- Grep for JWT secrets shorter than 32 characters or matching common values (`secret`, `key`, `jwt`).
- Check for missing `exp` claim in JWT creation — tokens that never expire.
- Flag JWT tokens stored in `localStorage` (accessible to XSS) vs. `httpOnly` cookies.
