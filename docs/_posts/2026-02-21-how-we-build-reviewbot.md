---
layout: post
title: "How we're building ReviewBot: decisions and reasons"
date: 2026-02-21 09:00:00 -0000
categories: architecture
excerpt: "A walkthrough of the key architectural choices behind ReviewBot — what we picked, what we skipped building, and why."
---

The core idea behind ReviewBot is straightforward: automated security code review that accumulates context over time, so each review of a familiar codebase is better than the last. Getting there involves a bunch of decisions about tooling and architecture. This post documents those decisions and the reasoning behind each one.

---

## Knowledge management: AnythingLLM

The first real problem we hit was context cost. To review code meaningfully, an AI agent needs to understand the codebase — its structure, its patterns, its previous vulnerabilities. Feeding that in fresh on every run is expensive and wasteful.

The answer is RAG: index documents once, retrieve only what's relevant at query time. We looked at building this ourselves, decided it wasn't a good use of time, and reached for [AnythingLLM](https://anythingllm.com) instead.

It's a self-hosted platform that handles document storage, embedding, hybrid search, and workspace management. We get a REST API for programmatic access and a web UI for browsing and testing what's indexed. Each repository gets its own workspace — so context from one codebase doesn't bleed into another.

Here's what it looks like when we're exploring an indexed knowledge base:

![AnythingLLM chat interface showing a conversation about security intel]({{ '/assets/images/anythingllm-chat-example.png' | relative_url }})

The practical result: after a first full review of a repo, subsequent reviews pull relevant context from the index rather than re-processing everything. Token cost drops significantly on repeat runs.

[More detail on the KB setup →]({{ '/deep/anythingllm' | relative_url }})

---

## Pipeline orchestration: n8n

A review run isn't a single step. There's receiving the webhook, fetching repo context, querying the knowledge base, triggering agents, aggregating output, posting back to GitHub. Each step can fail. Some can run in parallel. Retry logic matters.

We use [n8n](https://n8n.io) for this. It's an open-source workflow platform — self-hosted, visual, with good webhook support and built-in error handling. The pipeline is configured there rather than hardcoded, which makes it easy to adjust behavior without a redeploy.

![n8n pipeline showing the review workflow with connected nodes]({{ '/assets/images/n8n-pipeline-example.png' | relative_url }})

The tradeoff is one more service to operate. We decided that's worth it for the flexibility, especially while the review strategy is still evolving.

[More detail on the pipeline setup →]({{ '/deep/n8n' | relative_url }})

---

## Agent execution: containerized Claude and Gemini

The actual review work happens in Claude or Gemini agents. These run in containers — each one exposes a simple HTTP endpoint, receives a job (repo, query, callback URL), runs the agent, and POSTs results back when done.

This keeps the orchestration lightweight. n8n fires a request and moves on; the container does the slow work asynchronously and calls back when finished. It also means we can run multiple agents in parallel, test new models by adding a new container, and keep each execution isolated.

[More detail on the executor design →]({{ '/deep/executors' | relative_url }})

---

## KB access for agents: mcp-anythingllm

For agents to use the knowledge base, they need a way to query it. We built a small [MCP](https://modelcontextprotocol.io) server — `mcp-anythingllm` — that exposes AnythingLLM as a set of tools the agents can call natively. The agent searches for relevant context the same way it would use any other built-in tool, without needing custom API code in each executor.

It runs inside the containers, not exposed externally. Agents get access to the KB; nothing else does.

[More detail on the MCP integration →]({{ '/deep/mcp' | relative_url }})

---

## Security intelligence: structured intel and auto-import

The knowledge base needs good content to be useful. We've started building a structured library of security intelligence — processed from papers, OWASP reports, and research into tagged, categorized markdown documents. An `ai-redteaming-intel-pack`, for example, covers prompt injection patterns, jailbreak techniques, and mitigation strategies in a form that's useful for retrieval.

We also built `kb-maintainer`: a small service that watches the `intels/` directory and keeps AnythingLLM in sync automatically. Drop a file in, it gets indexed. Update it, the index updates. Delete it, it's removed. No manual steps.

This feeds a broader design: as ReviewBot reviews repos, findings and context get added to per-repo workspaces. The knowledge base grows over time, and future reviews start with more to work with.

[More detail on the intel database and importer →]({{ '/deep/intelligence' | relative_url }})

---

The common thread across all of these decisions is trying to spend engineering time on what's unique to the problem — the review logic, the context accumulation, the intelligence layer — and use mature tools for everything else. That's the plan, anyway. We'll document how well it holds up as things develop.
