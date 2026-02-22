---
id: crypto-missing-encryption-at-rest
title: Missing Encryption at Rest for Sensitive Data
severity: high
tags: [cryptography, encryption, data-at-rest, pii, gdpr, database]
taxonomy: security/cryptography/encryption-at-rest
references:
  - https://owasp.org/www-project-top-ten/2021/A02_2021-Cryptographic_Failures/
  - https://cheatsheetseries.owasp.org/cheatsheets/Cryptographic_Storage_Cheat_Sheet.html
---

# Missing Encryption at Rest for Sensitive Data

## Description

Sensitive data (PII, health records, financial data, credentials) stored in plaintext databases, files, or backups is exposed if those storage systems are compromised. GDPR, HIPAA, PCI DSS, and SOC2 require encryption at rest for regulated data. Application-level encryption provides an extra layer beyond disk encryption.

## Vulnerable Pattern

```python
# BAD — PII stored in plaintext database columns
class User(Base):
    id = Column(Integer, primary_key=True)
    email = Column(String)               # searchable — ok if hashed for lookup
    social_security_number = Column(String)  # plaintext SSN in DB!
    credit_card_number = Column(String)      # plaintext PAN — PCI DSS violation!
    date_of_birth = Column(Date)             # plaintext PII
    diagnosis = Column(String)               # plaintext health data — HIPAA violation!
```

```yaml
# BAD — database backups without encryption
backup:
  schedule: "0 2 * * *"
  destination: s3://my-backups/
  # No encryption — backup files accessible to anyone with S3 access
```

## Secure Pattern

```python
from cryptography.fernet import Fernet
from sqlalchemy_utils import EncryptedType
from sqlalchemy_utils.types.encrypted.encrypted_type import AesGcmEngine

# Application-level encryption for sensitive fields
class User(Base):
    id = Column(Integer, primary_key=True)
    email = Column(String)  # hashed for lookup, plaintext for UX if low sensitivity

    # GOOD — application-level encryption with KMS-backed key
    ssn = Column(EncryptedType(String, os.environ["FIELD_ENCRYPTION_KEY"],
                               AesGcmEngine, "pkcs5"))
    # SSN stored as encrypted blob — key managed separately from data
```

```yaml
# GOOD — encrypted database backups
backup:
  schedule: "0 2 * * *"
  destination: s3://my-backups/
  encryption:
    enabled: true
    kms_key_id: "arn:aws:kms:us-east-1:..."
    algorithm: "AES256"
```

## Checks to Generate

- Grep database models for columns named `ssn`, `social_security`, `credit_card`, `card_number`, `cvv`, `diagnosis`, `health_` without `EncryptedType` or equivalent.
- Flag database backup configurations without encryption.
- Grep for `Column(String)` / `Column(Text)` on HIPAA/PCI regulated fields.
- Check for database disk encryption enabled (AWS RDS `storage_encrypted = true`, GCP `disk_encryption`).
- Flag application-level file storage of sensitive documents without encryption.
- Check for separation of encryption keys from encrypted data (keys not stored in same DB).
