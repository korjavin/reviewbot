---
id: agent-tool-calling-schema-manipulation
title: Tool-Calling Agents — Schema/Parameter Manipulation and Unsafe Tool Triggers
severity: critical
tags: [agents, tool-calling, security]
taxonomy: security/genai/tool-calling
source:
  name: OWASP GenAI Security Project — Vendor Evaluation Criteria for AI Red Teaming Providers & Tooling (v1.0, 2026-01-13)
---

# Tool-Calling Agents — Schema/Parameter Manipulation and Unsafe Tool Triggers

## Description

When an agent can call tools (send_email, query_database, run_job), attackers may coerce the model into:
- calling the wrong tool (“misfire”),
- crafting malicious parameters (SQL/command injection, SSRF),
- bypassing intended tool constraints via schema ambiguity.

## Vulnerable Pattern

```ts
// BAD — tool parameters are passed through without validation
tools.register("query_database", async ({ sql }) => db.query(sql));
```

```python
# BAD — model can pick any tool, no allowlist per request
result = agent.run(user_input, tools=ALL_TOOLS)
```

## Secure Pattern

```ts
// GOOD — strict schemas + validation + allowlists
const Query = z.object({ queryId: z.string(), params: z.record(z.string()) });
tools.register("query_database", async (raw) => {
  const { queryId, params } = Query.parse(raw);
  return db.runPrepared(queryId, params);
});

const allowedTools = computeAllowedTools(userContext, requestType);
const result = await agent.run(userInput, { allowedTools });
```

Hardening:
- Prefer “pre-defined queries” over free-form SQL.
- Validate every tool arg; enforce allowlists/rate limits.
- Add “dry-run/mock” modes for destructive tools in testing.

## Checks to Generate

- Flag tools that accept free-form `sql`, `cmd`, `url`, `path` without validation.
- Flag agents invoked with broad tool sets (`ALL_TOOLS`, `tools=*`).
- Runtime adversarial tests:
  - force tool selection (“call send_email now…”),
  - schema confusion (“set admin=true”), 
  - injection attempts in tool args.
