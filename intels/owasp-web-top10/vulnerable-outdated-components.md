---
id: owasp-web-a06-vulnerable-outdated-components
title: OWASP A06:2021 — Vulnerable and Outdated Components
severity: high
tags: [owasp-top10, dependencies, sca, supply-chain, cve]
taxonomy: security/web/dependencies
references:
  - https://owasp.org/Top10/A06_2021-Vulnerable_and_Outdated_Components/
  - https://cheatsheetseries.owasp.org/cheatsheets/Vulnerable_Dependency_Management_Cheat_Sheet.html
---

# OWASP A06:2021 — Vulnerable and Outdated Components

## Description

Using components with known vulnerabilities (libraries, frameworks, OS packages) can undermine application defenses and enable attackers to gain control. Log4Shell (Log4j), Spring4Shell, and Heartbleed are examples of critical CVEs in widely-used components. Organizations often do not know which versions are deployed, and patch cadence is poor.

## Vulnerable Pattern

```toml
# BAD — pinned to old, vulnerable versions (Python)
[tool.poetry.dependencies]
django = "2.2.0"        # EOL, multiple CVEs
requests = "2.18.4"     # vulnerable to ReDoS
pillow = "6.2.0"        # multiple critical CVEs
```

```json
// BAD — npm with unpinned ranges allowing vulnerable minor versions
{
  "dependencies": {
    "lodash": "^4.0.0",      // ^4.0.0 matched 4.17.4 which has prototype pollution
    "express": "~4.14.0"     // old patch range with known CVEs
  }
}
```

```dockerfile
# BAD — FROM with no version pin — picks up whatever is latest/cached
FROM node:latest
FROM python:3
```

## Secure Pattern

```yaml
# CI/CD — automated dependency scanning
- name: Run OWASP Dependency-Check
  uses: dependency-check/Dependency-Check_Action@main
  with:
    path: '.'
    format: 'HTML'
    fail_builds_on: 'HIGH'

# Or: use Snyk / Dependabot / pip-audit / npm audit
- name: pip audit
  run: pip-audit --strict
```

```dockerfile
# GOOD — pin exact base image digest
FROM python:3.11.9-slim@sha256:abc123...
```

## Checks to Generate

- Check for absence of dependency scanning step in CI (no `npm audit`, `pip-audit`, `trivy`, `snyk`, or `dependency-check`).
- Flag `FROM <image>:latest` or `FROM <image>` without version tag in Dockerfiles.
- Grep for known EOL version patterns in `requirements.txt`, `package.json`, `pom.xml`.
- Check for missing `dependabot.yml` or `renovate.json` in `.github/` — no automated update PRs.
- Flag unpinned lockfiles — `package-lock.json` / `poetry.lock` not committed.
