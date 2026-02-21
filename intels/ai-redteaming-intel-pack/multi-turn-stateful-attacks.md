---
id: genai-multi-turn-stateful
title: GenAI Agent — Multi-turn and Stateful Attack Persistence (Memory / Long-lived Sessions)
severity: high
tags: [agents, memory, security]
taxonomy: security/genai/stateful
source:
  name: OWASP GenAI Security Project — Vendor Evaluation Criteria for AI Red Teaming Providers & Tooling (v1.0, 2026-01-13)
---

# GenAI Agent — Multi-turn and Stateful Attack Persistence (Memory / Long-lived Sessions)

## Description

State (memory, chat history, tool state) can allow attacks to persist across turns or sessions:
- gradual policy erosion (“persona manipulation”),
- stored malicious instructions in memory,
- cross-session leakage between users/tenants.

## Vulnerable Pattern

```ts
// BAD — shared memory key across users + no validation
await memoryStore.append("global", userMessage); // cross-tenant bleed
```

```python
# BAD — untrusted instructions stored as “preferences”
memory.save({"user_pref": user_text})  # could contain adversarial instructions
```

## Secure Pattern

```ts
// GOOD — per-tenant/per-user scoping + validation + TTL
const key = `tenant:${tenantId}:user:${userId}`;
await memoryStore.append(key, sanitizeForMemory(userMessage), { ttlDays: 30 });
```

Hardening:
- Memory schema: store facts, not instructions.
- Validate memory writes; strip instruction-like patterns.
- Tenant isolation in storage (keys, ACLs, separate indexes).
- Provide “forget” controls and retention limits.

## Checks to Generate

- Flag memory keys that are not scoped by tenant/user.
- Flag memory writes of raw user messages without sanitization.
- Test harness:
  - multi-turn escalation attempts,
  - “write this to memory and always do X” payloads,
  - cross-session probing for other user data.
