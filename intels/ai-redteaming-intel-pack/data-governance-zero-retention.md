---
id: genai-data-governance-retention
title: GenAI — Data Governance Failures (Retention, Isolation, Third-Party Sharing)
severity: high
tags: [privacy, governance, security]
taxonomy: governance/genai/data
source:
  name: OWASP GenAI Security Project — Vendor Evaluation Criteria for AI Red Teaming Providers & Tooling (v1.0, 2026-01-13)
---

# GenAI — Data Governance Failures (Retention, Isolation, Third-Party Sharing)

## Description

GenAI systems often store prompts, logs, tool outputs, and chat history. Weak governance can lead to:
- over-retention of sensitive data,
- mixing prod data into shared test environments,
- unclear sharing with third-party model providers,
- insufficient access controls to logs/tool data.

## Vulnerable Pattern

```yaml
# BAD — retain everything indefinitely
logging:
  store_raw_prompts: true
  store_tool_outputs: true
  retention_days: 0  # means forever
```

## Secure Pattern

```yaml
# GOOD — minimize, encrypt, restrict, delete
logging:
  store_raw_prompts: false
  store_tool_outputs: "redacted"
  retention_days: 30
security:
  encrypt_at_rest: true
  access_control: "least_privilege"
third_party:
  model_provider_zero_retention: true
  disclose_subprocessors: true
```

## Checks to Generate

- Flag configs with indefinite retention / raw prompt storage in prod.
- Verify tenant isolation for log storage and vector indexes.
- Verify disclosure/allowlist of third-party subprocessors for prompts/logs.
