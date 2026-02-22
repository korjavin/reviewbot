---
id: owasp-api-a05-broken-function-level-auth
title: OWASP API A05:2023 — Broken Function Level Authorization
severity: high
tags: [owasp-api-top10, authorization, rbac, admin, privilege-escalation]
taxonomy: security/api/authorization
references:
  - https://owasp.org/API-Security/editions/2023/en/0xa5-broken-function-level-authorization/
---

# OWASP API A05:2023 — Broken Function Level Authorization

## Description

BFLA occurs when APIs fail to properly restrict access to administrative or privileged functions. Unlike BOLA (object-level), BFLA is about what actions a user can perform, not which records they can see. Attackers discover admin endpoints through API documentation, client-side code, or fuzzing, and call them directly.

## Vulnerable Pattern

```python
# BAD — admin endpoint guessable, no role check
@router.delete("/api/v1/admin/users/{user_id}")
def admin_delete_user(user_id: int, current_user=Depends(get_current_user)):
    # relies on "security through obscurity" — no actual role check
    db.query(User).filter(User.id == user_id).delete()

# BAD — HTTP method confusion — GET vs POST vs DELETE not enforced
@router.api_route("/api/users/{id}", methods=["GET", "DELETE"])
def user_resource(id: int):
    if request.method == "DELETE":
        db.query(User).filter(User.id == id).delete()
    # no authorization difference between GET and DELETE
```

## Secure Pattern

```python
from functools import wraps

def require_role(*roles):
    def decorator(func):
        @wraps(func)
        async def wrapper(*args, current_user: User = Depends(get_current_user), **kwargs):
            if current_user.role not in roles:
                raise HTTPException(403, "Insufficient permissions")
            return await func(*args, current_user=current_user, **kwargs)
        return wrapper
    return decorator

# GOOD — explicit role guard on every privileged function
@router.delete("/api/v1/admin/users/{user_id}")
@require_role("admin", "super_admin")
def admin_delete_user(user_id: int, current_user=Depends(get_current_user)):
    db.query(User).filter(User.id == user_id).delete()

# GOOD — separate endpoints with separate auth requirements
@router.get("/api/users/{id}")
def get_user(id: int, current_user=Depends(get_current_user)):
    ...

@router.delete("/api/admin/users/{id}")
def delete_user(id: int, current_user=Depends(require_role("admin"))):
    ...
```

## Checks to Generate

- Flag endpoints with `/admin/`, `/internal/`, `/manage/`, `/superuser/` in path but no role-based auth guard.
- Flag `DELETE`, `PATCH`, `PUT` HTTP methods without stricter auth than corresponding `GET` on same resource.
- Check for API router groups with admin prefix but shared (non-admin) auth middleware.
- Grep for `if user.role == "admin"` inside route handler instead of middleware — easily bypassed by parameter tampering.
- Flag multi-method route handlers that perform destructive operations without method-specific auth checks.
