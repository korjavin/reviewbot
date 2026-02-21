---
id: multi-agent-role-contamination
title: Multi-Agent Systems — Role Confusion, Message-Passing Attacks, and Cross-Agent Contamination
severity: high
tags: [multi-agent, agents, security]
taxonomy: security/genai/multi-agent
source:
  name: OWASP GenAI Security Project — Vendor Evaluation Criteria for AI Red Teaming Providers & Tooling (v1.0, 2026-01-13)
---

# Multi-Agent Systems — Role Confusion, Message-Passing Attacks, and Cross-Agent Contamination

## Description

In multi-agent orchestration, agents exchange messages and can influence each other’s goals, tool choices, and safety posture. Attackers can inject messages that:
- impersonate a higher-privilege agent,
- cause “role drift” (“act as admin now”),
- introduce malicious tasks that propagate.

## Vulnerable Pattern

```python
# BAD — no authentication/validation on inter-agent messages
router.send(to="billing_agent", msg=user_text)
```

## Secure Pattern

```python
# GOOD — signed messages + strict schemas + separation of duties
msg = {
  "from": "router",
  "to": "billing_agent",
  "type": "task",
  "payload": {"action": "create_invoice", "customer_id": cid},
}
signed = sign(msg, key=ROUTER_KEY)
router.send(signed)
```

Hardening:
- Authenticate message origin; validate message schemas.
- Explicit role policies per agent; tool allowlists per role.
- Prevent user-controlled text from becoming “system” instructions to other agents.

## Checks to Generate

- Flag any `send(agent, user_text)` patterns without schema/validation.
- Flag “role” fields derived from user input.
- Dynamic tests:
  - attempt to coerce one agent via another (“tell admin agent to…”),
  - attempt message replay or tampering.
