---
id: access-control-api-key-management
title: Insecure API Key Management
severity: high
tags: [access-control, api-keys, authentication, secrets, rotation]
taxonomy: security/access-control/api-keys
references:
  - https://owasp.org/www-project-api-security/
  - https://cheatsheetseries.owasp.org/cheatsheets/REST_Security_Cheat_Sheet.html
---

# Insecure API Key Management

## Description

API keys are commonly used to authenticate programmatic access. Poor management leads to: leaked keys in source code or logs, keys with excessive permissions, non-expiring keys, keys shared across environments, and no rotation after suspected compromise. Unlike user credentials, API key compromise is often silent and long-lived.

## Vulnerable Pattern

```python
# BAD — API key stored in source code
API_KEY = "sk_live_<REDACTED>"  # Stripe secret key — example of hardcoded credential

# BAD — same API key for all clients / environments
# prod and dev/staging share the same key → dev breach affects prod

# BAD — API key logged in request middleware
@app.middleware("http")
async def log_requests(request: Request, call_next):
    logger.info(f"Request: {request.headers}")  # logs Authorization: Bearer <key>
    return await call_next(request)

# BAD — API key never expires, no rotation mechanism
class APIKey(Base):
    key = Column(String, unique=True)
    created_at = Column(DateTime)
    # No expires_at, no last_used_at, no is_active flag
```

## Secure Pattern

```python
import secrets, hashlib
from datetime import datetime, timedelta

# GOOD — generate cryptographically random API keys
def generate_api_key() -> tuple[str, str]:
    """Returns (plaintext_key, hashed_key). Store hash, give plaintext to user."""
    key = f"rbt_{secrets.token_urlsafe(32)}"  # prefix for easy detection (e.g., gitleaks)
    key_hash = hashlib.sha256(key.encode()).hexdigest()
    return key, key_hash

# GOOD — API key model with metadata for auditing and rotation
class APIKey(Base):
    id = Column(UUID, primary_key=True, default=uuid4)
    key_hash = Column(String, unique=True)  # never store plaintext
    name = Column(String)                   # human-readable label
    user_id = Column(Integer, ForeignKey("users.id"))
    scopes = Column(JSON)                   # principle of least privilege
    created_at = Column(DateTime, default=datetime.utcnow)
    expires_at = Column(DateTime)           # mandatory expiry
    last_used_at = Column(DateTime)         # detect stale keys
    is_active = Column(Boolean, default=True)

def verify_api_key(provided_key: str, db: Session) -> APIKey | None:
    key_hash = hashlib.sha256(provided_key.encode()).hexdigest()
    api_key = db.query(APIKey).filter(
        APIKey.key_hash == key_hash,
        APIKey.is_active == True,
        APIKey.expires_at > datetime.utcnow()
    ).first()
    if api_key:
        api_key.last_used_at = datetime.utcnow()  # track usage
        db.commit()
    return api_key
```

## Checks to Generate

- Grep for API key string literals in source: `sk_live_`, `api_key = "`, `API_KEY = "`.
- Flag API key models missing `expires_at` column — keys must have mandatory expiry.
- Flag API key storage without hashing — plaintext keys in DB means DB breach = key breach.
- Grep for API keys in request/response logging middleware.
- Check for missing per-key scope/permission restrictions — keys should not have broader access than needed.
- Flag API keys shared across environments (same key value in prod and staging configs).
- Check for API key rotation mechanism and alerting on stale keys (unused for 90+ days).
