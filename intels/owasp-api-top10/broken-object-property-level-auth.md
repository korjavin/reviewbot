---
id: owasp-api-a03-broken-object-property-level-auth
title: OWASP API A03:2023 — Broken Object Property Level Authorization
severity: high
tags: [owasp-api-top10, mass-assignment, excessive-data-exposure, api]
taxonomy: security/api/authorization
references:
  - https://owasp.org/API-Security/editions/2023/en/0xa3-broken-object-property-level-authorization/
  - https://cheatsheetseries.owasp.org/cheatsheets/Mass_Assignment_Cheat_Sheet.html
---

# OWASP API A03:2023 — Broken Object Property Level Authorization

## Description

This combines two previously separate risks: Excessive Data Exposure (exposing more fields than needed) and Mass Assignment (allowing users to update fields they shouldn't control, like `is_admin`, `role`, `balance`).

APIs that bind user input directly to model objects without property-level filtering are vulnerable to both.

## Vulnerable Pattern

```python
# BAD — mass assignment: user can set is_admin=True
@router.put("/users/me")
def update_profile(data: dict, current_user=Depends(get_current_user), db=Depends(get_db)):
    db.query(User).filter(User.id == current_user.id).update(data)
    # attacker sends: {"is_admin": true, "role": "superuser", "credit_balance": 99999}

# BAD — excessive data exposure: full model returned including sensitive fields
@router.get("/users/{user_id}")
def get_user(user_id: int, db=Depends(get_db)):
    return db.query(User).filter(User.id == user_id).first()
    # returns: password_hash, totp_secret, admin_notes, internal_score...
```

## Secure Pattern

```python
from pydantic import BaseModel

# GOOD — explicit allowlist of updateable fields
class ProfileUpdateRequest(BaseModel):
    display_name: str | None = None
    bio: str | None = None
    avatar_url: str | None = None
    # is_admin, role, balance NOT included

@router.put("/users/me")
def update_profile(data: ProfileUpdateRequest, current_user=Depends(get_current_user)):
    update_data = data.dict(exclude_unset=True)  # only set fields
    db.query(User).filter(User.id == current_user.id).update(update_data)

# GOOD — response schema strips sensitive fields
class PublicUserResponse(BaseModel):
    id: int
    display_name: str
    created_at: datetime
    # no password_hash, no totp_secret, no internal fields

    class Config:
        from_attributes = True

@router.get("/users/{user_id}", response_model=PublicUserResponse)
def get_user(user_id: int):
    return db.query(User).filter(User.id == user_id).first()
```

## Checks to Generate

- Flag endpoints accepting `dict` or `**kwargs` directly into ORM `.update()` — no field filtering.
- Grep for SQLAlchemy `.update(data.dict())` or Django `User.objects.filter(...).update(**data)` without allowlist.
- Flag API response models that include fields like `password`, `password_hash`, `secret`, `token`, `internal_*`.
- Check FastAPI/Django REST routes for missing `response_model` — response schema not enforced.
- Grep for `request.body`, `req.body`, `request.json()` being passed directly to model constructors.
