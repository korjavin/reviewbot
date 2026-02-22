---
id: owasp-api-a08-security-misconfiguration-api
title: OWASP API A08:2023 — Security Misconfiguration (API)
severity: high
tags: [owasp-api-top10, misconfiguration, cors, headers, api]
taxonomy: security/api/misconfiguration
references:
  - https://owasp.org/API-Security/editions/2023/en/0xa8-security-misconfiguration/
---

# OWASP API A08:2023 — Security Misconfiguration (API)

## Description

Security misconfiguration in APIs includes overly permissive CORS policies, verbose error messages exposing stack traces, unnecessary HTTP methods enabled, missing security headers, outdated TLS, and open API documentation accessible in production.

## Vulnerable Pattern

```python
# BAD — wildcard CORS allowing any origin with credentials
from fastapi.middleware.cors import CORSMiddleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,  # credentials + wildcard is invalid but some frameworks silently allow
    allow_methods=["*"],
    allow_headers=["*"],
)

# BAD — Swagger/OpenAPI UI exposed in production
from fastapi import FastAPI
app = FastAPI(docs_url="/docs", redoc_url="/redoc")  # public API docs with try-it-out

# BAD — verbose error with internal path / version info
{"error": "psycopg2.OperationalError: FATAL: password authentication failed for user 'prod_user'",
 "file": "/app/db/connection.py", "line": 42}
```

## Secure Pattern

```python
# GOOD — strict CORS with allowlist
app.add_middleware(
    CORSMiddleware,
    allow_origins=["https://app.example.com", "https://admin.example.com"],
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE"],
    allow_headers=["Authorization", "Content-Type"],
)

# GOOD — disable docs in production
import os
docs_url = "/docs" if os.getenv("ENV") == "development" else None
app = FastAPI(docs_url=docs_url, redoc_url=None)

# GOOD — generic error response
@app.exception_handler(Exception)
async def error_handler(request, exc):
    logger.error("Unhandled error", exc_info=exc)
    return JSONResponse(500, {"error": "Internal server error"})
```

## Checks to Generate

- Flag `allow_origins=["*"]` combined with `allow_credentials=True` — security anti-pattern.
- Flag Swagger/OpenAPI UI endpoints (`/docs`, `/swagger`, `/api-docs`) in production configs.
- Grep for `OPTIONS`, `TRACE` HTTP methods enabled globally — TRACE especially enables XST attacks.
- Check for missing security headers: `X-Content-Type-Options`, `X-Frame-Options`, `Content-Security-Policy` on API responses.
- Flag stack traces or internal error details in HTTP error responses.
- Check for API version endpoints exposing full technology stack info (`/api/version`, `/api/info`).
