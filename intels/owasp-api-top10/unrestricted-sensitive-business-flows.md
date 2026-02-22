---
id: owasp-api-a06-unrestricted-sensitive-business-flows
title: OWASP API A06:2023 — Unrestricted Access to Sensitive Business Flows
severity: high
tags: [owasp-api-top10, business-logic, automation, abuse, bot]
taxonomy: security/api/business-logic
references:
  - https://owasp.org/API-Security/editions/2023/en/0xa6-unrestricted-access-to-sensitive-business-flows/
---

# OWASP API A06:2023 — Unrestricted Access to Sensitive Business Flows

## Description

APIs that expose sensitive business flows (ticket purchasing, inventory reservation, referral bonuses, coupon generation) without compensating controls allow attackers to automate and abuse them at scale. This is different from resource consumption — the flow itself is legitimate, but automation provides unfair advantage or financial gain.

Examples: scalper bots buying concert tickets, account farming for referral bonuses, automated coupon harvesting.

## Vulnerable Pattern

```python
# BAD — ticket purchase with no bot detection or purchase limits
@router.post("/api/events/{event_id}/tickets/buy")
def buy_ticket(event_id: int, quantity: int, user=Depends(get_current_user)):
    tickets = reserve_tickets(event_id, quantity)
    charge_user(user, tickets)
    return tickets
# Bot buys 1000 tickets in seconds — resells at premium

# BAD — referral bonus with no abuse detection
@router.post("/api/referrals/claim")
def claim_referral(referral_code: str):
    bonus = process_referral(referral_code)
    # attacker creates 1000 accounts, farms bonuses
```

## Secure Pattern

```python
# GOOD — multi-layer protection for sensitive business flows
@router.post("/api/events/{event_id}/tickets/buy")
@limiter.limit("3/minute;10/hour")
def buy_ticket(
    event_id: int,
    quantity: int,
    captcha_token: str,
    user=Depends(get_current_user)
):
    # 1. CAPTCHA verification
    if not verify_captcha(captcha_token):
        raise HTTPException(400, "CAPTCHA required")
    # 2. Per-user purchase limit
    existing_purchases = count_user_tickets(user.id, event_id)
    if existing_purchases + quantity > MAX_TICKETS_PER_USER:
        raise HTTPException(400, "Purchase limit exceeded")
    # 3. Device fingerprint / velocity check
    if velocity_check_failed(user.id, request):
        raise HTTPException(429, "Suspicious activity detected")
    tickets = reserve_tickets(event_id, quantity)
    charge_user(user, tickets)
    return tickets
```

## Checks to Generate

- Flag ticket/reservation/purchase endpoints lacking per-user quantity limits.
- Flag bonus/reward/referral claim endpoints without account age or verification requirements.
- Check for absence of CAPTCHA or bot detection on high-value business flows.
- Flag endpoints processing financial transactions without velocity/anomaly checks.
- Grep for missing idempotency keys on payment endpoints — double-submission risk.
