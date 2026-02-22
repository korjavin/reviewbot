---
id: owasp-api-a07-ssrf
title: OWASP API A07:2023 — Server-Side Request Forgery (API)
severity: high
tags: [owasp-api-top10, ssrf, cloud, webhook, api]
taxonomy: security/api/ssrf
references:
  - https://owasp.org/API-Security/editions/2023/en/0xa7-server-side-request-forgery/
---

# OWASP API A07:2023 — Server-Side Request Forgery (API)

## Description

In API contexts, SSRF frequently appears in webhook registration, image/URL preview features, PDF generation services, and any API endpoint that fetches remote content on behalf of the user. APIs expose more such features and often lack network egress controls.

## Vulnerable Pattern

```javascript
// BAD — API renders remote URL content for user
app.post("/api/render-pdf", async (req, res) => {
    const { url } = req.body;
    const pdf = await pdfGenerator.fromURL(url);  // attacker: http://localhost:8080/admin
    res.send(pdf);
});

// BAD — avatar fetch without URL validation
app.put("/api/profile/avatar", async (req, res) => {
    const { avatarUrl } = req.body;
    const imageData = await fetch(avatarUrl);  // attacker: http://169.254.169.254/latest/meta-data/
    await saveAvatar(req.user.id, imageData);
});
```

## Secure Pattern

```javascript
const { URL } = require("url");
const dns = require("dns").promises;

async function isSafeUrl(rawUrl) {
    let parsed;
    try { parsed = new URL(rawUrl); } catch { return false; }
    if (!["https:"].includes(parsed.protocol)) return false;
    const addresses = await dns.resolve4(parsed.hostname);
    for (const addr of addresses) {
        if (isPrivateIP(addr)) return false;  // RFC1918 + link-local check
    }
    return true;
}

app.post("/api/render-pdf", async (req, res) => {
    const { url } = req.body;
    if (!await isSafeUrl(url)) return res.status(400).json({ error: "URL not allowed" });
    const pdf = await pdfGenerator.fromURL(url);
    res.send(pdf);
});
```

## Checks to Generate

- Flag webhook URL fields in API endpoints without domain allowlist validation.
- Grep for PDF generation libraries (`puppeteer`, `wkhtmltopdf`, `weasyprint`) accepting user URLs.
- Flag image/avatar fetch endpoints that accept external URLs without validation.
- Check for missing egress firewall rules — SSRF is easier to exploit if the server can reach internal services.
- Flag `fetch(url)` / `requests.get(url)` / `http.get(url)` where `url` comes from request body/params.
