---
id: crypto-key-management
title: Insecure Cryptographic Key Management
severity: high
tags: [cryptography, key-management, secrets-manager, rotation, kms]
taxonomy: security/cryptography/key-management
references:
  - https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html
  - https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html
---

# Insecure Cryptographic Key Management

## Description

Cryptographic keys are only as secure as how they're managed. Common failures: storing encryption keys alongside encrypted data, never rotating keys, using the same key for all operations, storing keys in environment variables without access controls, and using weak key derivation from passwords.

## Vulnerable Pattern

```python
# BAD — encryption key stored next to encrypted data in same DB
class EncryptedDocument(Base):
    id = Column(Integer, primary_key=True)
    encryption_key = Column(String)  # key stored next to ciphertext!
    encrypted_content = Column(LargeBinary)

# BAD — key derived from predictable password with no salt
import hashlib
key = hashlib.sha256("hardcoded_password".encode()).digest()  # fixed key, no rotation

# BAD — single key for all purposes (signing + encryption)
KEY = os.environ["APP_SECRET"]
jwt_signed = jwt.encode(payload, KEY)  # same key used for JWT signing
data_encrypted = AES_GCM(KEY).encrypt(sensitive_data)  # and data encryption
```

## Secure Pattern

```python
# GOOD — AWS KMS for key management
import boto3
from cryptography.hazmat.primitives.ciphers.aead import AESGCM

kms = boto3.client("kms", region_name="us-east-1")
KEY_ID = os.environ["KMS_KEY_ID"]

def encrypt(plaintext: bytes) -> bytes:
    # Generate data key — key never leaves KMS in plaintext on disk
    response = kms.generate_data_key(KeyId=KEY_ID, KeySpec="AES_256")
    plaintext_key = response["Plaintext"]   # use in memory only
    encrypted_key = response["CiphertextBlob"]  # store this

    nonce = os.urandom(12)
    ciphertext = AESGCM(plaintext_key).encrypt(nonce, plaintext, None)

    return encrypted_key + nonce + ciphertext  # store encrypted_key + nonce + ciphertext

# GOOD — separate keys per purpose
JWT_SIGNING_KEY = os.environ["JWT_SIGNING_KEY"]
DATA_ENCRYPTION_KEY = os.environ["DATA_ENCRYPTION_KEY"]
HMAC_VERIFICATION_KEY = os.environ["HMAC_VERIFICATION_KEY"]
```

## Checks to Generate

- Flag encryption keys stored in the same database table as encrypted data.
- Grep for static key derivation: `hashlib.sha256("literal_string".encode())` as encryption key.
- Flag single `APP_SECRET` or `SECRET_KEY` used for multiple cryptographic purposes.
- Check for key rotation mechanism — alert if no rotation schedule/automation exists.
- Grep for keys stored in `os.environ` without secrets manager backing (HashiCorp Vault, AWS Secrets Manager, GCP Secret Manager).
- Flag missing key versioning — encrypted data should include key version identifier for rotation.
