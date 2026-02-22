---
id: injection-redos
title: ReDoS — Regular Expression Denial of Service
severity: medium
tags: [injection, redos, regex, denial-of-service, performance]
taxonomy: security/injection/redos
references:
  - https://owasp.org/www-community/attacks/Regular_expression_Denial_of_Service_-_ReDoS
  - https://snyk.io/learn/regular-expression-denial-of-service-redos/
---

# ReDoS — Regular Expression Denial of Service

## Description

ReDoS exploits catastrophic backtracking in "evil" regular expressions — those with nested quantifiers or overlapping alternatives. A specially crafted input causes exponential time complexity, hanging the thread/event loop. Node.js's single-threaded event loop is particularly vulnerable — one ReDoS request blocks all other requests.

Common vulnerable patterns: `(a+)+`, `([a-zA-Z]+)*`, `(a|aa)+`, repeated groups with overlapping alternatives.

## Vulnerable Pattern

```javascript
// BAD — vulnerable regex with nested quantifiers (Node.js event loop DoS)
const emailRegex = /^([a-zA-Z0-9])(([a-zA-Z0-9])*([._-])*)*([a-zA-Z0-9])+$/;

app.post("/validate", (req, res) => {
    const { email } = req.body;
    if (emailRegex.test(email)) {  // attacker sends: "a".repeat(50) + "!"
        res.json({ valid: true });  // test() hangs for minutes/hours
    }
});
```

```python
# BAD — Python: catastrophically backtracking regex
import re
pattern = re.compile(r"^(\w+\s?)+$")
# Input: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa!" → hangs
```

## Secure Pattern

```javascript
// GOOD — use simple, non-backtracking regex or validator library
const validator = require("validator");

app.post("/validate", (req, res) => {
    const { email } = req.body;
    if (validator.isEmail(email, { allow_utf8_local_part: false })) {
        // validator uses safe, non-vulnerable patterns
        res.json({ valid: true });
    }
});

// GOOD — if custom regex needed, add length limit
app.post("/validate", (req, res) => {
    const { input } = req.body;
    if (input.length > 200) return res.status(400).json({ error: "Input too long" });
    if (/^[a-zA-Z0-9._-]+$/.test(input)) {  // simple character class, no nesting
        res.json({ valid: true });
    }
});
```

```python
# GOOD — timeout regex execution in Python
import signal

class TimeoutError(Exception): pass

def timeout_handler(signum, frame):
    raise TimeoutError("Regex timed out")

signal.signal(signal.SIGALRM, timeout_handler)
signal.alarm(2)  # 2 second timeout
try:
    result = re.match(pattern, user_input)
finally:
    signal.alarm(0)
```

## Checks to Generate

- Grep for nested quantifiers in regex patterns: `(\w+)+`, `([a-z]+)*`, `(a|aa)+`.
- Flag regex patterns applied to user-supplied strings without length limits.
- Grep for `re.compile(`, `new RegExp(` patterns with overlapping alternatives applied to request data.
- Flag email/URL/phone validation using custom regex — recommend validator library instead.
- Check for regex execution timeout mechanism in high-traffic validation endpoints.
