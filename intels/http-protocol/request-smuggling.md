---
id: http-request-smuggling
title: HTTP Request Smuggling (CL.TE / TE.CL / HTTP/2 Downgrade)
severity: critical
tags: [http, request-smuggling, desync, proxy, load-balancer, cache-poisoning]
taxonomy: security/http/request-smuggling
references:
  - https://owasp.org/www-community/attacks/HTTP_Request_Smuggling
  - https://portswigger.net/web-security/request-smuggling
  - https://cve.mitre.org/cgi-bin/cvekey.cgi?keyword=request+smuggling
---

# HTTP Request Smuggling (CL.TE / TE.CL / HTTP/2 Downgrade)

## Description

HTTP request smuggling exploits discrepancies between how a front-end proxy and back-end server parse HTTP request boundaries. When they disagree on where one request ends and the next begins, an attacker can "smuggle" a partial request prefix into the back-end's buffer, which is prepended to the next legitimate user's request.

**CL.TE**: Front-end uses `Content-Length`, back-end uses `Transfer-Encoding: chunked`
**TE.CL**: Front-end uses `TE: chunked`, back-end uses `Content-Length`
**HTTP/2 Downgrade**: H2 request downgraded to HTTP/1.1 with attacker-controlled headers

Impacts: bypass front-end auth, poison request cache, steal other users' requests, perform reflected XSS via request body, SSRF.

## Vulnerable Pattern

```
# CL.TE smuggling payload:
POST / HTTP/1.1
Host: vulnerable.com
Content-Length: 13
Transfer-Encoding: chunked

0

SMUGGLED
# Front-end reads 13 bytes (Content-Length) — sees complete request
# Back-end reads chunked body — "0\r\n\r\n" ends the chunk, "SMUGGLED" left in buffer
# "SMUGGLED" is prepended to the NEXT request from any user
```

```nginx
# BAD — nginx as front-end proxy not normalizing conflicting headers
# Default nginx may forward both CL and TE headers to backend
# Backend (Node.js/Python) may interpret differently

# BAD — allowing TE header passthrough to backend
proxy_pass http://backend;
# No: proxy_set_header Transfer-Encoding "";
```

```python
# BAD — backend accepting both Content-Length and Transfer-Encoding
# WSGI servers (gunicorn, uvicorn) have had historical differences
# in how they handle conflicting framing headers
```

## Secure Pattern

```nginx
# GOOD — normalize headers at reverse proxy
proxy_http_version 1.1;
proxy_set_header Connection "";
# Strip conflicting framing headers — let nginx reframe the request
proxy_pass http://backend;
```

```yaml
# GOOD — use HTTP/2 end-to-end (H2C) to eliminate downgrade surface
# Or use same HTTP parsing library front-to-back (no proxy boundary)

# GOOD — WAF rules to reject requests with both CL and TE headers
# HAProxy: option http-server-close; option forwardfor
# Caddy: automatic H2 prevents CL/TE conflicts
```

```python
# GOOD — backend rejects requests with conflicting framing
from starlette.middleware.base import BaseHTTPMiddleware

class RejectAmbiguousFraming(BaseHTTPMiddleware):
    async def dispatch(self, request, call_next):
        has_cl = "content-length" in request.headers
        has_te = "transfer-encoding" in request.headers
        if has_cl and has_te:
            return Response("Bad Request", status_code=400)
        return await call_next(request)
```

## Checks to Generate

- Flag reverse proxy configs (nginx, HAProxy, Caddy) not stripping `Transfer-Encoding` before passing to backend.
- Check for mixed HTTP version between front-end and back-end (HTTP/2 front → HTTP/1.1 back with header downgrade).
- Flag application servers that accept both `Content-Length` and `Transfer-Encoding` on same request.
- Grep for custom HTTP parsing middleware that processes `Content-Length` without rejecting duplicate `Transfer-Encoding`.
- Flag load balancer configs where `keep-alive` is enabled between front-end and back-end with no framing normalization.
- Check for HTTP/2 cleartext (H2C) upgrade handling — H2C upgrade requests can be abused for smuggling.
