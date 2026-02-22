---
id: auth-weak-password-policy
title: Weak Password Policy
severity: medium
tags: [authentication, password-policy, password-strength, hibp]
taxonomy: security/authentication/password-policy
references:
  - https://owasp.org/www-project-top-ten/2021/A07_2021-Identification_and_Authentication_Failures/
  - https://cheatsheetseries.owasp.org/cheatsheets/Authentication_Cheat_Sheet.html
  - https://pages.nist.gov/800-63-3/sp800-63b.html
---

# Weak Password Policy

## Description

NIST SP 800-63B revised best practices for passwords: focus on minimum length (8+ chars), check against breached password lists, allow all character types, do not impose complexity rules that drive predictable patterns, and eliminate periodic forced rotation (it causes "Password1!" → "Password2!"). Many applications still enforce outdated, counterproductive policies.

## Vulnerable Pattern

```python
# BAD — overly restrictive complexity rules (NIST deprecated)
def validate_password(password: str) -> bool:
    if len(password) < 8:
        return False
    if not any(c.isupper() for c in password):  # forces predictable patterns
        return False
    if not any(c.isdigit() for c in password):
        return False
    if not any(c in "!@#$%^&*()" for c in password):
        return False
    return True
    # Users create: Password1! → easy to guess, hard to remember

# BAD — allowing common/breached passwords
def register(username: str, password: str):
    if len(password) < 6:  # too short, no breach check
        raise HTTPException(400, "Password too short")
    create_user(username, hash_password(password))
    # "password123", "letmein", "qwerty" all accepted
```

## Secure Pattern

```python
import hashlib
import requests

def check_pwned_password(password: str) -> int:
    """Returns count of times password appears in HIBP breach database."""
    sha1 = hashlib.sha1(password.encode()).hexdigest().upper()
    prefix, suffix = sha1[:5], sha1[5:]
    response = requests.get(f"https://api.pwnedpasswords.com/range/{prefix}", timeout=5)
    for line in response.text.splitlines():
        hash_suffix, count = line.split(":")
        if hash_suffix == suffix:
            return int(count)
    return 0

COMMON_PASSWORDS = {"password", "123456", "qwerty", "letmein", "admin", "welcome"}

def validate_password(password: str, username: str) -> None:
    if len(password) < 12:  # NIST recommends 8+ but 12+ is better practice
        raise ValueError("Password must be at least 12 characters")
    if password.lower() in COMMON_PASSWORDS:
        raise ValueError("Password is too common")
    if username.lower() in password.lower():
        raise ValueError("Password must not contain username")
    breaches = check_pwned_password(password)
    if breaches > 0:
        raise ValueError(f"Password found in {breaches:,} data breaches — choose a different one")
    # No arbitrary complexity rules — length + breach check is sufficient
```

## Checks to Generate

- Grep for `len(password) < 6` or `< 8` — minimum should be 12 characters.
- Flag absence of HIBP (Have I Been Pwned) or equivalent breach-check integration.
- Grep for `COMMON_PASSWORDS` allowlist or similar check — must reject top 1000 passwords.
- Check for forced periodic rotation (`change_password_required_at`) — NIST deprecated, causes weak patterns.
- Flag password stored in plaintext for email confirmation flows.
- Grep for maximum password length limits below 64 characters — indicates plaintext storage or weak hashing.
