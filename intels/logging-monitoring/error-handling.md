---
id: logging-error-handling
title: Insecure Error Handling — Information Leakage via Errors
severity: medium
tags: [logging, error-handling, information-disclosure, debug, stack-trace]
taxonomy: security/logging/error-handling
references:
  - https://owasp.org/www-community/Improper_Error_Handling
  - https://cheatsheetseries.owasp.org/cheatsheets/Error_Handling_Cheat_Sheet.html
---

# Insecure Error Handling — Information Leakage via Errors

## Description

Verbose error responses expose internal implementation details — database queries, file paths, server versions, stack traces, and environment variables — that significantly aid attackers in reconnaissance. Error handling must be asymmetric: detailed internally for debugging, generic externally for users.

## Vulnerable Pattern

```python
# BAD — Flask default error responses include full traceback
# app.config["DEBUG"] = True  → shows traceback in browser

# BAD — database errors leaked to client
@app.get("/users/{id}")
def get_user(id: int):
    try:
        return db.query(User).filter(User.id == id).first()
    except Exception as e:
        return {"error": str(e)}
        # Returns: "psycopg2.errors.InvalidTextRepresentation: invalid input syntax
        #          for type integer: 'abc' LINE 1: SELECT * FROM users WHERE id = 'abc'"
```

```javascript
// BAD — Express: unhandled rejection reveals stack trace
app.get("/api/data", async (req, res) => {
    const data = await db.query("SELECT...");  // throws if DB down
    res.json(data);
    // Unhandled: Error: ECONNREFUSED 127.0.0.1:5432
    // Stack trace sent to client by default error handler
});
```

## Secure Pattern

```python
import logging
from fastapi import HTTPException, Request
from fastapi.responses import JSONResponse

logger = logging.getLogger(__name__)

# GOOD — global handler: log detailed, return generic
@app.exception_handler(Exception)
async def global_exception_handler(request: Request, exc: Exception):
    error_id = secrets.token_hex(8)  # correlation ID for support
    logger.error(
        "Unhandled exception",
        exc_info=exc,
        error_id=error_id,
        path=request.url.path,
        method=request.method,
    )
    return JSONResponse(
        status_code=500,
        content={
            "error": "Internal server error",
            "error_id": error_id,  # user can provide this to support
            # No stack trace, no SQL, no paths
        }
    )

# GOOD — specific database error handling
@app.get("/users/{id}")
def get_user(id: int):
    try:
        user = db.query(User).filter(User.id == id).first()
    except DatabaseError:
        logger.exception("Database error fetching user %s", id)
        raise HTTPException(503, "Service temporarily unavailable")
    if not user:
        raise HTTPException(404, "Not found")
    return user
```

## Checks to Generate

- Grep for `except Exception as e: return {"error": str(e)}` — raw error leaked to client.
- Flag `DEBUG = True` / `debug=True` in production framework configuration.
- Grep for `traceback.format_exc()` in HTTP response bodies.
- Flag missing global exception handler — unhandled exceptions reach client with defaults.
- Grep for database-specific exception types caught and re-raised with original message.
- Check for `try/except: pass` — silent error swallowing that hides security-relevant failures.
- Flag `console.error(err)` in frontend JavaScript where `err` contains server response bodies.
