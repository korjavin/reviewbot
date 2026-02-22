---
id: owasp-web-a10-ssrf
title: OWASP A10:2021 — Server-Side Request Forgery (SSRF)
severity: high
tags: [owasp-top10, ssrf, cloud, imds, internal-network]
taxonomy: security/web/ssrf
references:
  - https://owasp.org/Top10/A10_2021-Server_Side_Request_Forgery_%28SSRF%29/
  - https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html
---

# OWASP A10:2021 — Server-Side Request Forgery (SSRF)

## Description

SSRF occurs when an application fetches a remote resource based on a user-supplied URL without validating it. Attackers can use it to scan internal networks, access cloud metadata services (AWS IMDS, GCP metadata), exfiltrate credentials, or reach internal services not exposed to the internet.

Notable incidents: Capital One breach (AWS IMDS via SSRF), GitLab SSRF, and many cloud-native app breaches.

## Vulnerable Pattern

```python
# BAD — fetches arbitrary user-supplied URL
import requests

@app.post("/api/fetch-preview")
def fetch_preview(url: str):
    response = requests.get(url, timeout=10)  # attacker supplies http://169.254.169.254/...
    return {"content": response.text}

# BAD — webhook delivery without URL validation
@app.post("/api/webhooks")
def create_webhook(webhook_url: str, user=Depends(get_current_user)):
    db.add(Webhook(url=webhook_url, user_id=user.id))
    # attacker registers http://internal-service/admin as webhook
```

## Secure Pattern

```python
from urllib.parse import urlparse
import ipaddress

ALLOWED_SCHEMES = {"https"}
BLOCKED_HOSTS = {"169.254.169.254", "metadata.google.internal", "localhost", "127.0.0.1"}

def validate_url(url: str) -> bool:
    parsed = urlparse(url)
    if parsed.scheme not in ALLOWED_SCHEMES:
        return False
    hostname = parsed.hostname
    if hostname in BLOCKED_HOSTS:
        return False
    try:
        addr = ipaddress.ip_address(hostname)
        if addr.is_private or addr.is_loopback or addr.is_link_local:
            return False
    except ValueError:
        pass  # hostname, not IP
    return True

@app.post("/api/fetch-preview")
def fetch_preview(url: str):
    if not validate_url(url):
        raise HTTPException(400, "URL not allowed")
    response = requests.get(url, timeout=5, allow_redirects=False)
    return {"content": response.text[:1000]}  # limit response size
```

## Checks to Generate

- Grep for `requests.get(url`, `httpx.get(url`, `urllib.request.urlopen(url` where `url` derives from request parameters.
- Flag functions that accept a URL parameter and immediately fetch it without validation.
- Check for DNS rebinding protection absence — flag missing hostname re-resolution after validation.
- Flag `allow_redirects=True` (default) on requests that fetch user-supplied URLs — redirect can bypass IP blocklists.
- Grep for use of AWS IMDS endpoint in non-infrastructure code: `169.254.169.254`.
- Flag webhook/callback URL fields in API endpoints missing allowlist validation.
