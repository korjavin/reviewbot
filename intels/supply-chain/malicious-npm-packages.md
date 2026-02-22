---
id: supply-chain-malicious-npm
title: Malicious and Typosquatted NPM/PyPI Packages
severity: high
tags: [supply-chain, npm, pypi, typosquatting, malicious-package, dependency]
taxonomy: security/supply-chain/malicious-packages
references:
  - https://owasp.org/www-project-top-ten/2021/A06_2021-Vulnerable_and_Outdated_Components/
  - https://snyk.io/blog/npm-security-preventing-supply-chain-attacks/
---

# Malicious and Typosquatted NPM/PyPI Packages

## Description

Attackers publish packages with names similar to popular packages (typosquatting) or compromise maintainer accounts to inject malicious code into legitimate packages. NPM and PyPI have seen thousands of malicious packages including: `event-stream` (added backdoor to steal Bitcoin), `ua-parser-js` (cryptominer + password stealer), and numerous `colors`/`faker` style protest packages.

Attack patterns:
- **Typosquatting**: `requsts` instead of `requests`, `lodahs` instead of `lodash`
- **Dependency confusion**: internal package names published to public registry
- **Account takeover**: legitimate package maintainer account compromised
- **Malicious install scripts**: `preinstall`/`postinstall` scripts that run on `npm install`

## Vulnerable Pattern

```json
// BAD — unaudited dependency added without verification
{
  "dependencies": {
    "colouers": "^1.0.0",    // typosquatted 'colours' package
    "node-uuid": "latest",    // unpinned — maintainer can push malicious version
    "awesome-utils": "2.1.0"  // unverified new dependency added without review
  }
}
```

```bash
# BAD — installing packages without audit
pip install some-package     # no integrity check, no license/security review
npm install some-package     # no --audit, no lockfile commit
curl -sSf https://example.com/install.sh | sh  # never pipe to shell!
```

## Secure Pattern

```bash
# GOOD — audit before installing
pip install package-name
pip-audit  # check for known vulnerabilities after install

npm install package-name
npm audit  # check audit report
npm audit fix  # fix where possible
```

```yaml
# GOOD — CI: fail on audit findings
- name: npm security audit
  run: npm audit --audit-level=high  # fail on HIGH or CRITICAL

- name: Check for malicious install scripts
  run: npx lockfile-lint --path package-lock.json --type npm
```

```json
// GOOD — lock to exact versions + commit lockfile
{
  "dependencies": {
    "lodash": "4.17.21"  // exact version, no ^/~ range
  }
}
// Commit package-lock.json or yarn.lock to repository
```

## Checks to Generate

- Grep `package.json` / `requirements.txt` for `latest` version tag — must pin exact version.
- Grep for `^` or `~` ranges in `package.json` for security-critical packages (auth, crypto, HTTP).
- Flag `preinstall`/`postinstall` scripts in `package.json` — review carefully for suspicious commands.
- Check that `package-lock.json` / `poetry.lock` / `Pipfile.lock` are committed to repository.
- Flag `npm install` in CI without `--audit` or `npm audit` step.
- Flag `pip install` in CI without `pip-audit` or `safety check` step.
- Grep for recently added single-maintainer packages with < 100 downloads/week.
