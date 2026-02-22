---
id: owasp-web-a01-broken-access-control
title: OWASP A01:2021 — Broken Access Control
severity: critical
tags: [owasp-top10, access-control, authorization, idor]
taxonomy: security/web/access-control
references:
  - https://owasp.org/Top10/A01_2021-Broken_Access_Control/
  - https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html
---

# OWASP A01:2021 — Broken Access Control

## Description

Broken Access Control is the #1 OWASP risk. It occurs when users can act outside their intended permissions — accessing data or performing operations reserved for other users or roles. Common failures include missing authorization checks, insecure direct object references (IDOR), CORS misconfigurations, and privilege escalation.

94% of applications tested had some form of broken access control.

## Vulnerable Pattern

```python
# BAD — user ID comes from request; no ownership check
@app.get("/api/orders/{order_id}")
def get_order(order_id: int, current_user: User = Depends(get_current_user)):
    order = db.query(Order).filter(Order.id == order_id).first()
    return order  # any authenticated user can read any order
```

```python
# BAD — role check missing on admin endpoint
@app.delete("/api/admin/users/{user_id}")
def delete_user(user_id: int, current_user: User = Depends(get_current_user)):
    db.query(User).filter(User.id == user_id).delete()  # no admin check!
```

## Secure Pattern

```python
# GOOD — enforce ownership before returning data
@app.get("/api/orders/{order_id}")
def get_order(order_id: int, current_user: User = Depends(get_current_user)):
    order = db.query(Order).filter(
        Order.id == order_id,
        Order.user_id == current_user.id  # ownership enforced
    ).first()
    if not order:
        raise HTTPException(status_code=404)
    return order

# GOOD — role-based guard on privileged endpoint
@app.delete("/api/admin/users/{user_id}")
def delete_user(user_id: int, current_user: User = Depends(require_role("admin"))):
    db.query(User).filter(User.id == user_id).delete()
```

## Checks to Generate

- Flag endpoints that accept an ID parameter but do not filter by `user_id` or ownership field.
- Flag admin/privileged routes missing role/permission middleware.
- Grep for direct `db.query(Model).filter(Model.id == id)` without additional owner constraint.
- Flag `@app.route` / `@router.*` decorators on sensitive paths (`/admin`, `/internal`, `/manage`) lacking auth middleware.
- Check CORS config — flag wildcard `Access-Control-Allow-Origin: *` on authenticated endpoints.
