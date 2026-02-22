---
id: injection-xpath
title: XPath Injection
severity: high
tags: [injection, xpath, xml, authentication-bypass]
taxonomy: security/injection/xpath
references:
  - https://owasp.org/www-community/attacks/XPATH_Injection
  - https://owasp.org/www-project-web-security-testing-guide/latest/4-Web_Application_Security_Testing/07-Input_Validation_Testing/09-Testing_for_XPath_Injection
---

# XPath Injection

## Description

XPath injection affects applications using XPath queries against XML data stores or XML-based authentication systems. Like SQL injection, unsanitized user input can manipulate XPath expressions to bypass authentication or extract arbitrary data from the XML document.

## Vulnerable Pattern

```python
# BAD — XPath query with string concatenation
from lxml import etree

def authenticate_user(username: str, password: str) -> bool:
    tree = etree.parse("users.xml")
    query = f"//user[name='{username}' and password='{password}']"
    result = tree.xpath(query)
    # Payload: username = "' or '1'='1"  → //user[name='' or '1'='1' and password='...']
    return len(result) > 0
```

```java
// BAD — Java XPath injection
String xpath = "//employee[@id='" + userId + "']";
NodeList nodes = (NodeList) xpathExpr.evaluate(xpath, doc, XPathConstants.NODESET);
```

## Secure Pattern

```python
# GOOD — parameterized XPath using XPath variables
from lxml import etree

def authenticate_user(username: str, password: str) -> bool:
    tree = etree.parse("users.xml")
    # Use XPath variables to separate query from data
    result = tree.xpath(
        "//user[name=$uname and password=$pwd]",
        uname=username,
        pwd=password
    )
    return len(result) > 0
```

```java
// GOOD — use typed XPath evaluation with parameter binding
XPathVariableResolver resolver = new SafeVariableResolver(userId);
xpath.setXPathVariableResolver(resolver);
String safeXpath = "//employee[@id=$userId]";
```

## Checks to Generate

- Grep for XPath expressions using f-strings, `.format()`, or `+` with user input.
- Flag `tree.xpath(f"`, `xpath.evaluate("` + user variable concatenation.
- Check for XML-based authentication flows using XPath without parameterization.
- Grep for `//` XPath prefix combined with string concatenation of user data.
