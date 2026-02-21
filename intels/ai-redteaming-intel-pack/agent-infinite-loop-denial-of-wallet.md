---
id: agent-denial-of-wallet-infinite-loop
title: Agents — Infinite Loops, Resource Exhaustion, and Denial of Wallet
severity: high
tags: [agents, reliability, cost, security]
taxonomy: security/genai/resource-abuse
source:
  name: OWASP GenAI Security Project — Vendor Evaluation Criteria for AI Red Teaming Providers & Tooling (v1.0, 2026-01-13)
---

# Agents — Infinite Loops, Resource Exhaustion, and Denial of Wallet

## Description

Agents that plan + call tools can be forced into:
- infinite/very long loops,
- excessive tool calls,
- repeated retries/timeouts with unsafe fallbacks,

leading to cost spikes, degraded availability, or unintended actions.

## Vulnerable Pattern

```ts
// BAD — unbounded planning loop
while (!done) {
  const step = await llm.plan(state);
  state = await execute(step); // may never converge
}
```

## Secure Pattern

```ts
// GOOD — budgets + circuit breakers + safe fallback
const MAX_TURNS = 12;
const MAX_TOOL_CALLS = 20;

for (let turn = 0; turn < MAX_TURNS; turn++) {
  if (state.toolCalls >= MAX_TOOL_CALLS) throw new Error("budget exceeded");
  const step = await llm.plan(state, { remainingBudget: MAX_TURNS - turn });
  state = await execute(step);
  if (state.done) break;
}

if (!state.done) return safeFallbackAnswer();
```

Controls:
- token/turn budgets, tool-call rate limits,
- timeout handling that fails closed,
- monitoring + alerts for spikes.

## Checks to Generate

- Flag `while(true)` / unbounded loops around agent planning.
- Flag missing budgets for tool calls and tokens.
- Test:
  - prompts designed to keep the agent “searching forever”,
  - induce tool timeouts and verify safe behavior.
