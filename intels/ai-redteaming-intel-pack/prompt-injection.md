---
id: llm-prompt-injection
title: LLM App — Prompt Injection (Direct) and Instruction Override
severity: high
tags: [llm, prompt-injection, security]
taxonomy: security/genai/prompt-injection
source:
  name: OWASP GenAI Security Project — Vendor Evaluation Criteria for AI Red Teaming Providers & Tooling (v1.0, 2026-01-13)
---

# LLM App — Prompt Injection (Direct) and Instruction Override

## Description

User-supplied text can contain malicious instructions that override the system/developer intent (policy, tool-use rules, safety constraints) and cause the model/agent to reveal secrets, bypass guardrails, or execute unsafe actions.

This is especially dangerous when the application concatenates user input into a single prompt without robust instruction hierarchy, delimiters, and enforcement.

## Vulnerable Pattern

```python
# BAD — user text directly appended to instructions
SYSTEM = "You are a support assistant. Never reveal internal info. Use tools safely."
user_msg = request.json["message"]

prompt = SYSTEM + "\nUser: " + user_msg
answer = llm.generate(prompt)
```

Typical symptoms:
- “Ignore previous instructions…”
- “Reveal the system prompt…”
- “Print internal configuration / API keys…”

## Secure Pattern

```python
# BETTER — structured messages + explicit untrusted delimiters + policy enforcement
messages = [
  {"role": "system", "content": "You are a support assistant. Follow policy. Never reveal secrets."},
  {"role": "user", "content": user_msg},
]

answer = llm.chat(messages, safety_settings=POLICY)

# If answer triggers policy, block/transform + log evidence for review.
```

Additional hardening:
- Separate *policy* from *content* (do not mingle untrusted text inside system policy).
- Use model/tooling that supports system/developer messages and tool-call constraints.
- Add allowlists for actions; block requests for secrets and privileged operations.

## Checks to Generate

Static / semi-static checks:
- Flag code that concatenates `system` + `user` into a single prompt string.
  - Patterns: `system.*\+.*user`, f-strings mixing system/user, `.format(system=..., user=...)`.
- Flag templates where user input is injected into “instructions” section (e.g., before “RULES:”).
- Flag “system prompt debug” routes in prod (e.g., `/debug/prompt`, `print(system_prompt)`).

Dynamic tests (agent harness):
- Attempt instruction override payloads and verify:
  - system prompt is not disclosed,
  - sensitive data is not revealed,
  - tool actions are not triggered without authorization.
