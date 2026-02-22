---
id: access-control-idor
title: Insecure Direct Object Reference (IDOR)
severity: high
tags: [access-control, idor, authorization, object-reference]
taxonomy: security/access-control/idor
references:
  - https://owasp.org/www-community/attacks/Insecure_Direct_Object_Reference_Prevention_Cheat_Sheet
  - https://cheatsheetseries.owasp.org/cheatsheets/Authorization_Cheat_Sheet.html
---

# Insecure Direct Object Reference (IDOR)

## Description

IDOR occurs when an application uses user-controllable input to access objects without proper authorization. By modifying object identifiers (IDs, filenames, UUIDs), attackers access data belonging to other users. Often combined with API endpoints that accept integer sequential IDs.

## Vulnerable Pattern

```python
# BAD — document access by ID without ownership check
@app.get("/api/documents/{doc_id}")
def get_document(doc_id: int, user=Depends(get_current_user)):
    doc = db.query(Document).filter(Document.id == doc_id).first()
    return doc  # user can access any document by changing doc_id

# BAD — file download by filename (path-based IDOR)
@app.get("/download/{invoice_filename}")
def download_invoice(invoice_filename: str, user=Depends(get_current_user)):
    path = f"/app/invoices/{invoice_filename}"
    return FileResponse(path)  # user can guess other users' invoice filenames

# BAD — account modification by ID from request body
@app.post("/api/update-email")
def update_email(user_id: int, new_email: str, user=Depends(get_current_user)):
    db.query(User).filter(User.id == user_id).update({"email": new_email})
    # user_id comes from request — attacker changes any account's email
```

## Secure Pattern

```python
# GOOD — always use authenticated user's ID for ownership scoping
@app.get("/api/documents/{doc_id}")
def get_document(doc_id: int, user=Depends(get_current_user)):
    doc = db.query(Document).filter(
        Document.id == doc_id,
        Document.owner_id == user.id  # ownership enforced
    ).first()
    if not doc:
        raise HTTPException(404)
    return doc

# GOOD — use non-guessable identifiers
import uuid
class Document(Base):
    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    # UUID v4 is not sequential — harder to enumerate

# GOOD — derive resource from authenticated context, not request
@app.post("/api/update-email")
def update_email(new_email: str, user=Depends(get_current_user)):
    user.email = new_email  # use authenticated user directly
    db.commit()
```

## Checks to Generate

- Grep for `.filter(Model.id == path_param)` without `.filter(Model.user_id == current_user.id)`.
- Flag endpoints where resource ID in URL path is an integer (sequential, enumerable).
- Check for test coverage: verify that user A cannot access user B's resources.
- Grep for `user_id` as a request body parameter on authenticated endpoints.
- Flag file serving endpoints where filename is user-supplied without ownership validation.
