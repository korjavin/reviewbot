---
id: owasp-web-a02-cryptographic-failures
title: OWASP A02:2021 — Cryptographic Failures
severity: high
tags: [owasp-top10, cryptography, tls, sensitive-data, encryption]
taxonomy: security/web/cryptography
references:
  - https://owasp.org/Top10/A02_2021-Cryptographic_Failures/
  - https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html
---

# OWASP A02:2021 — Cryptographic Failures

## Description

Cryptographic failures (formerly "Sensitive Data Exposure") occur when data is not adequately protected in transit or at rest. This includes weak ciphers, missing TLS, improper key management, and storing passwords without strong hashing. Attackers exploit these to steal credentials, PII, financial data, and session tokens.

## Vulnerable Pattern

```python
# BAD — MD5 for password hashing (broken, reversible via rainbow tables)
import hashlib
def store_password(password: str):
    return hashlib.md5(password.encode()).hexdigest()

# BAD — AES-ECB mode (deterministic, leaks patterns)
from Crypto.Cipher import AES
cipher = AES.new(key, AES.MODE_ECB)
ciphertext = cipher.encrypt(pad(plaintext))

# BAD — HTTP connection for sensitive form submission
fetch("/login", { method: "POST", body: JSON.stringify(creds) })
# served over http:// — credentials travel in cleartext
```

## Secure Pattern

```python
# GOOD — bcrypt with cost factor for passwords
import bcrypt
def store_password(password: str) -> bytes:
    return bcrypt.hashpw(password.encode(), bcrypt.gensalt(rounds=12))

def verify_password(password: str, hashed: bytes) -> bool:
    return bcrypt.checkpw(password.encode(), hashed)

# GOOD — AES-GCM (authenticated encryption)
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
key = AESGCM.generate_key(bit_length=256)
aesgcm = AESGCM(key)
nonce = os.urandom(12)
ciphertext = aesgcm.encrypt(nonce, plaintext, associated_data)
```

```nginx
# GOOD — enforce TLS, disable old protocols
server {
    listen 443 ssl;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_prefer_server_ciphers on;
    ssl_ciphers 'ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384';
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
}
```

## Checks to Generate

- Grep for `md5`, `sha1`, `des`, `rc4`, `AES.MODE_ECB` in crypto code — flag weak/broken algorithms.
- Grep for `hashlib.md5(password`, `hashlib.sha1(password` — passwords must use bcrypt/argon2/scrypt.
- Flag HTTP (`http://`) URLs used for authentication or sensitive data endpoints.
- Check TLS config for `SSLv3`, `TLSv1.0`, `TLSv1.1` — these are deprecated.
- Grep for hardcoded keys: `key = b"..."`, `SECRET = "..."` patterns near crypto operations.
- Flag missing `Strict-Transport-Security` header on HTTPS responses.
