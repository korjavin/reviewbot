---
layout: post
title: "Claude in a box. Gemini in a box. Triggered by a webhook."
date: 2026-02-21 10:00:00 -0000
categories: architecture
excerpt: "We wanted the full reasoning power of frontier AI agents without building a custom orchestration layer. So we containerized them and gave them a simple HTTP interface."
---

Claude and Gemini are extraordinary at exploring codebases. Give Claude access to a repo and a question, and it'll find the auth bypass you missed. The challenge isn't the intelligence — it's the orchestration.

These models aren't instant. A proper code exploration run takes minutes. You can't block a workflow engine waiting for that. And you need to be able to run multiple agents in parallel, spin them up on demand, and not pay for idle compute.

---

## The pattern: containers with webhook callbacks

Each agent — Claude, Gemini, or any future model — runs as an independent container with a simple HTTP interface:

1. **n8n POSTs a job**: repo URL, query, knowledge base workspace, callback address
2. **The container starts the agent**: it has everything it needs — MCP access to the KB, GitHub API credentials, the task
3. **The agent does its work**: explores the code, queries context, reasons about findings
4. **Results POST back**: to a webhook endpoint, picked up by n8n, which continues the pipeline

That's it. No custom scheduler. No distributed job queue. No complex state machine.

---

## Why containers

Each run gets a clean environment. No state leaking between jobs. The agent can be given read access to only what it needs for that review. If it crashes, the container dies cleanly — no impact on anything else.

Scaling is straightforward: more concurrent reviews means more containers. Container runtime (Docker, Podman) handles the resource limits.

And crucially — it's testable. We run the exact same container locally that we run in production. The agent behaves the same in both environments.

---

## The lightweight shim

The container itself doesn't do much beyond starting the agent and handling the HTTP contract. The intelligence is entirely in Claude or Gemini. The shim is ~100 lines — receive job, configure agent, fire callback when done.

This means we can swap models, upgrade agent versions, or add entirely new reviewers (a security-specialized agent, a performance-focused one) just by building a new container. The pipeline doesn't change. n8n doesn't change.

---

## What this unlocks

We can run Claude and Gemini on the same codebase simultaneously and compare their findings. We can A/B test different agent prompting strategies without changing infrastructure. We can add a model the day it's released, just by wrapping it the same way.

The power is in keeping the interface simple. Webhook in, webhook out. Everything else is the agent's problem.
