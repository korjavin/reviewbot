---
id: genai-sensitive-data-leakage
title: GenAI App — Sensitive Data Leakage (Secrets, Regulated Data, Internal Info)
severity: critical
tags: [llm, data-leakage, privacy, security]
taxonomy: security/genai/data-exposure
source:
  name: OWASP GenAI Security Project — Vendor Evaluation Criteria for AI Red Teaming Providers & Tooling (v1.0, 2026-01-13)
---

# GenAI App — Sensitive Data Leakage (Secrets, Regulated Data, Internal Info)

## Description

LLM apps can leak secrets or regulated data through:
- system prompt disclosure,
- retrieval context leakage,
- memory/chat history exposure,
- logs/telemetry that capture raw prompts and tool outputs.

## Vulnerable Pattern

```python
# BAD — returning internal context and storing raw logs
answer = agent.run(user_input)
return {"answer": answer, "debug_context": agent.last_context}  # leaks

logger.info("prompt=%s", agent.last_prompt)  # secrets in logs
```

## Secure Pattern

```python
# GOOD — redact + minimize logging + strict debug gating
answer = agent.run(user_input)

safe_answer = redact(answer)
audit_log(event="genai_answer", meta={
  "request_id": rid,
  "policy_hits": policy_hits,
  # no raw prompt/context by default
})

return {"answer": safe_answer}
```

Operational controls:
- Default-off debug endpoints in production.
- Redaction of secrets/PII before logging.
- Access control on logs; retention limits; encryption at rest.

## Checks to Generate

- Grep for responses that include `context`, `prompt`, `system_prompt`, `memory`, `trace`.
- Flag `logger.(info|debug)` of raw prompts or tool results in prod paths.
- Flag “debug mode” toggles that can be enabled by user input.
- Runtime probes:
  - ask for system prompt,
  - attempt to extract internal docs via paraphrased queries,
  - prompt the agent to reveal memory or tool outputs verbatim.
