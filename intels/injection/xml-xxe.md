---
id: injection-xxe
title: XML External Entity (XXE) Injection
severity: high
tags: [injection, xxe, xml, file-read, ssrf]
taxonomy: security/injection/xxe
references:
  - https://owasp.org/www-community/vulnerabilities/XML_External_Entity_(XXE)_Processing
  - https://cheatsheetseries.owasp.org/cheatsheets/XML_External_Entity_Prevention_Cheat_Sheet.html
---

# XML External Entity (XXE) Injection

## Description

XXE occurs when XML parsers process external entity references included in untrusted XML input. Attackers can read arbitrary files from the server filesystem, perform SSRF to internal services, cause denial of service (billion laughs), or in rare configurations achieve remote code execution.

XXE is especially common in SOAP web services, file upload endpoints (DOCX, SVG, XML configs), and any feature that parses user-supplied XML.

## Vulnerable Pattern

```python
# BAD — lxml parsing user XML with entity resolution enabled (default)
from lxml import etree

@app.post("/api/parse-xml")
async def parse_xml(request: Request):
    body = await request.body()
    tree = etree.fromstring(body)  # external entities resolved by default in some configs!
    return {"result": tree.find("name").text}

# Malicious payload:
# <?xml version="1.0"?>
# <!DOCTYPE foo [<!ENTITY xxe SYSTEM "file:///etc/passwd">]>
# <root><name>&xxe;</name></root>
```

```java
// BAD — Java SAX parser with XXE enabled (default in older Java)
DocumentBuilderFactory dbf = DocumentBuilderFactory.newInstance();
DocumentBuilder db = dbf.newDocumentBuilder();
Document doc = db.parse(userInputStream);  // XXE by default
```

## Secure Pattern

```python
# GOOD — lxml with XXE disabled
from lxml import etree

def safe_parse(xml_bytes: bytes):
    parser = etree.XMLParser(
        resolve_entities=False,
        no_network=True,
        load_dtd=False,
    )
    return etree.fromstring(xml_bytes, parser=parser)
```

```java
// GOOD — Java: disable XXE features explicitly
DocumentBuilderFactory dbf = DocumentBuilderFactory.newInstance();
dbf.setFeature("http://apache.org/xml/features/disallow-doctype-decl", true);
dbf.setFeature("http://xml.org/sax/features/external-general-entities", false);
dbf.setFeature("http://xml.org/sax/features/external-parameter-entities", false);
dbf.setXIncludeAware(false);
dbf.setExpandEntityReferences(false);
DocumentBuilder db = dbf.newDocumentBuilder();
```

## Checks to Generate

- Grep for `etree.fromstring(`, `etree.parse(` without `XMLParser(resolve_entities=False)`.
- Grep for `DocumentBuilderFactory.newInstance()` in Java without subsequent XXE feature disabling.
- Flag SVG file upload handling — SVG is XML and can contain XXE payloads.
- Grep for DOCX/XLSX/ODT processing without XXE-safe parser configuration.
- Flag `xmltodict.parse(`, `ElementTree.parse(` on user-supplied XML without entity restriction.
- Check SOAP/WSDL endpoint parsers for XXE mitigation.
