---
id: access-control-privilege-escalation
title: Privilege Escalation via Mass Assignment and Parameter Tampering
severity: high
tags: [access-control, privilege-escalation, mass-assignment, parameter-tampering]
taxonomy: security/access-control/privilege-escalation
references:
  - https://owasp.org/www-project-top-ten/2021/A01_2021-Broken_Access_Control/
  - https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html
---

# Privilege Escalation via Mass Assignment and Parameter Tampering

## Description

Privilege escalation through parameter tampering allows low-privilege users to elevate their own permissions. This happens when APIs accept user-controlled role/permission fields, when hidden form fields are trusted, or when client-side authorization state is not verified server-side.

## Vulnerable Pattern

```python
# BAD — user can set their own role during registration
@app.post("/api/register")
def register(data: dict):
    # Attacker sends: {"username": "alice", "password": "pass", "role": "admin"}
    user = User(**data)  # mass assignment — role field accepted from client
    db.add(user)
    db.commit()

# BAD — price/discount accepted from client in purchase flow
@app.post("/api/purchase")
def purchase(item_id: int, price: float, user=Depends(get_current_user)):
    # Client sends price=0.01 instead of actual price
    charge_user(user, price)  # trusts client-supplied price!

# BAD — hidden form field for admin flag
# <input type="hidden" name="is_admin" value="false">
# Attacker changes to value="true"
```

## Secure Pattern

```python
from pydantic import BaseModel

# GOOD — explicit allowlist of user-settable fields at registration
class RegisterRequest(BaseModel):
    username: str
    email: str
    password: str
    # role, is_admin, permissions NOT included — server assigns defaults

@app.post("/api/register")
def register(data: RegisterRequest):
    user = User(
        username=data.username,
        email=data.email,
        password_hash=hash_password(data.password),
        role="user",          # server-assigned, not from client
        is_admin=False,       # server-assigned
        permissions=[],       # server-assigned
    )
    db.add(user)
    db.commit()

# GOOD — price always fetched from database
@app.post("/api/purchase")
def purchase(item_id: int, user=Depends(get_current_user)):
    item = db.query(Item).filter(Item.id == item_id).first()
    charge_user(user, item.price)  # server-authoritative price
```

## Checks to Generate

- Grep for `User(**data.dict())` or `User(**request.body)` at registration — mass assignment.
- Flag endpoints accepting `role`, `is_admin`, `permissions`, `privilege_level` from client request.
- Grep for price/amount/discount values accepted from client request without server-side validation.
- Flag hidden form fields (`<input type="hidden">`) used for authorization state.
- Check Pydantic models or Django serializers that include privileged fields without write protection.
- Grep for `user.is_admin = request.json.get("is_admin")` style assignments.
