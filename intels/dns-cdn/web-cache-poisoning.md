---
id: dns-cdn-web-cache-poisoning
title: Web Cache Poisoning
severity: high
tags: [cache-poisoning, cdn, proxy, unkeyed-headers, host-header, xss]
taxonomy: security/dns-cdn/cache-poisoning
references:
  - https://owasp.org/www-community/attacks/Cache_Poisoning
  - https://portswigger.net/web-security/web-cache-poisoning
  - https://portswigger.net/research/practical-web-cache-poisoning
---

# Web Cache Poisoning

## Description

Web cache poisoning occurs when an attacker injects malicious content into a shared cache (CDN, reverse proxy, in-process cache), causing it to serve the poisoned response to legitimate users. The attack exploits "unkeyed" inputs — HTTP headers the cache doesn't include in its key but the application uses to generate its response (e.g., `X-Forwarded-Host`, `X-Forwarded-Scheme`, `X-Original-URL`).

A poisoned cache entry affects every user who requests the same URL until the cache expires.

## Vulnerable Pattern

```python
# BAD — application uses X-Forwarded-Host to construct URLs (unkeyed by CDN)
from flask import request

@app.get("/")
def homepage():
    # CDN caches by URL only. X-Forwarded-Host is unkeyed.
    # Attacker sends: X-Forwarded-Host: evil.com
    # CDN caches response with: <script src="https://evil.com/analytics.js">
    # All users receive poisoned response!
    host = request.headers.get("X-Forwarded-Host", request.host)
    return render_template("index.html", analytics_url=f"https://{host}/analytics.js")
```

```nginx
# BAD — nginx caching without Vary header for unkeyed inputs
proxy_cache_path /var/cache/nginx levels=1:2 keys_zone=cache:10m;
location / {
    proxy_cache cache;
    proxy_cache_key "$scheme$request_method$host$request_uri";
    # X-Forwarded-Host NOT in cache key — but app uses it!
    proxy_pass http://backend;
}
```

```javascript
// BAD — Express: response varies on X-Original-URL (unkeyed by Varnish)
app.get("/search", (req, res) => {
    const query = req.headers["x-original-url"]?.split("?q=")[1] || req.query.q;
    // Attacker overrides X-Original-URL to inject XSS payload into cached search results
    res.send(`<h1>Results for: ${query}</h1>`);
});
```

## Secure Pattern

```python
# GOOD — never use unkeyed headers to generate response content
from flask import current_app

@app.get("/")
def homepage():
    # Use configured base URL, never request-derived host for cached content
    analytics_url = current_app.config["ANALYTICS_URL"]
    return render_template("index.html", analytics_url=analytics_url)
```

```nginx
# GOOD — if application uses a header, include it in cache key
proxy_cache_key "$scheme$request_method$host$request_uri$http_x_forwarded_host";
# OR: strip the header before passing to application
proxy_set_header X-Forwarded-Host "";  # normalize away — app can't use it
```

```python
# GOOD — add Vary header for every header the response depends on
@app.after_request
def add_vary(response):
    if "X-Forwarded-Host" in request.headers:
        response.headers["Vary"] = "X-Forwarded-Host"
        # CDN now caches separately per X-Forwarded-Host value
    return response
```

## Checks to Generate

- Grep for `request.headers.get("X-Forwarded-Host"`, `X-Forwarded-Scheme"`, `X-Original-URL"`, `X-Rewrite-URL"` used in response generation.
- Grep for CDN/nginx cache configs where cache key doesn't include all headers the application reads.
- Flag missing `Vary` header on responses that differ based on request headers.
- Grep for `request.headers.get("X-Forwarded-For"` used in business logic (not just logging) — can be unkeyed.
- Check `Cache-Control: no-store` or `private` on pages that use user-specific or header-dependent content.
- Flag any response that embeds `request.host` or `request.headers.*` content while being publicly cached.
