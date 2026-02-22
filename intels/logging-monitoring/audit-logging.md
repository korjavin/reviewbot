---
id: logging-audit-logging
title: Insufficient Audit Logging for Security Events
severity: medium
tags: [logging, audit, monitoring, incident-response, forensics]
taxonomy: security/logging/audit
references:
  - https://owasp.org/www-project-top-ten/2021/A09_2021-Security_Logging_and_Monitoring_Failures/
  - https://cheatsheetseries.owasp.org/cheatsheets/Logging_Cheat_Sheet.html
---

# Insufficient Audit Logging for Security Events

## Description

Security audit logs are the foundation of incident detection and response. Without logging key security events — authentication, authorization failures, data access, configuration changes — breaches go undetected for months. GDPR and compliance frameworks (SOC2, PCI DSS) require audit trails for sensitive operations.

Minimum events to log: logins (success/fail), logouts, password changes, MFA events, privilege changes, admin actions, data exports, API key operations.

## Vulnerable Pattern

```python
# BAD — no logging on security-sensitive operations
@app.post("/admin/export-users")
def export_users(user=Depends(require_admin)):
    users = db.query(User).all()
    return users  # sensitive data export — not logged!

@app.post("/api/change-role")
def change_role(target_user_id: int, new_role: str, admin=Depends(require_admin)):
    db.query(User).filter(User.id == target_user_id).update({"role": new_role})
    # No audit record of who changed whose role to what
```

## Secure Pattern

```python
import structlog
from datetime import datetime

audit_log = structlog.get_logger("audit")

def log_security_event(
    event_type: str,
    actor_id: int,
    target: str = None,
    details: dict = None,
    ip: str = None,
    outcome: str = "success"
):
    audit_log.info(
        event_type,
        actor_id=actor_id,
        target=target,
        details=details or {},
        ip=ip,
        outcome=outcome,
        timestamp=datetime.utcnow().isoformat(),
    )

# GOOD — log admin data export
@app.post("/admin/export-users")
def export_users(request: Request, admin=Depends(require_admin)):
    log_security_event(
        "user_data_export",
        actor_id=admin.id,
        details={"format": "csv", "record_count": user_count},
        ip=request.client.host,
    )
    users = db.query(User).all()
    return users

# GOOD — log privilege change
@app.post("/api/change-role")
def change_role(target_user_id: int, new_role: str, request: Request, admin=Depends(require_admin)):
    old_role = db.query(User).get(target_user_id).role
    db.query(User).filter(User.id == target_user_id).update({"role": new_role})
    log_security_event(
        "role_change",
        actor_id=admin.id,
        target=f"user:{target_user_id}",
        details={"old_role": old_role, "new_role": new_role},
        ip=request.client.host,
    )
```

## Checks to Generate

- Flag admin endpoints (`/admin/*`) without `log_security_event` / audit log call.
- Flag data export endpoints without logging.
- Flag user account operations (create, delete, role change, password change) without audit logging.
- Grep for authentication events logged without IP address.
- Check for structured logging library — plain `print()` or unstructured `logger.info(str)` insufficient for SIEM.
- Flag audit logs written to same log file as application logs — security logs should be separate and tamper-evident.
