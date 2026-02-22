---
id: access-control-missing-function-level
title: Missing Function-Level Access Control
severity: high
tags: [access-control, authorization, rbac, middleware, endpoint-protection]
taxonomy: security/access-control/function-level
references:
  - https://owasp.org/www-project-top-ten/2021/A01_2021-Broken_Access_Control/
  - https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html
---

# Missing Function-Level Access Control

## Description

Applications often properly secure their UI — hiding admin menus from regular users — but fail to enforce equivalent restrictions at the API level. Attackers who discover admin API endpoints (through source code, documentation, or fuzzing) can call them directly without being stopped server-side.

This differs from BOLA (object-level) — here the issue is that the entire function/endpoint is accessible, not just specific objects within it.

## Vulnerable Pattern

```python
# BAD — role check only in UI, not API
# Frontend: shows "Export Users" button only to admins
# But the API endpoint has no enforcement:

@app.get("/api/export-users")
def export_users(user=Depends(get_current_user)):
    return db.query(User).all()  # any authenticated user can call this!

# BAD — inconsistent auth across HTTP methods
@app.route("/api/users/<int:id>", methods=["GET", "DELETE"])
def user_resource(id):
    if request.method == "GET":
        return get_user(id)  # auth checked above
    elif request.method == "DELETE":
        delete_user(id)  # same function, but DELETE has no separate auth check!
```

## Secure Pattern

```python
from functools import wraps

def require_permission(permission: str):
    def decorator(f):
        @wraps(f)
        def wrapper(*args, current_user: User = Depends(get_current_user), **kwargs):
            if permission not in current_user.permissions:
                raise HTTPException(403, f"Requires permission: {permission}")
            return f(*args, current_user=current_user, **kwargs)
        return wrapper
    return decorator

# GOOD — explicit permission on every sensitive endpoint
@app.get("/api/export-users")
@require_permission("users:export")
def export_users(current_user=Depends(get_current_user)):
    return db.query(User).all()

# GOOD — separate endpoints with different auth requirements
@app.get("/api/users/{id}")
def get_user(id: int, user=Depends(get_current_user)):
    # Regular user auth
    ...

@app.delete("/api/admin/users/{id}")
@require_permission("users:delete")
def delete_user(id: int, user=Depends(get_current_user)):
    # Admin permission required
    ...
```

## Checks to Generate

- Flag `@app.route` with multiple methods where auth differs between methods or is missing for destructive methods.
- Grep for admin functionality (`export`, `delete-all`, `impersonate`, `reset-all`) without admin role check.
- Flag endpoints that check `current_user` but don't verify specific permissions/roles for the action.
- Check for route groups with auth middleware applied at group level but individual routes can override or bypass.
- Grep for `if request.method == "POST":` inside route handler — should use separate decorated routes.
