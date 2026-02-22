---
id: dns-cdn-web-cache-deception
title: Web Cache Deception
severity: high
tags: [cache-deception, cdn, proxy, path-confusion, pii, session]
taxonomy: security/dns-cdn/cache-deception
references:
  - https://owasp.org/www-project-web-security-testing-guide/
  - https://portswigger.net/web-security/web-cache-deception
  - https://omergil.blogspot.com/2017/02/web-cache-deception-attack.html
---

# Web Cache Deception

## Description

Web Cache Deception (WCD) is the inverse of cache poisoning. The attacker tricks the cache into storing a private, authenticated response. They craft a URL like `/account/settings/evil.jpg` — the application ignores the `.jpg` suffix and returns the authenticated user's settings page, but the CDN caches it because `.jpg` extensions are often cached by default.

The attacker then fetches the same URL while unauthenticated, receiving the victim's cached private data.

Attack URL pattern: `https://victim.com/profile/me/anything.css`
1. Victim visits the link (while authenticated)
2. Application serves `/profile/me` content (auth required)
3. CDN caches the response (`.css` = cacheable by default)
4. Attacker fetches the URL (unauthenticated) → gets victim's profile

## Vulnerable Pattern

```python
# BAD — application ignores path suffix, CDN caches by extension
# FastAPI serves /profile regardless of trailing path extension
@app.get("/profile/{username}")
def profile(username: str, current_user=Depends(get_current_user)):
    return {"email": current_user.email, "phone": current_user.phone}
# URL: /profile/me/nonexistent.jpg → still serves auth user's data
# CDN caches /profile/me/nonexistent.jpg as if it were a static asset!
```

```nginx
# BAD — CDN/nginx caching all .css, .js, .jpg, .png by default
# without requiring authentication for cached responses
location ~* \.(css|js|jpg|png|gif|ico)$ {
    proxy_cache cache;
    proxy_cache_valid 200 1d;  # caches ANY response ending in these extensions
    proxy_pass http://backend;
}
```

## Secure Pattern

```python
# GOOD — set Cache-Control: no-store on ALL authenticated responses
from fastapi import Response

@app.get("/profile/{username}")
def profile(username: str, response: Response, current_user=Depends(get_current_user)):
    response.headers["Cache-Control"] = "no-store, private"
    return {"email": current_user.email, "phone": current_user.phone}

# GOOD — global middleware: private/no-store for authenticated sessions
@app.middleware("http")
async def prevent_cache_deception(request: Request, call_next):
    response = await call_next(request)
    # If request has session cookie → response is private
    if "session" in request.cookies or "Authorization" in request.headers:
        response.headers["Cache-Control"] = "no-store, private, must-revalidate"
    return response
```

```nginx
# GOOD — only cache truly static assets (served from /static/ path)
location /static/ {
    proxy_cache static_cache;
    proxy_cache_valid 200 7d;
    proxy_pass http://backend;
}

# Application routes: never cached
location /api/ {
    proxy_cache_bypass 1;
    proxy_no_cache 1;
    proxy_pass http://backend;
}
location /profile/ {
    proxy_cache_bypass 1;
    proxy_no_cache 1;
    proxy_pass http://backend;
}
```

## Checks to Generate

- Grep for authenticated endpoints missing `Cache-Control: no-store` or `private` response header.
- Flag CDN/nginx configs caching by file extension without excluding application routes.
- Check for global cache middleware that doesn't check for authentication headers/cookies before caching.
- Grep for API responses returning user-specific data without `Cache-Control: private` or `no-store`.
- Flag `@cache_page` (Django), `functools.lru_cache` on request handlers, or similar response caching on authenticated views.
- Test: can unauthenticated access to `/profile/me/test.css` retrieve another user's profile data?
