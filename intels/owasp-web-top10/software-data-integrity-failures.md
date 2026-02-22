---
id: owasp-web-a08-software-data-integrity-failures
title: OWASP A08:2021 — Software and Data Integrity Failures
severity: high
tags: [owasp-top10, integrity, deserialization, ci-cd, supply-chain, update-mechanism]
taxonomy: security/web/integrity
references:
  - https://owasp.org/Top10/A08_2021-Software_and_Data_Integrity_Failures/
  - https://cheatsheetseries.owasp.org/cheatsheets/Deserialization_Cheat_Sheet.html
---

# OWASP A08:2021 — Software and Data Integrity Failures

## Description

This category covers failures to verify software updates, CI/CD pipeline steps, and critical data. Insecure deserialization of untrusted data, unsigned software updates, and lack of integrity checks on dependencies can allow remote code execution and data manipulation. The SolarWinds and Codecov incidents are notable examples.

## Vulnerable Pattern

```python
# BAD — deserializing untrusted data with pickle (RCE)
import pickle

@app.post("/restore-session")
def restore_session(data: bytes = Body(...)):
    session = pickle.loads(data)  # attacker can craft malicious pickle payload
    return {"user": session["user"]}

# BAD — loading YAML with full loader (arbitrary Python object instantiation)
import yaml
config = yaml.load(user_input, Loader=yaml.FullLoader)  # unsafe
```

```yaml
# BAD — CI/CD using unversioned/unverified third-party actions
- uses: some-org/some-action@main   # no SHA pin — action can change maliciously
- uses: another/action@v1           # semver tag can be moved
```

## Secure Pattern

```python
# GOOD — use JSON or safe deserialization
import json

@app.post("/restore-session")
def restore_session(data: str = Body(...)):
    session = json.loads(data)  # safe — no arbitrary object creation
    return {"user": session.get("user")}

# GOOD — safe YAML loader
config = yaml.safe_load(user_input)
```

```yaml
# GOOD — pin third-party actions to commit SHA
- uses: actions/checkout@11bd71901bbe5b1630ceea73d27597364c9af683  # v4.2.2
- uses: docker/build-push-action@4f58ea79222b3b9dc2c8bbdd6debcef730109a75  # v6
```

## Checks to Generate

- Grep for `pickle.loads(`, `pickle.load(` consuming request data — flag as critical RCE risk.
- Grep for `yaml.load(` without `Loader=yaml.SafeLoader` — unsafe YAML deserialization.
- Grep for `marshal.loads(`, `shelve.open(` with untrusted input.
- Flag CI/CD actions pinned to branch names (`@main`, `@master`) or semver tags instead of commit SHAs.
- Check for missing integrity hashes on downloaded artifacts in CI (`curl | sh`, `wget | bash`).
- Flag auto-update mechanisms that download and execute code without signature verification.
