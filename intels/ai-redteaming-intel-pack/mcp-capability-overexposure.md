---
id: mcp-capability-overexposure
title: MCP — Capability Overexposure and Unsafe Tool Registration
severity: high
tags: [mcp, agents, tool-calling, security]
taxonomy: security/genai/mcp
source:
  name: OWASP GenAI Security Project — Vendor Evaluation Criteria for AI Red Teaming Providers & Tooling (v1.0, 2026-01-13)
---

# MCP — Capability Overexposure and Unsafe Tool Registration

## Description

In MCP-based systems, capability registries can accidentally expose tools that are:
- overpowered (admin operations),
- under-sandboxed (filesystem/network),
- incorrectly scoped to agents/users.

This increases the “agentic action space” and enables escalation.

## Vulnerable Pattern

```json
// BAD — all tools registered to all agents
{ "agent": "*", "capabilities": ["query_db", "send_email", "delete_records", "run_shell"] }
```

## Secure Pattern

```json
// GOOD — least privilege registration per agent role
{ "agent": "support_bot", "capabilities": ["search_kb", "create_ticket"] }
{ "agent": "ops_bot", "capabilities": ["query_db_readonly", "run_job_dryrun"] }
```

Hardening:
- Capability review gates (code review + security sign-off).
- Sandboxed execution for destructive tools.
- Capability provenance + traceability in logs.

## Checks to Generate

- Detect wildcard capability grants (`agent="*"` / `capabilities=["*"]`).
- Flag tools with write/delete/shell/network privileges exposed to user-facing agents.
- Require “dry-run/mock” for destructive tools in non-prod testing.
