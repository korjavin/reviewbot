---
id: concurrency-race-conditions
title: Race Conditions and TOCTOU (Time-of-Check Time-of-Use)
severity: high
tags: [concurrency, race-condition, toctou, double-spend, transaction, atomicity]
taxonomy: security/concurrency/race-conditions
references:
  - https://owasp.org/www-community/vulnerabilities/Time_of_check_time_of_use
  - https://portswigger.net/web-security/race-conditions
---

# Race Conditions and TOCTOU (Time-of-Check Time-of-Use)

## Description

Race conditions occur when security decisions are made based on a state check (TOCTOU) that can change between the check and the use. Concurrent requests exploit this window. Common impacts: double-spending funds, using a promo code multiple times, bypassing one-time-token limits, creating duplicate accounts, and inventory overselling.

Web applications are especially vulnerable because multiple requests run concurrently and databases are often checked and updated in non-atomic operations.

## Vulnerable Pattern

```python
# BAD — check-then-act without locking (double-spend)
@app.post("/api/redeem-coupon")
def redeem_coupon(code: str, user=Depends(get_current_user)):
    coupon = db.query(Coupon).filter(
        Coupon.code == code,
        Coupon.used == False  # CHECK
    ).first()
    if not coupon:
        raise HTTPException(400, "Invalid or used coupon")
    # WINDOW: another request can pass the check simultaneously ↑
    coupon.used = True        # USE (too late — race already won)
    db.commit()
    apply_discount(user, coupon.discount)

# BAD — balance check then debit (double-spend)
@app.post("/api/transfer")
def transfer(amount: float, user=Depends(get_current_user)):
    if user.balance < amount:  # CHECK
        raise HTTPException(400, "Insufficient balance")
    # WINDOW ↑
    user.balance -= amount    # USE
    db.commit()
```

```javascript
// BAD — Node.js: async race between check and update
app.post("/use-token", async (req, res) => {
    const token = await db.tokens.findOne({ value: req.body.token, used: false });
    if (!token) return res.status(400).json({ error: "Invalid token" });
    // WINDOW: another concurrent request already here ↑
    await db.tokens.updateOne({ _id: token._id }, { $set: { used: true } });
    // Two requests can both pass the check before either updates
});
```

## Secure Pattern

```python
# GOOD — atomic update with conditional WHERE clause
from sqlalchemy import update

@app.post("/api/redeem-coupon")
def redeem_coupon(code: str, user=Depends(get_current_user)):
    # Atomic: only succeeds if coupon is currently unused
    result = db.execute(
        update(Coupon)
        .where(Coupon.code == code, Coupon.used == False)
        .values(used=True, redeemed_by=user.id)
        .returning(Coupon.discount)
    )
    coupon_row = result.fetchone()
    if not coupon_row:
        raise HTTPException(400, "Invalid or already-used coupon")
    apply_discount(user, coupon_row.discount)
    db.commit()

# GOOD — database-level locking with SELECT FOR UPDATE
@app.post("/api/transfer")
def transfer(amount: float, user=Depends(get_current_user)):
    with db.begin():
        # Lock the user row for the duration of this transaction
        locked_user = db.query(User).filter(
            User.id == user.id
        ).with_for_update().first()
        if locked_user.balance < amount:
            raise HTTPException(400, "Insufficient balance")
        locked_user.balance -= amount
    # Lock released at transaction commit — serialized
```

```javascript
// GOOD — MongoDB: atomic findOneAndUpdate with condition
app.post("/use-token", async (req, res) => {
    const token = await db.tokens.findOneAndUpdate(
        { value: req.body.token, used: false },  // condition
        { $set: { used: true } },                // atomic update
        { returnDocument: "after" }
    );
    if (!token.value) return res.status(400).json({ error: "Invalid token" });
    // No window — check and update are atomic
});
```

## Checks to Generate

- Grep for `if coupon.used == False` / `if token.used == false` followed by a separate `.update()` — non-atomic TOCTOU.
- Grep for balance/credit checks (`user.balance < amount`) without `SELECT FOR UPDATE` or atomic decrement.
- Flag promo code, one-time token, and referral bonus redemption handlers without atomic check-and-set.
- Grep for `findOne` followed by `updateOne` on the same document without atomicity — use `findOneAndUpdate`.
- Flag inventory reservation patterns without database-level locking or optimistic concurrency control.
- Grep for `user.balance -= amount` without transaction and row lock.
- Check for missing idempotency keys on payment/transfer endpoints — duplicate requests must be detected.
