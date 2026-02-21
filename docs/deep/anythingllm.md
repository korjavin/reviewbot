---
layout: post
title: "Knowledge management in depth: AnythingLLM"
permalink: /deep/anythingllm
---

[← Back to main post]({{ '/blog/2026/02/21/how-we-build-reviewbot/' | relative_url }})

## Why not build a RAG system ourselves?

The components of a RAG system aren't complicated individually — document chunking, embedding generation, vector storage, retrieval ranking, a storage layer — but putting them together and keeping them running is a real maintenance burden. There are also subtler decisions that take time to get right: chunk size, overlap, hybrid vs. pure vector search, re-ranking strategies. Getting good retrieval quality is an ongoing tuning problem, not a one-time implementation.

We made the call early that this wasn't where we wanted to spend time. The value we're building is in the review intelligence, not the retrieval infrastructure.

## What AnythingLLM gives us

[AnythingLLM](https://anythingllm.com) is a self-hosted RAG platform. We run it in a container alongside the rest of the stack.

The key features we use:

**Workspaces.** Each repository gets its own workspace — a separate embedding namespace. Context from one codebase doesn't affect retrieval for another. We also maintain a shared `intels` workspace with universal security knowledge (OWASP, papers, vulnerability patterns) that all agents can query regardless of which repo they're reviewing.

**Hybrid search.** Retrieval uses both vector similarity and full-text matching. For code-related queries, this tends to outperform pure semantic search — exact term matches matter when you're looking for a specific library or function name.

**REST API.** All document management and querying is available programmatically. Our `kb-maintainer` service uses this to sync documents, and our MCP server uses it to serve queries to agents.

**Web UI.** We can browse what's indexed, run test queries, and chat with the knowledge base directly. This is useful for verifying that the indexed content is actually useful before relying on it in automated reviews.

## Workspace structure

```
AnythingLLM
├── workspace: intels
│   ├── OWASP API Security Top 10
│   ├── AI red-teaming intel pack
│   ├── JWT vulnerability patterns
│   └── ... (universal security knowledge)
│
├── workspace: repo-abc
│   ├── Previous review findings
│   ├── Architecture notes
│   └── Dependency analysis
│
└── workspace: repo-xyz
    └── (separate, isolated)
```

The per-repo workspaces grow as ReviewBot reviews the codebase repeatedly. Earlier findings become part of the context for later reviews.

## Embedding and retrieval

We use the default embedding model that ships with AnythingLLM. For security-specific content, we may eventually switch to a code-tuned model, but the default has been adequate for the intel documents we've indexed so far.

Documents are chunked automatically. Chunk size and overlap can be configured per workspace — we haven't needed to tune this yet, but it's available if retrieval quality becomes an issue.

## Self-hosting tradeoffs

Running AnythingLLM ourselves means we control the data — nothing goes to an external service. It also means we operate the service. So far this has been low-maintenance: it's a single Docker image with a persistent volume for the database.

The main operational consideration is the embedding model, which runs locally and needs enough CPU (or GPU) to be responsive. On the hardware we're using, embedding new documents takes a few seconds per document — acceptable for our current volume.
