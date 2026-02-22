---
id: javascript-prototype-pollution
title: Prototype Pollution (JavaScript / Node.js)
severity: high
tags: [javascript, nodejs, prototype-pollution, rce, object-injection, lodash]
taxonomy: security/javascript/prototype-pollution
references:
  - https://owasp.org/www-community/vulnerabilities/Prototype_Pollution
  - https://portswigger.net/web-security/prototype-pollution
  - https://github.com/advisories?query=prototype+pollution
---

# Prototype Pollution (JavaScript / Node.js)

## Description

Prototype pollution occurs when attacker-controlled keys (like `__proto__`, `constructor`, `prototype`) are merged into a JavaScript object, poisoning `Object.prototype`. Every object in the process inherits the injected properties, leading to:

- **Application logic bypass**: checking `if (user.isAdmin)` returns `true` for all objects
- **Remote Code Execution**: via gadget chains in template engines (Handlebars, Pug, EJS), `child_process.exec`, or `vm.runInNewContext`
- **Denial of Service**: overwriting `toString`, `valueOf`, or `length`

High-profile affected packages: lodash `_.merge()`, `_.set()`, `_.defaultsDeep()`, jQuery `$.extend(deep=true)`, and dozens of deep merge utilities.

## Vulnerable Pattern

```javascript
// BAD — unsafe deep merge (classic lodash <4.17.21)
const _ = require("lodash");

function updateUserSettings(user, newSettings) {
    return _.merge(user, newSettings);
    // Payload: newSettings = { "__proto__": { "isAdmin": true } }
    // → Object.prototype.isAdmin = true
    // → ALL objects now have .isAdmin === true
}

// BAD — recursive merge without key validation
function deepMerge(target, source) {
    for (const key of Object.keys(source)) {
        if (typeof source[key] === "object") {
            target[key] = target[key] || {};
            deepMerge(target[key], source[key]);  // key "__proto__" not blocked!
        } else {
            target[key] = source[key];
        }
    }
    return target;
}

// BAD — property access via bracket notation with user-controlled key
function setProperty(obj, key, value) {
    obj[key] = value;  // key = "__proto__" pollutes prototype
}
```

```javascript
// Downstream impact of polluted prototype:
if (user.isAdmin) {  // true for ALL users after pollution!
    return adminPanel();
}

// RCE via Handlebars gadget:
const template = Handlebars.compile("{{foo}}")({ foo: "bar" });
// If __proto__.pendingContent is set, Handlebars executes it
```

## Secure Pattern

```javascript
// GOOD — use safe merge that blocks prototype keys
function safeMerge(target, source) {
    const forbidden = new Set(["__proto__", "constructor", "prototype"]);
    for (const key of Object.keys(source)) {
        if (forbidden.has(key)) continue;  // block polluting keys
        if (typeof source[key] === "object" && source[key] !== null) {
            target[key] = target[key] || Object.create(null);  // null-prototype object
            safeMerge(target[key], source[key]);
        } else {
            target[key] = source[key];
        }
    }
    return target;
}

// GOOD — use Object.create(null) for data objects from user input
const userConfig = Object.create(null);  // no prototype — immune to pollution
Object.assign(userConfig, JSON.parse(userInput));

// GOOD — use lodash >=4.17.21 (patched) or structured-clone
const merged = structuredClone(target);  // deep clone without prototype chain
Object.assign(merged, safeSource);

// GOOD — validate keys before property access
function setProperty(obj, key, value) {
    if (["__proto__", "constructor", "prototype"].includes(key)) {
        throw new Error("Forbidden key");
    }
    obj[key] = value;
}
```

## Checks to Generate

- Grep for `_.merge(`, `_.defaultsDeep(`, `_.set(` with user-controlled input — check lodash version >= 4.17.21.
- Grep for recursive merge functions without `__proto__`/`constructor`/`prototype` key exclusion.
- Flag `obj[userKey] = value` patterns where `userKey` comes from request body or query params.
- Grep for `JSON.parse(input)` results passed directly to `Object.assign`, `_.merge`, or spread into existing objects.
- Flag old versions of `qs`, `ajv`, `yargs-parser`, `node-forge` — all had prototype pollution CVEs.
- Grep for `hasOwnProperty` checks without `Object.prototype.hasOwnProperty.call(obj, key)` — polluted `hasOwnProperty` bypasses the check.
- Check for `npm audit` output mentioning prototype pollution in dependency tree.
