---
id: auth-insecure-password-reset
title: Insecure Password Reset Flows
severity: high
tags: [authentication, password-reset, account-takeover, token]
taxonomy: security/authentication/password-reset
references:
  - https://owasp.org/www-community/vulnerabilities/Forgot_Password_Cheat_Sheet
  - https://cheatsheetseries.owasp.org/cheatsheets/Forgot_Password_Cheat_Sheet.html
---

# Insecure Password Reset Flows

## Description

Password reset flows are a frequent target for account takeover. Common vulnerabilities: predictable reset tokens, tokens that don't expire, reusable tokens, user enumeration via different responses, reset tokens sent in the URL (logged), and secret-question-based reset.

## Vulnerable Pattern

```python
# BAD — predictable token (timestamp-based)
import time
def generate_reset_token(user_id: int) -> str:
    return hashlib.md5(f"{user_id}{int(time.time())}".encode()).hexdigest()

# BAD — reset token in URL query param (appears in server logs, referer headers)
reset_link = f"https://app.com/reset?token={token}&email={email}"

# BAD — token never expires
class PasswordReset(Base):
    token = Column(String)
    used = Column(Boolean, default=False)
    # no created_at, no expiry check

# BAD — different response for valid vs invalid email (enumeration)
if not user:
    return {"error": "No account with this email"}  # reveals account existence
```

## Secure Pattern

```python
import secrets
from datetime import datetime, timedelta

# GOOD — cryptographically random token with expiry
def generate_reset_token(user: User, db: Session) -> str:
    token = secrets.token_urlsafe(32)  # 256 bits
    reset = PasswordReset(
        user_id=user.id,
        token=hashlib.sha256(token.encode()).hexdigest(),  # store hash only
        expires_at=datetime.utcnow() + timedelta(minutes=15),  # short expiry
        used=False,
    )
    db.add(reset)
    db.commit()
    return token  # return plain token for email

def verify_reset_token(token: str, db: Session) -> PasswordReset | None:
    token_hash = hashlib.sha256(token.encode()).hexdigest()
    reset = db.query(PasswordReset).filter(
        PasswordReset.token == token_hash,
        PasswordReset.used == False,
        PasswordReset.expires_at > datetime.utcnow()
    ).first()
    return reset

# GOOD — use POST body for token (not URL)
@app.post("/reset-password")
def reset_password(token: str = Body(...), new_password: str = Body(...)):
    ...
```

## Checks to Generate

- Grep for `reset_token = ` using `random`, `time`, `hash(user_id)` — must use `secrets.token_urlsafe`.
- Flag reset links with token in URL query parameter — should use POST body or fragment.
- Grep for `PasswordReset` / `reset_tokens` table without `expires_at` column.
- Flag reset tokens not marked as used after successful reset (token reuse attack).
- Check for same response body for valid/invalid email in forgot-password endpoint.
- Flag secret question / security question based reset — deprecated, easily bypassed.
