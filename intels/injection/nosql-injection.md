---
id: injection-nosql
title: NoSQL Injection (MongoDB, Redis, Elasticsearch)
severity: high
tags: [injection, nosql, mongodb, elasticsearch, redis]
taxonomy: security/injection/nosql
references:
  - https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/07-Input_Validation_Testing/05.6-Testing_for_NoSQL_Injection
  - https://cheatsheetseries.owasp.org/cheatsheets/Injection_Prevention_Cheat_Sheet.html
---

# NoSQL Injection (MongoDB, Redis, Elasticsearch)

## Description

NoSQL databases have their own injection vectors. MongoDB accepts JSON-based query operators (`$gt`, `$ne`, `$where`) that, when passed from user input, can bypass authentication or exfiltrate data. Elasticsearch accepts full query DSL from user input. Redis SSRF via RESP protocol injection enables command execution.

## Vulnerable Pattern

```javascript
// BAD — MongoDB: operator injection (authentication bypass)
const user = await db.collection("users").findOne({
    username: req.body.username,
    password: req.body.password
    // Attacker sends: { "password": { "$ne": null } } → matches any user!
});

// BAD — MongoDB: $where with user input (JavaScript execution in DB)
db.collection("items").find({
    $where: `this.owner == '${req.query.owner}'`
    // Attacker: '; sleep(5000); //  → time-based injection
})
```

```python
# BAD — Elasticsearch: user-controlled query DSL
query = {
    "query": {
        "query_string": {
            "query": user_input  # attacker can craft any ES query
        }
    }
}
es.search(index="users", body=query)
```

## Secure Pattern

```javascript
// GOOD — MongoDB: type coercion prevents operator injection
const user = await db.collection("users").findOne({
    username: String(req.body.username),   // forces string type
    password: String(req.body.password)    // cannot be an object/operator
});

// GOOD — avoid $where; use regular operators
db.collection("items").find({ owner: String(req.query.owner) })
```

```python
# GOOD — Elasticsearch: match query (no DSL injection)
query = {
    "query": {
        "match": {
            "title": user_input  # treated as text, not DSL
        }
    }
}
es.search(index="items", body=query)
```

## Checks to Generate

- Grep for MongoDB `findOne`, `find`, `updateOne` where query fields come directly from `req.body` without `String()` coercion.
- Flag use of `$where` operator with any dynamic content.
- Flag Elasticsearch `query_string` queries accepting raw user input.
- Grep for Redis commands built with string concatenation from user input.
- Check for `mongoose` `Model.find(req.body)` — passes entire body as query filter.
