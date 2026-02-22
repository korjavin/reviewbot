---
id: frontend-xss-dom
title: DOM-Based Cross-Site Scripting (XSS)
severity: high
tags: [xss, frontend, dom, javascript, browser-security]
taxonomy: security/frontend/xss-dom
references:
  - https://owasp.org/www-community/attacks/DOM_Based_XSS
  - https://cheatsheetseries.owasp.org/cheatsheets/DOM_based_XSS_Prevention_Cheat_Sheet.html
---

# DOM-Based Cross-Site Scripting (XSS)

## Description

DOM XSS occurs entirely client-side: JavaScript reads attacker-controlled data (URL hash, query params, `localStorage`, `postMessage`) and writes it to dangerous DOM sinks without sanitization. The server never sees the payload — it bypasses server-side filtering.

Sources: `location.hash`, `location.search`, `document.referrer`, `postMessage`, `localStorage`
Sinks: `innerHTML`, `outerHTML`, `document.write`, `eval`, `setTimeout(string)`, `src`, `href`

## Vulnerable Pattern

```javascript
// BAD — reading URL hash into innerHTML
const params = new URLSearchParams(location.search);
const name = params.get("name");
document.getElementById("greeting").innerHTML = `Hello, ${name}!`;
// URL: ?name=<img src=x onerror=alert(document.cookie)>

// BAD — eval with user-controlled data
const callback = location.hash.slice(1);
eval(callback + "()");  // URL: #alert(document.cookie)

// BAD — jQuery .html() with user input
$(selector).html(userInput);

// BAD — postMessage listener without origin check
window.addEventListener("message", (event) => {
    document.getElementById("content").innerHTML = event.data;
    // Any window can send malicious HTML via postMessage
});
```

## Secure Pattern

```javascript
// GOOD — use textContent instead of innerHTML for text display
const name = new URLSearchParams(location.search).get("name") || "";
document.getElementById("greeting").textContent = `Hello, ${name}!`;  // no HTML parsing

// GOOD — sanitize if HTML rendering is required
import DOMPurify from "dompurify";
const safeHTML = DOMPurify.sanitize(userInput);
element.innerHTML = safeHTML;

// GOOD — postMessage with origin validation
window.addEventListener("message", (event) => {
    if (event.origin !== "https://trusted-app.example.com") return;
    document.getElementById("content").textContent = event.data;
});
```

## Checks to Generate

- Grep for `innerHTML`, `outerHTML`, `document.write(` assignments from `location.search`, `location.hash`, `document.referrer`.
- Flag `eval(`, `setTimeout(string,`, `setInterval(string,` with non-literal arguments.
- Grep for `$(selector).html(variable)` or `$(selector).append(variable)` in jQuery code.
- Flag `postMessage` listeners without `event.origin` validation.
- Grep for `location.href = userInput` without URL validation — open redirect + DOM XSS.
- Check for missing Content-Security-Policy with `unsafe-eval` restriction.
