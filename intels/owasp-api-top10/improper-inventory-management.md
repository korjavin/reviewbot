---
id: owasp-api-a09-improper-inventory-management
title: OWASP API A09:2023 — Improper Inventory Management
severity: medium
tags: [owasp-api-top10, api-versioning, shadow-api, inventory]
taxonomy: security/api/inventory
references:
  - https://owasp.org/API-Security/editions/2023/en/0xa9-improper-inventory-management/
---

# OWASP API A09:2023 — Improper Inventory Management

## Description

Organizations often have multiple API versions, unintentionally exposed debug endpoints, or forgotten microservices — "shadow APIs" that lack current security controls. Attackers probe for old API versions (`/api/v1/`, `/api/v2/`) which may have been patched in current versions but remain accessible.

## Vulnerable Pattern

```python
# BAD — old API version still active, lacks auth added in v2
@router.get("/api/v1/users")          # old, forgotten, no auth
def list_users_v1():
    return db.query(User).all()

@router.get("/api/v2/users")          # current, has auth
def list_users_v2(user=Depends(get_current_user)):
    return db.query(User).all()

# BAD — debug endpoint left accessible in production
@router.get("/api/debug/config")
def debug_config():
    return {"db_url": settings.DATABASE_URL, "secret": settings.SECRET_KEY}
```

## Secure Pattern

```python
# GOOD — explicitly deprecate and remove old API versions
# Or redirect v1 to v2 and enforce same auth
@router.get("/api/v1/users", deprecated=True)
def list_users_v1(user=Depends(get_current_user)):
    return RedirectResponse("/api/v2/users")

# GOOD — never expose debug endpoints in production
if settings.ENVIRONMENT != "production":
    @router.get("/api/debug/config")
    def debug_config():
        return {"db_url": "***", "env": settings.ENVIRONMENT}
```

```yaml
# GOOD — API inventory checklist in CI
# Document all active API versions and their auth requirements
# Run automated discovery scan to detect unregistered endpoints
- name: API inventory check
  run: |
    python scripts/check_api_inventory.py --spec openapi.yaml --server $BASE_URL
```

## Checks to Generate

- Grep for multiple `/api/v1/`, `/api/v2/`, `/api/v3/` route prefixes — verify consistent auth across all versions.
- Flag routes with `/debug/`, `/test/`, `/dev/`, `/internal/` in path that may be active in production.
- Check for undocumented endpoints not present in OpenAPI spec but registered in router.
- Flag routes registered outside the main application router (often missed in security reviews).
- Check for deprecated API version endpoints still returning 200 status (should be decommissioned).
