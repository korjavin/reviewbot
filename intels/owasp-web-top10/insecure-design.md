---
id: owasp-web-a04-insecure-design
title: OWASP A04:2021 — Insecure Design
severity: high
tags: [owasp-top10, design, threat-modeling, secure-design]
taxonomy: security/web/design
references:
  - https://owasp.org/Top10/A04_2021-Insecure_Design/
  - https://cheatsheetseries.owasp.org/cheatsheets/Threat_Modeling_Cheat_Sheet.html
---

# OWASP A04:2021 — Insecure Design

## Description

Insecure Design is a new category in 2021 focusing on fundamental design and architectural flaws — not implementation bugs. Even correct implementation of a flawed design cannot be secured by patching. This includes lack of threat modeling, missing business logic controls, and absent security controls by design.

Examples: credential recovery via secret questions, unlimited purchase quantities without fraud checks, password reset flows that reveal account existence.

## Vulnerable Pattern

```python
# BAD — password reset reveals account existence (user enumeration)
@app.post("/forgot-password")
def forgot_password(email: str):
    user = db.query(User).filter(User.email == email).first()
    if not user:
        return {"error": "No account with that email"}  # enumeration!
    send_reset_email(user)
    return {"message": "Reset email sent"}

# BAD — no rate limiting / business logic on expensive operation
@app.post("/api/purchase")
def purchase(item_id: int, quantity: int, user=Depends(get_current_user)):
    # no maximum quantity, no fraud check, no inventory validation
    process_order(item_id, quantity, user)
```

## Secure Pattern

```python
# GOOD — generic response regardless of account existence
@app.post("/forgot-password")
def forgot_password(email: str):
    user = db.query(User).filter(User.email == email).first()
    if user:
        send_reset_email(user)
    # always return same response
    return {"message": "If an account exists, a reset email was sent"}

# GOOD — business logic guard
@app.post("/api/purchase")
def purchase(item_id: int, quantity: int, user=Depends(get_current_user)):
    if quantity > MAX_ORDER_QUANTITY:
        raise HTTPException(400, "Quantity exceeds maximum")
    if not inventory_available(item_id, quantity):
        raise HTTPException(400, "Insufficient stock")
    if fraud_score(user, item_id, quantity) > THRESHOLD:
        flag_for_review(user)
        raise HTTPException(403, "Order flagged for review")
    process_order(item_id, quantity, user)
```

## Checks to Generate

- Flag password reset / account recovery endpoints that return different responses for existing vs non-existing accounts (user enumeration).
- Flag numeric input parameters (quantity, amount, count) on business operations lacking upper-bound validation.
- Flag multi-step flows (checkout, registration) where steps can be skipped via direct URL access.
- Check for absence of rate limiting decorators on sensitive endpoints (login, password reset, OTP verification).
- Flag direct object references in account recovery (reset tokens in URL parameters without expiry checks).
