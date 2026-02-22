---
id: javascript-type-confusion
title: JavaScript Type Confusion and Loose Comparison Vulnerabilities
severity: high
tags: [javascript, php, type-juggling, loose-comparison, authentication-bypass]
taxonomy: security/javascript/type-confusion
references:
  - https://owasp.org/www-community/vulnerabilities/Type_Juggling
  - https://portswigger.net/web-security/logic-flaws/examples
---

# JavaScript Type Confusion and Loose Comparison Vulnerabilities

## Description

JavaScript's `==` operator and PHP's type juggling perform type coercion before comparison, creating surprising equality results that can bypass authentication and validation logic. Node.js APIs receiving JSON input are particularly vulnerable because JSON can represent values as different types than expected.

Key surprising equalities:
- `"0" == false` → `true`
- `[] == false` → `true`
- `null == undefined` → `true`
- `"1e0" == 1` → `true`
- `[] == 0` → `true`

## Vulnerable Pattern

```javascript
// BAD — loose comparison for admin check
const { role } = req.body;
if (role == 0) {  // attacker sends role=[] or role=false — also == 0!
    return res.status(403).json({ error: "Not authorized" });
}
// Role check bypassed

// BAD — token comparison with ==
const { token } = req.body;
const expected = getExpectedToken(req.session.userId);
if (token == expected) {  // if expected is 0, token=false also passes
    allowAction();
}

// BAD — JSON body integer expected but string accepted
app.post("/transfer", (req, res) => {
    const amount = req.body.amount;
    if (amount < 0) return res.status(400).json({ error: "Negative amount" });
    // attacker sends: amount = "1e308" → passes < 0 check → Infinity in calculation
    processTransfer(amount);
});
```

```php
# BAD — PHP type juggling in authentication
if ($_POST["password"] == $stored_hash) {
    // $stored_hash = "0e123456" (starts with 0e)
    // $_POST["password"] = "240610708" (MD5 = "0e462097431906509019562988736854")
    // Both "magic hashes" are == 0 in PHP!
    login_user();
}
```

## Secure Pattern

```javascript
// GOOD — strict equality (===) always for security comparisons
const { role } = req.body;
if (role === 0) { ... }  // only exactly 0, not [] or false

// GOOD — type validation before comparison
const amount = req.body.amount;
if (typeof amount !== "number" || !Number.isFinite(amount)) {
    return res.status(400).json({ error: "Amount must be a finite number" });
}
if (amount <= 0) return res.status(400).json({ error: "Amount must be positive" });

// GOOD — constant-time comparison for tokens (also prevents timing attacks)
const crypto = require("crypto");
function safeCompare(a, b) {
    const aStr = String(a);
    const bStr = String(b);
    if (aStr.length !== bStr.length) return false;
    return crypto.timingSafeEqual(Buffer.from(aStr), Buffer.from(bStr));
}
```

```python
# GOOD — Python: explicit type validation with Pydantic
from pydantic import BaseModel, validator

class TransferRequest(BaseModel):
    amount: float  # Pydantic enforces type

    @validator("amount")
    def amount_must_be_positive_finite(cls, v):
        if not (0 < v < 1_000_000):
            raise ValueError("Amount must be between 0 and 1,000,000")
        return v
```

## Checks to Generate

- Grep for `==` (double equals) in authentication, authorization, and token comparison logic — must use `===`.
- Flag `req.body.amount` / `req.params.id` used in arithmetic without `typeof` / `Number.isFinite` check.
- Grep for `parseInt(userInput)` without radix and NaN check — `parseInt("1 UNION SELECT")` returns `1`.
- Flag PHP `==` for password/token comparisons — use `===` or `hash_equals()`.
- Grep for `if (req.body.isAdmin)` — truthy check on user-supplied value; must be `=== true`.
- Flag comparisons against `0` or `false` using `==` on values from request body.
