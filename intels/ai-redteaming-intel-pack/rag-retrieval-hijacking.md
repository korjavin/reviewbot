---
id: rag-retrieval-hijacking
title: RAG — Retrieval Override and Semantic Hijacking
severity: high
tags: [rag, retrieval, security]
taxonomy: security/genai/rag
source:
  name: OWASP GenAI Security Project — Vendor Evaluation Criteria for AI Red Teaming Providers & Tooling (v1.0, 2026-01-13)
---

# RAG — Retrieval Override and Semantic Hijacking

## Description

Attackers can craft queries or documents to manipulate embedding similarity and reranking so that malicious or irrelevant content is retrieved preferentially (“semantic hijacking”). This can cause wrong answers, policy bypass, or downstream tool misuse.

## Vulnerable Pattern

```python
# BAD — no query constraints, no source trust, no retrieval auditing
docs = vector_db.similarity_search(user_query, k=8)
answer = llm.answer(question=user_query, context="\n".join(d.page_content for d in docs))
```

## Secure Pattern

```python
# BETTER — retrieval constraints + trust tiers + auditing
docs = vector_db.similarity_search(
    sanitize_query(user_query),
    k=8,
    filters={"source_trust": {"$gte": 2}},  # e.g., only approved corp sources
)

docs = rerank_with_rules(docs, user_query)
docs = dedupe_and_cap_per_source(docs)

answer = llm.answer(
    question=user_query,
    context=annotate_with_provenance(docs),
)
log_retrieval_trace(user_query, docs)
```

Mitigations:
- Source allowlists / signed corpora.
- Per-source caps; deduping; freshness constraints.
- Retrieval audit logs and replay for debugging.
- Evaluate RSR@k (Retrieval Success Rate under adversarial conditions) where applicable.

## Checks to Generate

- Flag retrieval without filters (no trust constraints) when system handles sensitive/internal domains.
- Flag missing retrieval logging/tracing (`log_retrieval_trace`, `langfuse`, etc.).
- Add adversarial tests:
  - query stuffing, synonym flooding, long “instructional” queries,
  - doc injection in staging index to see if it becomes top-k retrieved.
