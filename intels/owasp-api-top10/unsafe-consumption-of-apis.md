---
id: owasp-api-a10-unsafe-consumption-of-apis
title: OWASP API A10:2023 — Unsafe Consumption of APIs
severity: high
tags: [owasp-api-top10, third-party, api-client, trust, validation]
taxonomy: security/api/third-party
references:
  - https://owasp.org/API-Security/editions/2023/en/0xaa-unsafe-consumption-of-apis/
---

# OWASP API A10:2023 — Unsafe Consumption of APIs

## Description

Developers often trust data from third-party APIs without validation. If a third-party service is compromised or returns unexpected data, this can lead to injection attacks, SSRF, path traversal, or logic bypasses. The trust relationship is asymmetric — your app receives data it didn't generate.

## Vulnerable Pattern

```python
# BAD — third-party API response used without validation
import requests

def get_user_from_sso(token: str) -> dict:
    resp = requests.get(f"https://sso.example.com/userinfo", headers={"Authorization": f"Bearer {token}"})
    data = resp.json()
    # Blindly trust SSO response — no schema validation
    user_email = data["email"]                    # could be injected payload
    redirect_url = data.get("redirect_after_login")  # SSRF/open redirect if used
    return {"email": user_email, "redirect": redirect_url}

# BAD — external API data inserted into DB without sanitization
def sync_products_from_supplier(supplier_id: int):
    products = supplier_api.get_products(supplier_id)
    for p in products:
        db.execute(f"INSERT INTO products VALUES ('{p['name']}', {p['price']})")
        # supplier API returns: {"name": "'; DROP TABLE products; --", "price": 0}
```

## Secure Pattern

```python
from pydantic import BaseModel, EmailStr, HttpUrl

# GOOD — validate third-party response against schema
class SSOUserInfo(BaseModel):
    sub: str
    email: EmailStr
    email_verified: bool
    name: str | None = None
    # redirect_after_login NOT allowed — controlled by your app only

def get_user_from_sso(token: str) -> SSOUserInfo:
    resp = requests.get("https://sso.example.com/userinfo",
                        headers={"Authorization": f"Bearer {token}"},
                        timeout=5)
    resp.raise_for_status()
    return SSOUserInfo(**resp.json())  # validates schema, rejects extra fields

# GOOD — parameterized insert regardless of data source
def sync_products_from_supplier(supplier_id: int):
    products = supplier_api.get_products(supplier_id)
    for p in products:
        db.execute(
            "INSERT INTO products VALUES (%s, %s)",
            (str(p["name"])[:255], float(p["price"]))  # type coercion + parameterized
        )
```

## Checks to Generate

- Flag direct use of third-party API response fields without schema validation (Pydantic, jsonschema, zod).
- Grep for URL/redirect values extracted from external API responses and used in redirects or fetches.
- Flag SQL queries using data from external API without parameterization.
- Check for missing `timeout` on HTTP client calls to third-party APIs.
- Flag missing TLS certificate verification: `verify=False` in Python requests, `rejectUnauthorized: false` in Node.
- Check for missing error handling when third-party API returns unexpected status codes.
