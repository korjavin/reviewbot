---
id: owasp-web-a05-security-misconfiguration
title: OWASP A05:2021 — Security Misconfiguration
severity: high
tags: [owasp-top10, misconfiguration, hardening, headers, debug]
taxonomy: security/web/misconfiguration
references:
  - https://owasp.org/Top10/A05_2021-Security_Misconfiguration/
  - https://cheatsheetseries.owasp.org/cheatsheets/Infrastructure_Security_Cheat_Sheet.html
---

# OWASP A05:2021 — Security Misconfiguration

## Description

Security misconfiguration is the most commonly seen issue. It includes missing security hardening, improperly configured permissions, unnecessary features enabled, default accounts unchanged, overly informative error messages, and missing security headers. This encompasses cloud misconfigurations, XML external entities, and directory listing.

## Vulnerable Pattern

```python
# BAD — debug mode enabled in production (Django)
# settings.py
DEBUG = True
ALLOWED_HOSTS = ["*"]  # exposes detailed tracebacks to users

# BAD — verbose error response leaks stack trace
@app.exception_handler(Exception)
async def global_exception_handler(request, exc):
    return JSONResponse(
        status_code=500,
        content={"error": str(exc), "traceback": traceback.format_exc()}
    )
```

```yaml
# BAD — Docker container runs as root, exposes debug port
FROM python:3.11
# no USER instruction — runs as root
EXPOSE 5678  # debugpy port exposed
CMD ["python", "-m", "debugpy", "--listen", "0.0.0.0:5678", "app.py"]
```

## Secure Pattern

```python
# GOOD — production settings
DEBUG = False
ALLOWED_HOSTS = ["myapp.example.com"]

@app.exception_handler(Exception)
async def global_exception_handler(request, exc):
    logger.error("Unhandled exception", exc_info=exc)  # log internally
    return JSONResponse(status_code=500, content={"error": "Internal server error"})
```

```yaml
# GOOD — non-root user, no debug ports
FROM python:3.11-slim
RUN useradd --system --no-create-home appuser
USER appuser
EXPOSE 8080
CMD ["uvicorn", "app:app", "--host", "0.0.0.0", "--port", "8080"]
```

## Checks to Generate

- Grep for `DEBUG = True` or `debug=True` in production config files.
- Flag `ALLOWED_HOSTS = ["*"]` or `cors_origins=["*"]` on non-development configs.
- Grep for `traceback.format_exc()` or stack trace output in HTTP responses.
- Flag Docker images without `USER` instruction (runs as root).
- Grep for debug ports exposed: `5678`, `5005`, `9229` in Dockerfiles or docker-compose.
- Flag directory listing enabled: `autoindex on` (nginx), `Options +Indexes` (Apache).
- Check for default admin credentials patterns: `admin/admin`, `admin/password` in config files.
