---
id: logging-sensitive-data
title: Sensitive Data Exposure in Log Files
severity: high
tags: [logging, sensitive-data, pii, gdpr, credentials, privacy]
taxonomy: security/logging/sensitive-data
references:
  - https://owasp.org/www-project-top-ten/2021/A09_2021-Security_Logging_and_Monitoring_Failures/
  - https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html
---

# Sensitive Data Exposure in Log Files

## Description

Logging sensitive data (passwords, tokens, PII, credit card numbers, SSNs, API keys) creates secondary exposure: log files aggregated in SIEM systems, accessible to log operators, stored in cloud log services, and potentially retained for years. GDPR considers log files containing PII as personal data requiring proper handling.

## Vulnerable Pattern

```python
# BAD — logging full request body (may contain passwords/tokens)
@app.middleware("http")
async def request_logging(request: Request, call_next):
    body = await request.body()
    logger.info(f"Request: {request.method} {request.url} body={body.decode()}")
    # Logs: POST /login body={"username":"alice","password":"SuperSecret123!"}

# BAD — logging auth headers
logger.debug(f"Calling API with headers: {headers}")
# Logs: Authorization: Bearer sk_live_verylongsecrettoken

# BAD — Django: logging request.POST (includes form passwords)
logger.info(f"Form submitted: {request.POST}")
```

```python
# BAD — exception logging includes sensitive local variables
def process_payment(card_number: str, cvv: str, amount: float):
    try:
        charge(card_number, cvv, amount)
    except Exception as e:
        logger.exception("Payment failed")  # may log local vars including card_number, cvv
```

## Secure Pattern

```python
# GOOD — log only safe fields, never request body on auth endpoints
SENSITIVE_FIELDS = {"password", "token", "secret", "card_number", "cvv", "ssn", "api_key"}

def sanitize_dict(data: dict) -> dict:
    return {k: "***REDACTED***" if k.lower() in SENSITIVE_FIELDS else v
            for k, v in data.items()}

@app.middleware("http")
async def request_logging(request: Request, call_next):
    log.info(
        "request",
        method=request.method,
        path=request.url.path,
        # body NOT logged; query params sanitized
        query_params=sanitize_dict(dict(request.query_params)),
    )
    return await call_next(request)

# GOOD — sanitize headers before logging
safe_headers = {k: v for k, v in headers.items() if k.lower() != "authorization"}
logger.debug("API call", headers=safe_headers)
```

## Checks to Generate

- Grep for `logger.*password`, `log.*token`, `print.*secret`, `logging.*api_key`.
- Flag logging of entire request body (`request.body`, `req.body`, `request.POST`) without sanitization.
- Grep for `logger.exception(` in payment/auth code — may capture sensitive local variables.
- Flag logging of `Authorization` header content.
- Grep for credit card patterns in logging: 16-digit numbers, CVV references.
- Check log aggregation config — ensure PII fields are masked/excluded before shipping to external SIEM.
