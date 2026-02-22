---
id: crypto-insecure-random
title: Insecure Randomness for Security Purposes
severity: high
tags: [cryptography, randomness, prng, token-generation, secrets]
taxonomy: security/cryptography/randomness
references:
  - https://owasp.org/www-community/vulnerabilities/Insecure_Randomness
  - https://cheatsheetseries.owasp.org/cheatsheets/Cryptography_Cheat_Sheet.html
---

# Insecure Randomness for Security Purposes

## Description

Pseudo-random number generators (PRNG) like Python's `random`, Java's `java.util.Random`, and `Math.random()` in JavaScript are designed for statistical use — not security. Their output is predictable if the seed is known. Attackers who observe PRNG output can predict future values, forging session tokens, CSRF tokens, or password reset links.

## Vulnerable Pattern

```python
# BAD — Python random module for security tokens
import random, string

def generate_session_token() -> str:
    chars = string.ascii_letters + string.digits
    return "".join(random.choice(chars) for _ in range(32))  # predictable!

def generate_reset_code() -> int:
    return random.randint(100000, 999999)  # only 900000 possibilities

# BAD — seeded with predictable value
random.seed(42)
token = random.getrandbits(128)
```

```javascript
// BAD — Math.random() for security
const token = Math.random().toString(36).slice(2);  // not cryptographically random
const otp = Math.floor(Math.random() * 1000000);    // predictable
```

## Secure Pattern

```python
# GOOD — secrets module (CSPRNG)
import secrets

def generate_session_token() -> str:
    return secrets.token_urlsafe(32)  # 256 bits, URL-safe, cryptographically random

def generate_otp() -> str:
    return str(secrets.randbelow(1000000)).zfill(6)

def generate_api_key() -> str:
    return secrets.token_hex(32)  # 64 hex chars = 256 bits
```

```javascript
// GOOD — Web Crypto API (browser) or crypto module (Node)
const array = new Uint8Array(32);
crypto.getRandomValues(array);
const token = Buffer.from(array).toString("hex");

// Node.js
const token = require("crypto").randomBytes(32).toString("hex");
```

## Checks to Generate

- Grep for `random.choice(`, `random.randint(`, `random.random()`, `random.getrandbits(` in security contexts (token, session, OTP, CSRF, reset).
- Grep for `Math.random()` used for generating tokens, codes, or IDs.
- Flag `uuid.uuid4()` seeded from non-random source — while UUIDs are random by default, some UUID libraries use weak RNG.
- Grep for `random.seed(` with hardcoded or predictable values.
- Flag token generation functions that produce < 128 bits of entropy.
