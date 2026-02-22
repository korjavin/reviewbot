---
id: auth-missing-mfa
title: Missing or Bypassable Multi-Factor Authentication (MFA)
severity: high
tags: [authentication, mfa, totp, 2fa, account-takeover]
taxonomy: security/authentication/mfa
references:
  - https://owasp.org/www-project-top-ten/2021/A07_2021-Identification_and_Authentication_Failures/
  - https://cheatsheetseries.owasp.org/cheatsheets/Multifactor_Authentication_Cheat_Sheet.html
---

# Missing or Bypassable Multi-Factor Authentication (MFA)

## Description

MFA is the single most effective control against credential-based attacks. Missing MFA on admin accounts and sensitive operations leaves them vulnerable to credential stuffing, phishing, and password reuse attacks. Poorly implemented MFA (OTP not time-limited, codes reusable, MFA skippable via API) provides false security.

## Vulnerable Pattern

```python
# BAD — MFA check only in frontend, not enforced server-side
@app.post("/api/admin/action")
def admin_action(user=Depends(get_current_user)):
    # Frontend shows MFA prompt but API doesn't enforce it
    perform_admin_action()  # attacker calls API directly, bypassing MFA UI

# BAD — OTP code not time-limited, reusable
def verify_otp(user_id: int, code: str) -> bool:
    stored_otp = db.query(OTP).filter(OTP.user_id == user_id, OTP.code == code).first()
    return stored_otp is not None  # no expiry check, code reusable indefinitely!

# BAD — SMS OTP without rate limiting (SIM swap attack vector)
@app.post("/send-otp")
def send_otp(phone: str):
    otp = str(random.randint(1000, 9999))  # weak 4-digit OTP, predictable RNG
    sms.send(phone, f"Your code: {otp}")
```

## Secure Pattern

```python
import pyotp
from datetime import datetime, timedelta

# GOOD — TOTP (time-based OTP) using RFC 6238
def verify_totp(user: User, code: str) -> bool:
    totp = pyotp.TOTP(user.totp_secret)
    # valid_window=1 allows ±30s clock drift
    return totp.verify(code, valid_window=1)

def setup_totp(user: User) -> str:
    secret = pyotp.random_base32()
    user.totp_secret = secret
    db.commit()
    provisioning_uri = pyotp.totp.TOTP(secret).provisioning_uri(
        user.email, issuer_name="MyApp"
    )
    return provisioning_uri  # encode as QR code for user

# GOOD — enforce MFA server-side on sensitive operations
@app.post("/api/admin/action")
def admin_action(
    totp_code: str,
    user=Depends(get_current_user),
    mfa=Depends(require_mfa)  # always verify MFA in middleware
):
    if not verify_totp(user, totp_code):
        raise HTTPException(401, "Invalid MFA code")
    perform_admin_action()
```

## Checks to Generate

- Flag admin/privileged endpoints lacking MFA verification in server-side middleware.
- Grep for OTP verification without expiry check (`expires_at > datetime.utcnow()`).
- Flag OTP codes not invalidated after successful use (replay attack).
- Grep for 4 or 6-digit OTP codes generated with `random.randint` — must use `secrets`.
- Check for absence of MFA enrollment flow for admin users.
- Flag SMS-only OTP without backup methods — SMS is vulnerable to SIM swap.
- Check for MFA bypass mechanisms (backup codes, support override) without audit logging.
