---
id: infra-default-credentials
title: Default and Weak Credentials on Services
severity: critical
tags: [infrastructure, default-credentials, credentials, misconfiguration, databases]
taxonomy: security/infrastructure/default-credentials
references:
  - https://owasp.org/www-project-top-ten/2021/A05_2021-Security_Misconfiguration/
  - https://cve.mitre.org/cgi-bin/cvekey.cgi?keyword=default+credentials
---

# Default and Weak Credentials on Services

## Description

Default credentials on databases, message queues, management interfaces, and IoT devices are one of the most exploited attack vectors. Attackers run automated scans using databases of known default credentials. Services deployed with factory defaults provide no real access control.

Common defaults: `admin/admin`, `root/root`, `postgres/postgres`, `elastic/`, `redis/` (no password), `admin/password`, `mongo/` (no auth by default in older versions).

## Vulnerable Pattern

```yaml
# BAD — docker-compose with default/weak database credentials
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD: postgres  # default, widely known
      POSTGRES_USER: postgres

  redis:
    image: redis:7
    # No password configured — open to anyone with network access

  elasticsearch:
    image: elasticsearch:8
    environment:
      xpack.security.enabled: "false"  # auth disabled!
```

```python
# BAD — connecting with hardcoded default credentials
DATABASES = {
    "default": {
        "ENGINE": "django.db.backends.postgresql",
        "NAME": "mydb",
        "USER": "admin",
        "PASSWORD": "admin",  # default credential
        "HOST": "db",
    }
}
```

## Secure Pattern

```yaml
# GOOD — strong random credentials from secrets manager
services:
  postgres:
    image: postgres:15
    environment:
      POSTGRES_PASSWORD_FILE: /run/secrets/db_password
      POSTGRES_USER: appuser
    secrets:
      - db_password

  redis:
    image: redis:7
    command: redis-server --requirepass "${REDIS_PASSWORD}" --bind 127.0.0.1

secrets:
  db_password:
    external: true  # from Docker secrets / Vault
```

```bash
# GOOD — generate strong random password for service setup
openssl rand -base64 32  # 256-bit random password
# Store in secrets manager, not in code/config files
```

## Checks to Generate

- Grep docker-compose / k8s env vars for `PASSWORD: postgres`, `PASSWORD: admin`, `PASSWORD: password`.
- Flag Redis deployments without `requirepass` in command or config.
- Flag Elasticsearch with `xpack.security.enabled: false`.
- Grep for MongoDB connection strings without authentication (`mongodb://localhost/db` no user/pass).
- Flag RabbitMQ with default `guest/guest` credentials.
- Grep for `admin/admin`, `root/root`, `test/test` in any connection string or config.
- Check for Grafana/Kibana/Jenkins default admin credentials not rotated.
