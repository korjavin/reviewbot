---
id: owasp-web-a03-injection
title: OWASP A03:2021 — Injection
severity: critical
tags: [owasp-top10, injection, sql, command, ldap, nosql]
taxonomy: security/web/injection
references:
  - https://owasp.org/Top10/A03_2021-Injection/
  - https://cheatsheetseries.owasp.org/cheatsheets/Injection_Prevention_Cheat_Sheet.html
  - https://cheatsheetseries.owasp.org/cheatsheets/Query_Parameterization_Cheat_Sheet.html
---

# OWASP A03:2021 — Injection

## Description

Injection flaws occur when untrusted data is sent to an interpreter as part of a command or query. SQL, NoSQL, OS command, LDAP, and expression-language injections allow attackers to read/modify databases, execute system commands, and bypass authentication. Injection is consistently in the top OWASP risks.

## Vulnerable Pattern

```python
# BAD — SQL injection via string concatenation
def get_user(username: str):
    query = f"SELECT * FROM users WHERE username = '{username}'"
    cursor.execute(query)  # attacker input: ' OR '1'='1

# BAD — OS command injection
import subprocess
def ping_host(host: str):
    result = subprocess.run(f"ping -c 1 {host}", shell=True, capture_output=True)
    # attacker input: "8.8.8.8; rm -rf /"
```

```javascript
// BAD — NoSQL injection (MongoDB)
const user = await db.collection("users").findOne({
    username: req.body.username,
    password: req.body.password  // attacker sends { "$ne": null }
});
```

## Secure Pattern

```python
# GOOD — parameterized query (SQL)
def get_user(username: str):
    cursor.execute("SELECT * FROM users WHERE username = %s", (username,))

# GOOD — avoid shell=True, pass list
def ping_host(host: str):
    import ipaddress
    ipaddress.ip_address(host)  # validate input
    result = subprocess.run(["ping", "-c", "1", host], capture_output=True)
```

```javascript
// GOOD — explicit type check / sanitize for MongoDB
const user = await db.collection("users").findOne({
    username: String(req.body.username),
    password: String(req.body.password)
});
```

## Checks to Generate

- Grep for string concatenation in SQL: `f"SELECT.*{`, `"SELECT" + `, `% username` without parameterization.
- Flag `subprocess.run(..., shell=True)` with user-controlled input.
- Flag `eval(`, `exec(` with any non-literal argument.
- Grep for MongoDB queries where operator keys (`$gt`, `$ne`, `$where`) can come from request body.
- Flag ORM raw query methods: `.raw(`, `execute(f"`, `cursor.execute(query` without parameterized form.
