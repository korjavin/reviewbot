---
id: access-control-cors-misconfiguration
title: CORS Misconfiguration
severity: high
tags: [access-control, cors, cross-origin, api, browser-security]
taxonomy: security/access-control/cors
references:
  - https://owasp.org/www-community/attacks/CORS_OriginHeaderScrutiny
  - https://portswigger.net/web-security/cors
---

# CORS Misconfiguration

## Description

CORS misconfigurations allow malicious websites to make authenticated cross-origin API requests from a victim's browser. Common mistakes: reflecting the Origin header without validation, using wildcard with credentials, or trusting all subdomains without restriction. Attackers can steal data by hosting JavaScript on any domain.

## Vulnerable Pattern

```python
# BAD — reflecting Origin header blindly (trusts any origin)
@app.middleware("http")
async def cors_middleware(request: Request, call_next):
    response = await call_next(request)
    origin = request.headers.get("Origin")
    response.headers["Access-Control-Allow-Origin"] = origin  # reflects attacker.com!
    response.headers["Access-Control-Allow-Credentials"] = "true"
    return response

# BAD — regex with substring match (attacker uses "notmyapp.com" or "myapp.com.evil.com")
def is_origin_allowed(origin: str) -> bool:
    return "myapp.com" in origin  # vulnerable!
```

```nginx
# BAD — wildcard with credentials (browsers block but some configs allow)
add_header Access-Control-Allow-Origin *;
add_header Access-Control-Allow-Credentials true;
```

## Secure Pattern

```python
from fastapi.middleware.cors import CORSMiddleware

ALLOWED_ORIGINS = [
    "https://app.mycompany.com",
    "https://admin.mycompany.com",
]

app.add_middleware(
    CORSMiddleware,
    allow_origins=ALLOWED_ORIGINS,      # explicit allowlist
    allow_credentials=True,
    allow_methods=["GET", "POST", "PUT", "DELETE", "PATCH"],
    allow_headers=["Authorization", "Content-Type", "X-CSRF-Token"],
)

# GOOD — strict regex matching for wildcard subdomain (if needed)
import re
ORIGIN_PATTERN = re.compile(r"^https://[a-z0-9-]+\.mycompany\.com$")

def is_origin_allowed(origin: str) -> bool:
    return bool(ORIGIN_PATTERN.match(origin))
```

## Checks to Generate

- Grep for `Access-Control-Allow-Origin: *` combined with `Access-Control-Allow-Credentials: true`.
- Flag `response.headers["Access-Control-Allow-Origin"] = request.headers.get("Origin")` — blind reflection.
- Grep for CORS origin checks using `in` / `contains` / `startswith` — must be exact match or strict regex.
- Flag `null` as an allowed origin — enables attacks from sandboxed iframes.
- Check for missing Vary header (`Vary: Origin`) when CORS responses are cached.
- Flag `allow_methods=["*"]` on sensitive APIs — restrict to needed methods.
