---
id: supply-chain-dependency-confusion
title: Dependency Confusion Attack
severity: critical
tags: [supply-chain, dependency-confusion, npm, pip, package-manager]
taxonomy: security/supply-chain/dependency-confusion
references:
  - https://owasp.org/www-project-top-ten/2017/A9_2017-Using_Components_with_Known_Vulnerabilities
  - https://medium.com/@alex.birsan/dependency-confusion-4a5d60fec610
---

# Dependency Confusion Attack

## Description

Dependency confusion occurs when a package manager fetches a malicious public package instead of an internal private package with the same name. If an internal package `mycompany-auth` exists on a private registry but also exists on the public registry (uploaded by an attacker with a higher version number), the package manager may prefer the public malicious version.

This attack affected Apple, Microsoft, Tesla, and dozens of other organizations in 2021.

## Vulnerable Pattern

```toml
# BAD — internal package referenced by name without scoped registry
# requirements.txt or pyproject.toml
[dependencies]
mycompany-internal-sdk = "1.2.0"  # resolves from PyPI if attacker uploads mycompany-internal-sdk 9.9.9

# .npmrc — no registry scoping for internal packages
# BAD: @mycompany packages resolve to public npm without registry override
```

```yaml
# BAD — CI/CD installs from default registry without verification
- name: Install dependencies
  run: pip install -r requirements.txt
  # no --index-url or --extra-index-url configured for internal packages
```

## Secure Pattern

```ini
# GOOD — .npmrc: scope internal packages to private registry
@mycompany:registry=https://npm.mycompany.com/
# Public packages still resolve from npm, scoped packages from private registry

# pypi — use --index-url to ensure internal packages only come from internal index
# pip.conf:
[global]
index-url = https://pypi.mycompany.com/simple/
extra-index-url = https://pypi.org/simple/
```

```yaml
# GOOD — pin package hash in requirements.txt
mycompany-internal-sdk==1.2.0 \
    --hash=sha256:abc123...   # exact hash — wrong package rejected
```

```yaml
# GOOD — CI: verify package integrity
- name: Verify package hashes
  run: pip install --require-hashes -r requirements.txt
```

## Checks to Generate

- Check `.npmrc` or `.yarnrc.yml` for `@company-scope` registry configuration — internal scopes must point to private registry.
- Grep `requirements.txt` / `pyproject.toml` for internal package names without hash pinning.
- Flag `pip install` commands in CI without `--require-hashes` for internal packages.
- Check `package.json` for unscoped package names that look internal (`company-name-*`, `internal-*`).
- Flag absence of package lock files — prevents version pinning.
