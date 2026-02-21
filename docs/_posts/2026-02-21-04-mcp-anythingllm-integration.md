---
layout: post
title: "Giving our agents native access to the knowledge base"
date: 2026-02-21 11:00:00 -0000
categories: architecture
excerpt: "Model Context Protocol lets Claude and Gemini treat our knowledge base as a tool they just... use. No API wrappers, no custom glue code."
---

The AI agent needs to search the knowledge base. Simple enough — hit the AnythingLLM API, parse the response, format it into context.

Except you'd need to write that wrapper for Claude. Then again for Gemini. Then again for any future model. Each one needs to know the API shape, handle errors, manage authentication. It's not hard work, but it's repetitive, fragile, and it lives in every executor you ship.

There's a better way.

---

## MCP: a universal tool interface for AI models

[Model Context Protocol](https://modelcontextprotocol.io) is an open standard from Anthropic that defines how AI models discover and call external tools. Instead of writing per-model API wrappers, you write one MCP server — and every MCP-compatible model can use it.

We built `mcp-anythingllm`: a small MCP server that exposes our knowledge base as a set of native tools. The agent sees tools like `search_workspace`, `get_document`, `list_workspaces` — and uses them the same way it would use any other built-in capability.

---

## What it looks like in practice

When Claude is reviewing code and needs security context, it doesn't make an explicit API call. It reasons:

> "I need to understand common JWT implementation vulnerabilities for this codebase."

And then it just calls the search tool. The MCP layer handles the rest — routing the request to AnythingLLM, returning relevant documents, formatting them into context Claude can use.

From the agent's perspective, the knowledge base is just another capability it has. From our infrastructure perspective, it's a small server that translates MCP calls into AnythingLLM API calls.

---

## Internal only, intentionally

`mcp-anythingllm` isn't exposed externally. It runs inside the executor containers — available to the agents, invisible to everything else. The knowledge base is only accessible through the reviewers that need it.

This keeps the security model simple: agents get access to KB through a controlled interface, nothing else in the stack touches it directly.

---

## Adding new tools the same way

The same pattern works for anything else we want to give agents access to — GitHub repo browsing, SAST tool output, dependency vulnerability databases. Write an MCP server, add it to the executor, and every model gets the tool immediately.

We're not building a custom tool integration layer. We're using a standard that the model providers are converging on, which means the ecosystem of available tools will only grow.

One protocol to connect the agents to everything they need.
