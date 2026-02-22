---
id: owasp-api-a04-unrestricted-resource-consumption
title: OWASP API A04:2023 — Unrestricted Resource Consumption
severity: high
tags: [owasp-api-top10, rate-limiting, dos, resource-exhaustion, api]
taxonomy: security/api/availability
references:
  - https://owasp.org/API-Security/editions/2023/en/0xa4-unrestricted-resource-consumption/
---

# OWASP API A04:2023 — Unrestricted Resource Consumption

## Description

APIs without rate limiting, request size limits, or resource quotas allow attackers to exhaust server resources (CPU, memory, bandwidth, database connections) or abuse metered services (SMS, email, AI inference) to cause financial damage. Formerly called "Lack of Resources & Rate Limiting."

## Vulnerable Pattern

```python
# BAD — no rate limiting on expensive operation
@router.post("/api/send-sms")
def send_sms(phone: str, message: str, user=Depends(get_current_user)):
    sms_provider.send(phone, message)  # each call costs money — no limit

# BAD — no pagination or result limit on list endpoint
@router.get("/api/logs")
def get_logs(db=Depends(get_db)):
    return db.query(Log).all()  # returns millions of rows → OOM

# BAD — no request body size limit
@router.post("/api/upload-text")
async def upload_text(request: Request):
    body = await request.body()  # no max size — client can send GB of data
    process(body)
```

## Secure Pattern

```python
from slowapi import Limiter
from slowapi.util import get_remote_address

limiter = Limiter(key_func=get_remote_address)

# GOOD — rate limit + per-user quota
@router.post("/api/send-sms")
@limiter.limit("5/hour")
def send_sms(phone: str, message: str, user=Depends(get_current_user)):
    if user.sms_count_today >= 10:
        raise HTTPException(429, "Daily SMS limit reached")
    sms_provider.send(phone, message)
    user.sms_count_today += 1

# GOOD — pagination enforced
@router.get("/api/logs")
def get_logs(page: int = 1, page_size: int = 50, db=Depends(get_db)):
    if page_size > 100:
        page_size = 100  # hard cap
    return db.query(Log).offset((page - 1) * page_size).limit(page_size).all()

# GOOD — body size limit (FastAPI)
from fastapi import FastAPI
app = FastAPI()
app.add_middleware(LimitUploadSize, max_upload_size=5_242_880)  # 5MB
```

## Checks to Generate

- Flag endpoints missing rate-limiting decorator or middleware (especially: auth, SMS, email, payment, AI inference).
- Grep for `.all()` on ORM queries without `.limit()` — unbounded result set.
- Flag file/body upload endpoints without size limit enforcement.
- Check for missing pagination parameters on list/search endpoints.
- Flag AI/LLM inference endpoints without token/request budgets.
- Grep for missing timeout on external HTTP calls (`requests.get(url)` without `timeout=`).
