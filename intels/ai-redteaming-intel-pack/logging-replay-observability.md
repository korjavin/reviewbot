---
id: genai-logging-replay-observability
title: GenAI Systems — Missing Logs, Traces, and Deterministic Replay (Poor Observability)
severity: medium
tags: [observability, agents, rag]
taxonomy: reliability/genai/observability
source:
  name: OWASP GenAI Security Project — Vendor Evaluation Criteria for AI Red Teaming Providers & Tooling (v1.0, 2026-01-13)
---

# GenAI Systems — Missing Logs, Traces, and Deterministic Replay (Poor Observability)

## Description

Without full traces (messages, tool calls, retrieval, state transitions), security findings are hard to reproduce and fix. For complex multi-turn or multi-agent failures, deterministic replay is critical to debug and validate mitigations.

## Vulnerable Pattern

```python
# BAD — only final answer logged
logger.info("answer=%s", answer)
```

## Secure Pattern

```python
# GOOD — structured tracing with redaction and replay IDs
trace_id = new_trace_id()
trace.record(trace_id, event="user_message", data=redact(user_msg))
trace.record(trace_id, event="retrieval", data=redact(retrieval_docs_meta))
trace.record(trace_id, event="tool_call", data=redact(tool_call))
trace.record(trace_id, event="model_output", data=redact(answer))

return {"answer": answer, "trace_id": trace_id}
```

Notes:
- store enough to reproduce in a safe environment (seed, prompts, tool-call inputs, versions),
- do not store secrets/PII unredacted.

## Checks to Generate

- Flag systems that log only outputs but not tool-call/retrieval traces.
- Flag missing correlation IDs (`trace_id`, `request_id`) across components.
- Ensure replay harness exists (capture versions, seeds, configs) and is used in CI regression tests.
