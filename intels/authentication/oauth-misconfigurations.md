---
id: auth-oauth-misconfigurations
title: OAuth 2.0 / OIDC Misconfigurations
severity: high
tags: [authentication, oauth, oidc, csrf, redirect-uri, token]
taxonomy: security/authentication/oauth
references:
  - https://owasp.org/www-community/vulnerabilities/OAuth_2.0_Vulnerabilities
  - https://cheatsheetseries.owasp.org/cheatsheets/OAuth2_Cheat_Sheet.html
  - https://portswigger.net/web-security/oauth
---

# OAuth 2.0 / OIDC Misconfigurations

## Description

OAuth 2.0 implementations frequently suffer from: missing `state` parameter (CSRF in auth flow), open redirect URIs, token leakage via referrer headers, implicit flow exposing tokens in URL fragments, and insufficient scope validation.

## Vulnerable Pattern

```python
# BAD — missing state parameter (CSRF in OAuth flow)
@app.get("/auth/github")
def github_auth():
    return redirect(
        f"https://github.com/login/oauth/authorize"
        f"?client_id={CLIENT_ID}&redirect_uri={REDIRECT_URI}"
        # missing &state=<random> — CSRF attack possible
    )

# BAD — open redirect URI validation (substring match)
ALLOWED_REDIRECT = "https://myapp.com"
def validate_redirect(uri: str) -> bool:
    return ALLOWED_REDIRECT in uri  # attacker: "https://myapp.com.evil.com"

# BAD — token in URL (logged by servers, visible in browser history)
# Response: redirect_uri=https://app.com/callback#access_token=TOKEN (implicit flow)
```

## Secure Pattern

```python
import secrets, hashlib

# GOOD — state parameter with CSRF validation
@app.get("/auth/github")
def github_auth():
    state = secrets.token_urlsafe(32)
    session["oauth_state"] = state
    return redirect(
        f"https://github.com/login/oauth/authorize"
        f"?client_id={CLIENT_ID}&redirect_uri={REDIRECT_URI}&state={state}"
    )

@app.get("/auth/github/callback")
def github_callback(code: str, state: str):
    if state != session.pop("oauth_state", None):
        raise HTTPException(400, "Invalid state — possible CSRF attack")
    # exchange code for token

# GOOD — exact redirect URI comparison
ALLOWED_REDIRECTS = {"https://myapp.com/callback", "https://myapp.com/auth/done"}
def validate_redirect(uri: str) -> bool:
    return uri in ALLOWED_REDIRECTS  # exact match only
```

## Checks to Generate

- Flag OAuth authorization URL construction missing `state` parameter.
- Grep for redirect URI validation using `in` / `startswith` / substring — must be exact match.
- Flag use of implicit flow (`response_type=token`) — use authorization code + PKCE instead.
- Check PKCE implementation: `code_challenge_method=S256` (not plain).
- Flag `access_token` in URL query parameters (should be in Authorization header or body).
- Grep for missing nonce in OIDC ID token validation.
