---
id: concurrency-double-submit
title: Double Submit / Duplicate Request Vulnerabilities
severity: medium
tags: [concurrency, double-submit, idempotency, payment, duplicate, retry]
taxonomy: security/concurrency/double-submit
references:
  - https://owasp.org/www-community/vulnerabilities/Time_of_check_time_of_use
  - https://stripe.com/docs/api/idempotent_requests
---

# Double Submit / Duplicate Request Vulnerabilities

## Description

Double submit vulnerabilities arise when the same state-changing operation is executed multiple times due to: network retries, user double-clicking, browser back/forward cache, or intentional replay by an attacker. Without idempotency controls, this causes double-charges, duplicate orders, multiple transfers, or multiple account registrations.

Idempotency is the property that executing the same operation multiple times produces the same result as executing it once.

## Vulnerable Pattern

```python
# BAD — payment processed without idempotency key
@app.post("/api/checkout")
def checkout(cart_id: int, payment_method: str, user=Depends(get_current_user)):
    cart = get_cart(cart_id)
    charge = stripe.PaymentIntent.create(
        amount=cart.total_cents,
        currency="usd",
        payment_method=payment_method,
    )
    # User clicks "Pay" twice → two charges created
    # Network retry → two charges created
    create_order(cart, charge.id, user)
    return {"order_id": order.id}

# BAD — no duplicate detection on resource creation
@app.post("/api/orders")
def create_order(data: OrderCreate, user=Depends(get_current_user)):
    order = Order(**data.dict(), user_id=user.id)
    db.add(order)
    db.commit()
    # Retry → duplicate order with same data
```

## Secure Pattern

```python
import hashlib

# GOOD — idempotency key prevents duplicate processing
@app.post("/api/checkout")
def checkout(
    cart_id: int,
    payment_method: str,
    idempotency_key: str = Header(...),  # client sends unique key per attempt
    user=Depends(get_current_user)
):
    # Check if this key was already processed
    existing = db.query(IdempotencyRecord).filter(
        IdempotencyRecord.key == idempotency_key,
        IdempotencyRecord.user_id == user.id,
    ).first()
    if existing:
        return existing.response  # return cached response — no duplicate charge

    cart = get_cart(cart_id)
    charge = stripe.PaymentIntent.create(
        amount=cart.total_cents,
        currency="usd",
        payment_method=payment_method,
        idempotency_key=idempotency_key,  # also pass to Stripe
    )
    order = create_order(cart, charge.id, user)

    # Store result for future duplicate requests
    record = IdempotencyRecord(
        key=idempotency_key,
        user_id=user.id,
        response={"order_id": order.id},
        expires_at=datetime.utcnow() + timedelta(hours=24),
    )
    db.add(record)
    db.commit()
    return {"order_id": order.id}

# GOOD — database unique constraint as last-resort deduplication
class Order(Base):
    __table_args__ = (
        UniqueConstraint("user_id", "cart_id", "idempotency_key"),
    )
```

## Checks to Generate

- Grep for payment processing endpoints without `idempotency_key` parameter or header.
- Flag Stripe/PayPal/Braintree API calls without `idempotency_key` argument.
- Grep for `POST /orders`, `POST /payments`, `POST /transfers` without duplicate-detection logic.
- Check database models for missing `UniqueConstraint` on business-key fields that should be unique per operation.
- Flag form submission endpoints without CSRF token (which also prevents double-submit in some frameworks).
- Grep for retry logic in clients/workers that calls non-idempotent endpoints without idempotency keys.
