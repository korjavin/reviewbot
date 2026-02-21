---
layout: post
title: "MCP integration in depth: mcp-anythingllm"
permalink: /deep/mcp
---

[← Back to main post]({{ '/blog/2026/02/21/how-we-build-reviewbot/' | relative_url }})

## The problem with per-model API wrappers

Without a standard interface, each executor needs its own code to query AnythingLLM. That means duplicated HTTP client code, duplicated error handling, and per-model maintenance when the AnythingLLM API changes. It's not a lot of code, but it's the kind of scattered duplication that becomes a problem over time.

[Model Context Protocol](https://modelcontextprotocol.io) (MCP) solves this by defining a standard way for models to discover and call external tools. Write one MCP server; every compatible model can use it.

## What mcp-anythingllm does

`mcp-anythingllm` is a small MCP server that wraps the AnythingLLM REST API. It exposes a set of tools that models can call:

- `search_workspace(query, workspace)` — semantic + full-text search
- `get_document(id)` — retrieve a specific document by ID
- `list_workspaces()` — list available workspaces
- `chat_with_context(query, workspace)` — conversational query with RAG context

From the agent's perspective, these are just tools — the same way a web search tool or a code execution tool works. The agent decides when to call them and what to do with the results.

## Deployment

The MCP server runs inside the executor containers. It's not a separate service and it's not network-accessible from outside the containers.

This means:
- Each executor has its own MCP server instance
- Access to the knowledge base is mediated through the executor (no external exposure)
- Configuration (AnythingLLM URL, API key) is injected at container startup

## Why MCP specifically

MCP is the protocol Anthropic ships with Claude and is increasingly supported by other models. Choosing it means:

- Tool discovery is automatic — the agent knows what tools are available without custom prompting
- The same server works for Claude and Gemini (and any other MCP-compatible model)
- As the MCP ecosystem grows, we can add other servers (GitHub, SAST tools, CVE databases) using the same pattern

The alternative was bespoke integration code in each executor. MCP gives us a cleaner boundary.

## Current limitations

The MCP server is currently tightly coupled to AnythingLLM's API. If we switch the backend KB system, we'd need to update `mcp-anythingllm`. The interface the agents see stays the same, but the implementation changes.

This feels like the right tradeoff for now — we don't anticipate switching away from AnythingLLM in the near term, and keeping the server small makes it easy to understand and maintain.
