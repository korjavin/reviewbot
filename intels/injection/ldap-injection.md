---
id: injection-ldap
title: LDAP Injection
severity: high
tags: [injection, ldap, directory, authentication-bypass]
taxonomy: security/injection/ldap
references:
  - https://owasp.org/www-community/attacks/LDAP_Injection
  - https://cheatsheetseries.owasp.org/cheatsheets/LDAP_Injection_Prevention_Cheat_Sheet.html
---

# LDAP Injection

## Description

LDAP injection occurs when user input is concatenated into LDAP queries without escaping. Attackers can bypass authentication, enumerate directory entries, or modify LDAP data. Common in enterprise applications using LDAP/Active Directory for authentication.

Special LDAP characters: `* ( ) \ / NUL`

## Vulnerable Pattern

```python
# BAD — LDAP filter built with string concatenation
import ldap

def authenticate(username: str, password: str) -> bool:
    conn = ldap.initialize("ldap://internal-ad:389")
    base_dn = "dc=example,dc=com"
    filter_str = f"(&(uid={username})(userPassword={password}))"
    # Payload: username = "*)(uid=*))(|(uid=*" → bypasses auth
    result = conn.search_s(base_dn, ldap.SCOPE_SUBTREE, filter_str)
    return len(result) > 0
```

```java
// BAD — Java LDAP injection
String filter = "(&(uid=" + username + ")(objectClass=person))";
NamingEnumeration results = ctx.search("", filter, controls);
```

## Secure Pattern

```python
# GOOD — escape special characters before building filter
import ldap
import ldap.filter

def authenticate(username: str, password: str) -> bool:
    conn = ldap.initialize("ldap://internal-ad:389")
    # ldap.filter.escape_filter_chars escapes: * ( ) \ / NUL
    safe_username = ldap.filter.escape_filter_chars(username)
    safe_password = ldap.filter.escape_filter_chars(password)
    filter_str = f"(&(uid={safe_username})(userPassword={safe_password}))"
    result = conn.search_s("dc=example,dc=com", ldap.SCOPE_SUBTREE, filter_str)
    return len(result) > 0
```

```java
// GOOD — Java: use parameterized LDAP attributes
String safeUsername = LdapEncoder.filterEncode(username);
String filter = "(&(uid=" + safeUsername + ")(objectClass=person))";
```

## Checks to Generate

- Grep for LDAP filter strings built with f-strings, `.format()`, or `+` concatenation.
- Grep for `ldap.initialize(` without subsequent `escape_filter_chars` on user data.
- Flag LDAP authentication that binds with user-supplied credentials without pre-search sanitization.
- Check for LDAP DN injection — user input in Distinguished Name without DN escaping.
- Grep for `ctx.search(` in Java LDAP code using string concatenation in filter parameter.
