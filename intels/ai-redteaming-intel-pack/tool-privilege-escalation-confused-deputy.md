---
id: agent-privilege-escalation-confused-deputy
title: Tool-Using Agents — Privilege Escalation and Confused Deputy
severity: critical
tags: [agents, authorization, tool-calling, security]
taxonomy: security/genai/authorization
source:
  name: OWASP GenAI Security Project — Vendor Evaluation Criteria for AI Red Teaming Providers & Tooling (v1.0, 2026-01-13)
---

# Tool-Using Agents — Privilege Escalation and Confused Deputy

## Description

A “User-level” agent can be manipulated to execute “Admin-level” actions through tools, especially when tools rely on the agent’s service credentials instead of the end-user’s authorization context.

Common symptoms:
- agent performs writes/deletes on behalf of unauthorized users,
- agent bypasses approval flows (“human-in-the-loop” bypass),
- tool ACLs do not bind to the requester identity.

## Vulnerable Pattern

```python
# BAD — tool uses service account, ignores end-user authorization
def delete_user(user_id: str):
    return admin_api.delete_user(user_id)  # always admin

agent.tools.register("delete_user", delete_user)
```

## Secure Pattern

```python
# GOOD — enforce authorization per action
def delete_user(ctx, user_id: str):
    if not ctx.user.can("admin:delete_user"):
        raise PermissionError("not authorized")
    return admin_api.delete_user(user_id)

agent.tools.register("delete_user", lambda args: delete_user(request_ctx, **args))
```

Hardening:
- Bind tool calls to end-user identity (ABAC/RBAC).
- Separate “read” vs “write” tools; default deny.
- Require explicit approvals for high-impact actions (HITL), and test bypass resistance.

## Checks to Generate

- Identify tools calling privileged APIs without checking `ctx.user` / authz.
- Flag use of long-lived service tokens in tool executors.
- Test cases:
  - “Vertical privilege escalation” attempts (user prompts to perform admin tasks),
  - multi-step chains (search → draft impersonation → send).
