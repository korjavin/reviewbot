---
id: github-actions-secret-leakage
title: GitHub Actions — Secret Leakage via Environment Variables
severity: high
tags: [github-actions, secrets, ci-cd, security]
taxonomy: security/ci-cd/github-actions
references:
  - https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions
  - https://owasp.org/www-community/vulnerabilities/Improper_Access_Control
---

# GitHub Actions — Secret Leakage via Environment Variables

## Description

Secrets stored as GitHub Actions environment variables can be inadvertently exposed
in workflow logs, pull request comments, or via third-party actions with broad
permissions.

## Vulnerable Pattern

```yaml
# BAD — secret printed to log via env var in run step
jobs:
  build:
    runs-on: ubuntu-latest
    env:
      API_KEY: ${{ secrets.API_KEY }}
    steps:
      - run: echo "Using key $API_KEY"   # leaks to log
      - uses: some-action/print-env@v1   # third-party action can read all env vars
```

## Secure Pattern

```yaml
# GOOD — pass secrets only to steps that need them, never echo
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Deploy
        env:
          API_KEY: ${{ secrets.API_KEY }}
        run: ./deploy.sh   # script reads env, log masked by GitHub
```

## Checks to Generate

- Grep for `echo.*\${{ secrets.` — direct secret echo in run steps.
- Grep for `env:` at job level containing `secrets.` — unnecessarily broad scope.
- Flag third-party actions (`uses: <non-github-org>`) with access to job-level env containing secrets.
