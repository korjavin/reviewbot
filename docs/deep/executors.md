---
layout: post
title: "Agent executors in depth: containerized Claude and Gemini"
permalink: /deep/executors
---

[← Back to main post]({{ '/blog/2026/02/21/how-we-build-reviewbot/' | relative_url }})

## The design constraint

Agent reviews take time — typically several minutes for meaningful code exploration. The orchestrator (n8n) shouldn't block waiting for that. We also need to run multiple agents concurrently and be able to add new models without reworking the pipeline.

The design constraint we settled on: each agent implementation must expose a simple, consistent HTTP interface, run in a container, and communicate results asynchronously via callback.

## How an executor works

An executor is a container that wraps a specific model. It receives job requests over HTTP:

```json
{
  "repo_url": "https://github.com/org/repo",
  "ref": "pull/42/head",
  "query": "Review for authentication and authorization vulnerabilities",
  "kb_workspace": "repo-abc",
  "callback_url": "https://n8n.internal/webhook/review-results"
}
```

It starts the agent with the appropriate tools configured (GitHub access, MCP connection to AnythingLLM), lets the agent do its work, and when the agent finishes, POSTs the findings to the callback URL.

The container itself is lightweight — it's mostly configuration and the HTTP shim. The intelligence is in the model.

## Agent tooling

Each agent is configured with access to:

- **GitHub tools**: read file contents, list changed files, fetch PR context
- **MCP tools**: query the AnythingLLM knowledge base for relevant security patterns
- **Search**: web search for CVEs and vulnerability references where needed

The agent decides what to look at and in what order. We don't script the exploration path — the model figures that out based on the query and what it finds.

## Model-specific notes

**Claude executor**: uses the Claude API with tool use enabled. Claude is currently our primary executor — it handles code exploration well and produces structured output that's easy to parse.

**Gemini executor**: parallel alternative. Useful for cross-checking findings and for cases where Gemini's training gives it different instincts about a particular pattern.

Running both on the same PR and comparing findings is something we want to do more systematically. Different models catch different things.

## Container lifecycle

Executors are started on demand when a job arrives and run until the agent completes. They're stateless — all persistent state lives in AnythingLLM and the job payload. This makes it straightforward to run many of them in parallel without coordination.

Resource limits (CPU, memory) are set at the container level. For GPU-accelerated models in the future, the same pattern applies — the executor container requests the GPU resource, the rest of the stack doesn't change.

## Adding a new model

To add a new executor:

1. Write a small Go or Python service that implements the HTTP job interface
2. Package it in a Dockerfile
3. Add it to the compose configuration
4. Register the new executor's endpoint in n8n

The pipeline doesn't change. The existing executors keep working.
