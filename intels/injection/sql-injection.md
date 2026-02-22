---
id: injection-sql
title: SQL Injection — Classic, Blind, and Out-of-Band
severity: critical
tags: [injection, sql, database, sqli]
taxonomy: security/injection/sql
references:
  - https://owasp.org/www-community/attacks/SQL_Injection
  - https://cheatsheetseries.owasp.org/cheatsheets/SQL_Injection_Prevention_Cheat_Sheet.html
  - https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/07-Input_Validation_Testing/05-Testing_for_SQL_Injection
---

# SQL Injection — Classic, Blind, and Out-of-Band

## Description

SQL injection is one of the oldest and most impactful web vulnerabilities. Unsanitized user input concatenated into SQL queries allows attackers to read/modify any data, bypass authentication, execute stored procedures, and in some configurations achieve OS command execution (`xp_cmdshell`, `LOAD_FILE`).

Types:
- **Classic / In-band**: Error-based or UNION-based — attacker sees results directly
- **Blind Boolean**: Infers data from true/false responses
- **Blind Time-based**: Uses `SLEEP()`, `WAITFOR DELAY` to infer data
- **Out-of-band**: Exfiltrates via DNS/HTTP (e.g., MySQL `LOAD_FILE` + UNC path)

## Vulnerable Pattern

```python
# BAD — string formatting in SQL
username = request.form["username"]
password = request.form["password"]
query = f"SELECT * FROM users WHERE username='{username}' AND password='{password}'"
cursor.execute(query)
# Payload: username = ' OR '1'='1' -- (bypasses auth)
# Payload: username = '; DROP TABLE users; -- (destructive)
```

```java
// BAD — Java string concatenation
String query = "SELECT * FROM orders WHERE id = " + orderId;
Statement stmt = conn.createStatement();
ResultSet rs = stmt.executeQuery(query);
```

## Secure Pattern

```python
# GOOD — parameterized queries (Python)
cursor.execute(
    "SELECT * FROM users WHERE username = %s AND password_hash = %s",
    (username, hash_password(password))
)

# GOOD — SQLAlchemy ORM (auto-parameterized)
user = session.query(User).filter(User.username == username).first()
```

```java
// GOOD — PreparedStatement (Java)
PreparedStatement stmt = conn.prepareStatement(
    "SELECT * FROM orders WHERE id = ?"
);
stmt.setInt(1, Integer.parseInt(orderId));  // type-safe, parameterized
```

## Checks to Generate

- Grep for SQL string concatenation: `f"SELECT.*{`, `"WHERE.*" + `, `% ` followed by user variable.
- Flag raw `cursor.execute(query)` where `query` is built from f-strings or `.format()`.
- Flag `.raw(f"`, `.extra(where=f"`, `Manager.raw(` in Django ORM with f-strings.
- Grep for `Statement` (not `PreparedStatement`) in Java JDBC code.
- Check for SQLi in ORDER BY clauses — parameterization doesn't work for column names; use allowlisting.
