---
id: crypto-hardcoded-secrets
title: Hardcoded Secrets and Credentials in Source Code
severity: critical
tags: [cryptography, secrets, hardcoded, credentials, api-keys]
taxonomy: security/cryptography/secrets
references:
  - https://owasp.org/www-community/vulnerabilities/Use_of_Hard-coded_Password
  - https://cheatsheetseries.owasp.org/cheatsheets/Secrets_Management_Cheat_Sheet.html
---

# Hardcoded Secrets and Credentials in Source Code

## Description

Hardcoded credentials in source code are a critical risk: they end up in version control history, logs, build artifacts, and container images. Anyone with repository access — including former employees — can extract them. Known secret prefixes (GitHub tokens, AWS keys, Stripe keys) are actively scanned by attackers in public repos.

## Vulnerable Pattern

```python
# BAD — hardcoded in source
DATABASE_URL = "postgresql://admin:SuperSecret123@prod-db:5432/myapp"
STRIPE_SECRET_KEY = "sk_live_51Hxxxxxxxxxxxxx"
AWS_ACCESS_KEY = "AKIAIOSFODNN7EXAMPLE"
GITHUB_TOKEN = "ghp_xxxxxxxxxxxxxxxxxxxx"
JWT_SECRET = "mysecretkey"

# BAD — in config files committed to repo
# config.yaml
database:
  password: "ProductionPassword!"
smtp:
  password: "emailpassword"
```

```dockerfile
# BAD — secret as build ARG (visible in image layers)
ARG DB_PASSWORD=SuperSecret123
ENV DB_PASSWORD=$DB_PASSWORD
```

## Secure Pattern

```python
# GOOD — load from environment variables
import os
from dotenv import load_dotenv  # only for local dev; not in production

DATABASE_URL = os.environ["DATABASE_URL"]  # raises if not set
JWT_SECRET = os.environ["JWT_SECRET"]

# GOOD — use a secrets manager
import boto3

def get_secret(secret_name: str) -> dict:
    client = boto3.client("secretsmanager", region_name="us-east-1")
    response = client.get_secret_value(SecretId=secret_name)
    return json.loads(response["SecretString"])

db_creds = get_secret("prod/myapp/database")
```

```yaml
# GOOD — GitHub Actions: use encrypted secrets
env:
  DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
  STRIPE_KEY: ${{ secrets.STRIPE_SECRET_KEY }}
```

## Checks to Generate

- Run `git-secrets`, `truffleHog`, `detect-secrets`, or `gitleaks` in CI to scan for credential patterns.
- Grep for known secret prefixes: `AKIA`, `ghp_`, `ghs_`, `sk_live_`, `sk_test_`, `glpat-`, `xoxb-`.
- Flag string literals matching password patterns: `password = "`, `secret = "`, `key = "` (non-env-var).
- Check `.gitignore` for `.env`, `*.pem`, `*.key`, `credentials.json`, `secrets.yaml`.
- Grep for `os.getenv("SECRET", "default_value")` — fallback default defeats the purpose.
- Flag secrets in Docker `ENV` or `ARG` instructions — use secrets mount instead.
