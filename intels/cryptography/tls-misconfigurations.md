---
id: crypto-tls-misconfigurations
title: TLS/SSL Misconfigurations
severity: high
tags: [cryptography, tls, ssl, certificate, mitm, https]
taxonomy: security/cryptography/tls
references:
  - https://owasp.org/www-community/controls/Transport_Layer_Protection_Cheat_Sheet
  - https://cheatsheetseries.owasp.org/cheatsheets/Transport_Layer_Security_Cheat_Sheet.html
  - https://ssl-config.mozilla.org/
---

# TLS/SSL Misconfigurations

## Description

TLS misconfigurations allow network attackers to intercept or manipulate encrypted traffic. Common issues: accepting self-signed certificates, disabling certificate verification, supporting deprecated protocol versions (TLS 1.0/1.1, SSLv3), weak cipher suites, and missing HSTS headers.

## Vulnerable Pattern

```python
# BAD — certificate verification disabled
import requests
response = requests.get("https://api.example.com", verify=False)  # MITM risk

# BAD — custom SSL context disabling verification
import ssl
ctx = ssl.create_default_context()
ctx.check_hostname = False
ctx.verify_mode = ssl.CERT_NONE
```

```javascript
// BAD — Node.js: rejecting certificate validation
const https = require("https");
const agent = new https.Agent({ rejectUnauthorized: false });
fetch(url, { agent });  // MITM attack possible
```

```nginx
# BAD — nginx supporting old TLS versions
ssl_protocols SSLv3 TLSv1 TLSv1.1 TLSv1.2;  # SSLv3=POODLE, TLS1.0=BEAST/POODLE
ssl_ciphers ALL;  # includes weak ciphers (RC4, DES, EXPORT)
```

## Secure Pattern

```nginx
# GOOD — Mozilla Intermediate configuration
ssl_protocols TLSv1.2 TLSv1.3;
ssl_ciphers ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384;
ssl_prefer_server_ciphers off;
ssl_session_timeout 1d;
ssl_session_cache shared:SSL:10m;
ssl_session_tickets off;
add_header Strict-Transport-Security "max-age=63072000; includeSubDomains; preload" always;
add_header X-Frame-Options DENY;
add_header X-Content-Type-Options nosniff;
```

```python
# GOOD — always verify certificates
response = requests.get("https://api.example.com", verify=True)  # default
response = requests.get("https://api.example.com", verify="/path/to/ca-bundle.crt")
```

## Checks to Generate

- Grep for `verify=False` in Python requests/httpx calls.
- Grep for `rejectUnauthorized: false` in Node.js HTTPS agents.
- Grep for `check_hostname = False` or `CERT_NONE` in Python ssl context.
- Flag `ssl_protocols` including `SSLv3`, `TLSv1`, `TLSv1.1` in nginx/Apache config.
- Flag missing `Strict-Transport-Security` header on HTTPS responses.
- Grep for `InsecureRequestWarning` suppression (`urllib3.disable_warnings()`).
