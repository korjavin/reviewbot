---
id: owasp-web-a09-security-logging-monitoring-failures
title: OWASP A09:2021 — Security Logging and Monitoring Failures
severity: medium
tags: [owasp-top10, logging, monitoring, incident-response, audit-trail]
taxonomy: security/web/logging
references:
  - https://owasp.org/Top10/A09_2021-Security_Logging_and_Monitoring_Failures/
  - https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html
---

# OWASP A09:2021 — Security Logging and Monitoring Failures

## Description

Insufficient logging and monitoring means that breaches go undetected. Without adequate audit trails, attackers can persist for months without discovery. OWASP reports that most breach studies show detection time exceeds 200 days. Login failures, access control violations, and high-value transactions must be logged with enough context to support incident response.

## Vulnerable Pattern

```python
# BAD — login failure not logged
@app.post("/login")
def login(credentials: LoginRequest):
    user = authenticate(credentials)
    if not user:
        raise HTTPException(401, "Invalid credentials")  # silent failure
    return create_session(user)

# BAD — logging sensitive data (PII/secrets in logs)
logger.info(f"User login: {username}, password: {password}, token: {token}")

# BAD — no structured logging — cannot be parsed by SIEM
print(f"Error occurred: {e}")
```

## Secure Pattern

```python
import structlog
log = structlog.get_logger()

# GOOD — log security events with context, without sensitive data
@app.post("/login")
def login(request: Request, credentials: LoginRequest):
    user = authenticate(credentials)
    if not user:
        log.warning(
            "authentication_failure",
            username=credentials.username,
            ip=request.client.host,
            user_agent=request.headers.get("user-agent"),
        )
        raise HTTPException(401, "Invalid credentials")
    log.info(
        "authentication_success",
        user_id=user.id,
        ip=request.client.host,
    )
    return create_session(user)

# GOOD — log access control violations separately (high priority)
@app.get("/admin")
def admin_panel(current_user: User = Depends(get_current_user)):
    if not current_user.is_admin:
        log.error("access_control_violation", user_id=current_user.id, endpoint="/admin")
        raise HTTPException(403)
```

## Checks to Generate

- Flag login/authentication endpoints with no logging call on failure path.
- Grep for `logger.*password`, `log.*token`, `print.*secret` — sensitive data in logs.
- Flag absence of structured logging library (structlog, python-json-logger) — plain `print()` is unstructured.
- Check for missing audit log on privileged operations (user deletion, role change, payment).
- Flag absence of log aggregation / SIEM integration in infrastructure config.
- Grep for exception handlers that swallow errors silently (`except: pass`, `catch(e) {}`).
