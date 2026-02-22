---
id: owasp-api-a01-bola
title: OWASP API A01:2023 — Broken Object Level Authorization (BOLA/IDOR)
severity: critical
tags: [owasp-api-top10, bola, idor, authorization, api]
taxonomy: security/api/authorization
references:
  - https://owasp.org/API-Security/editions/2023/en/0xa1-broken-object-level-authorization/
  - https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html
---

# OWASP API A01:2023 — Broken Object Level Authorization (BOLA/IDOR)

## Description

BOLA is the most common and impactful API vulnerability. APIs expose endpoints that handle object identifiers — if the server does not verify that the requesting user has permission to access the specific object, attackers can read or modify any record by guessing/iterating IDs.

Also known as IDOR (Insecure Direct Object Reference). Differs from A01 Web because APIs make it especially easy to enumerate IDs via predictable integers or UUIDs.

## Vulnerable Pattern

```python
# BAD — GET /api/v1/invoices/12345 — no ownership check
@router.get("/invoices/{invoice_id}")
def get_invoice(invoice_id: int, db: Session = Depends(get_db)):
    invoice = db.query(Invoice).filter(Invoice.id == invoice_id).first()
    if not invoice:
        raise HTTPException(404)
    return invoice  # any authenticated user can access any invoice

# BAD — PUT /api/v1/users/{user_id}/profile
@router.put("/users/{user_id}/profile")
def update_profile(user_id: int, data: ProfileUpdate, db=Depends(get_db)):
    db.query(User).filter(User.id == user_id).update(data.dict())
    # no check that user_id == current_user.id
```

## Secure Pattern

```python
# GOOD — ownership check enforced on every object-level access
@router.get("/invoices/{invoice_id}")
def get_invoice(
    invoice_id: int,
    current_user: User = Depends(get_current_user),
    db: Session = Depends(get_db)
):
    invoice = db.query(Invoice).filter(
        Invoice.id == invoice_id,
        Invoice.owner_id == current_user.id  # ownership enforced
    ).first()
    if not invoice:
        raise HTTPException(404)  # same error for not-found and not-authorized
    return invoice

# GOOD — use UUIDs instead of sequential integers to reduce enumeration ease
# (defense-in-depth only — still need ownership checks)
import uuid
class Invoice(Base):
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
```

## Checks to Generate

- Flag REST endpoint functions with `{id}` / `{resource_id}` path parameter but no filter on `owner_id`, `user_id`, or `organization_id`.
- Grep for `db.get(Model, id)` or `.filter(Model.id == id).first()` without additional authorization filter.
- Flag sequential integer IDs in public-facing APIs — recommend UUID to reduce enumeration.
- Check test suite for cross-user access tests (user A trying to access user B's resource).
