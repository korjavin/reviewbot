---
id: crypto-weak-cipher-algorithms
title: Weak and Broken Cryptographic Algorithms
severity: high
tags: [cryptography, cipher, md5, sha1, des, rc4, algorithm]
taxonomy: security/cryptography/algorithms
references:
  - https://owasp.org/www-community/vulnerabilities/Insecure_Randomness
  - https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html
  - https://nvlpubs.nist.gov/nistpubs/SpecialPublications/NIST.SP.800-131Ar2.pdf
---

# Weak and Broken Cryptographic Algorithms

## Description

Using deprecated or broken cryptographic algorithms provides false security. MD5 and SHA-1 are collision-broken. DES and 3DES have insufficient key lengths. RC4 has known biases. AES-ECB reveals plaintext patterns. These should be replaced with modern equivalents.

NIST SP 800-131A disallows MD5, SHA-1, DES, 2TDEA, and RC4 for federal use — industry standard follows.

## Vulnerable Pattern

```python
# BAD — MD5 for any security purpose
import hashlib
password_hash = hashlib.md5(password.encode()).hexdigest()  # broken for passwords
file_integrity = hashlib.md5(file_data).hexdigest()  # collision attacks possible

# BAD — DES encryption
from Crypto.Cipher import DES
cipher = DES.new(key[:8], DES.MODE_ECB)  # 56-bit key, broken since 1999

# BAD — RC4 stream cipher
from Crypto.Cipher import ARC4
cipher = ARC4.new(key)  # statistical biases, known attacks

# BAD — AES-ECB mode (deterministic, leaks patterns)
from Crypto.Cipher import AES
cipher = AES.new(key, AES.MODE_ECB)
```

## Secure Pattern

```python
# GOOD — bcrypt/argon2 for passwords
import bcrypt
hash = bcrypt.hashpw(password.encode(), bcrypt.gensalt(rounds=12))

# GOOD — SHA-256 or SHA-3 for integrity/fingerprinting
import hashlib
file_integrity = hashlib.sha256(file_data).hexdigest()

# GOOD — AES-GCM (authenticated encryption) for data
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
key = AESGCM.generate_key(bit_length=256)
aesgcm = AESGCM(key)
nonce = os.urandom(12)
ciphertext = aesgcm.encrypt(nonce, plaintext, aad)

# GOOD — X25519 / ECDH for key exchange; Ed25519 for signing
from cryptography.hazmat.primitives.asymmetric.x25519 import X25519PrivateKey
private_key = X25519PrivateKey.generate()
```

## Checks to Generate

- Grep for `hashlib.md5(`, `hashlib.sha1(` in security contexts (passwords, HMAC, signatures).
- Grep for `AES.MODE_ECB`, `DES.new(`, `ARC4.new(`, `Blowfish.new(`.
- Flag RSA key sizes below 2048 bits: `RSA.generate(1024`, `RSA.generate(512`.
- Grep for `ssl.PROTOCOL_TLSv1` or `ssl.PROTOCOL_SSLv3` — deprecated TLS versions.
- Grep for `MD5withRSA`, `SHA1withRSA` in Java signature algorithms.
- Flag `hmac.new(key, msg, hashlib.md5)` — MD5-based HMAC is weakened.
