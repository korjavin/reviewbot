---
layout: post
title: "Pipeline orchestration in depth: n8n"
permalink: /deep/n8n
---

[← Back to main post]({{ '/blog/2026/02/21/how-we-build-reviewbot/' | relative_url }})

## What we're orchestrating

A single review run involves:

1. Receive GitHub webhook (PR opened, comment posted, etc.)
2. Parse event, extract relevant metadata
3. Fetch additional context from GitHub API (files changed, repo info)
4. Query AnythingLLM for relevant security patterns
5. Trigger one or more agent executors with the assembled context
6. Wait for asynchronous callbacks from agents
7. Aggregate findings
8. Post results back to GitHub (comment, review, label)

Each step can fail for different reasons — API rate limits, transient network issues, agent timeouts. Some steps can run in parallel. The shape of the pipeline may vary based on the event type or the repository configuration.

Writing this as application code works, but it becomes rigid. Changing step order, adding a notification, adjusting retry behavior — each change requires a code edit and redeploy.

## Why n8n

[n8n](https://n8n.io) is an open-source workflow automation platform. We self-host it. The pipeline is defined in its editor rather than in code, which means changes to orchestration logic don't require touching the application.

The properties we care about:

**Webhook handling.** n8n receives GitHub webhooks directly. It also provides webhook URLs for agents to POST their results back to when they're done — which is how we handle the async callback pattern.

**Execution history.** Every workflow run is logged with inputs and outputs at each node. When something goes wrong, we can look at the specific execution and see exactly where it failed and what the data looked like at that point. This is significantly easier than reconstructing failures from logs.

**Built-in retry.** Network errors and transient API failures are handled with configurable retry logic at the node level. We don't write retry code.

**Branching.** Different event types (PR opened vs. comment vs. review requested) can route to different sub-workflows. This is straightforward to configure in the editor.

**Self-hosted.** GitHub webhook payloads, API tokens, and intermediate data stay in our infrastructure.

## The async callback pattern

Agents take a while to run — often several minutes for a thorough code review. n8n workflows can't just block and wait for that without holding up the workflow engine.

The pattern we use: n8n sends the job to an executor container and moves on. The executor runs the agent, and when it's done, it POSTs results to a separate n8n webhook endpoint. That triggers the continuation of the workflow — aggregation, posting results to GitHub, etc.

This keeps the workflow engine free and makes it easy to run multiple agents in parallel for the same review.

## Operational notes

n8n is one more service to run. We operate it alongside AnythingLLM and the executor containers. The main maintenance consideration is persisting the workflow definitions and execution history — n8n uses a database for this, which needs a volume mount.

We've found it reasonably low-maintenance once set up. The web UI is where most of the active work happens when adjusting pipeline behavior.
