---
layout: post
title: "n8n for Pipeline Orchestration: Configurable Over Custom"
date: 2026-02-21 09:00:00 -0000
categories: architecture orchestration
excerpt: "Why we use n8n for workflow automation instead of building a custom orchestrator."
---

## The Challenge: Complex Workflows Without Code

ReviewBot's work involves multiple steps:

1. **Receive GitHub webhook** → Parse PR/comment
2. **Prepare context** → Fetch repo info, files, dependencies
3. **Enrich knowledge** → Query AnythingLLM for relevant patterns
4. **Execute reviewers** → Call Claude/Gemini agents
5. **Post results** → Comment on GitHub, update PR status
6. **Monitor & retry** → Handle failures gracefully

Managing this without an orchestrator means:
- 🔴 Custom retry logic in every service
- 🔴 Complex state management
- 🔴 Hard to modify workflows without code changes
- 🔴 Difficult to debug failures

## Why Not Build Our Own Orchestrator?

We could write a Go/Python scheduler, but:
- Configuring workflows requires code changes + redeployment
- Debugging requires logs scattered across services
- Retries, timeouts, branching logic need custom implementation
- Team members non-fluent in our code can't modify workflows

## Why n8n?

**n8n** is a mature, visual workflow platform designed for exactly this problem:

✅ **Visual Workflow Editor** - Drag & drop pipeline design
✅ **500+ Integrations** - GitHub, HTTP, databases, messaging, etc.
✅ **Conditional Logic** - If/then branches without code
✅ **Error Handling** - Built-in retry, error workflow triggers
✅ **Webhooks** - Receive events from external systems
✅ **Execution History** - See what happened in each run
✅ **Self-Hosted** - Full control, no vendor lock-in

### How We Use It

n8n receives GitHub webhooks and orchestrates the entire review process:

```
GitHub Webhook
    ↓
n8n Workflow
    ├→ Extract PR details
    ├→ Query AnythingLLM
    ├→ Call Claude executor
    ├→ Call Gemini executor
    ├→ Aggregate results
    └→ Post back to GitHub
```

### Example: Error Recovery

With n8n, handling transient failures is built-in:

```
If API call fails:
  ├→ Wait 5 seconds
  ├→ Retry (up to 3 times)
  └→ If still failing:
      └→ Alert maintainers via Slack
```

No custom code needed. Just configuration.

### Real-World Visibility

Here's what a complex ReviewBot pipeline looks like in n8n:

![n8n Pipeline Example](/assets/images/n8n-pipeline-example.png)

Every node is visible, inputs/outputs are logged, and we can modify the workflow in seconds without redeploying anything.

## Key Benefits

| Benefit | Impact |
|---------|--------|
| **Visual debugging** | See exactly where reviews fail |
| **No code redeploy** | Change workflows instantly |
| **Team friendly** | Non-developers can adjust rules |
| **Monitoring** | Built-in execution history & logs |
| **Scaling** | Handle more concurrent reviews |

## What's Next?

As ReviewBot's review strategies evolve:
- A/B test different reviewer configurations
- Add new LLM models without code changes
- Create specialized workflows for different repo types
- Build feedback loops to improve review quality

n8n frees us to focus on the intelligence layer (Claude agents, context retrieval) rather than reinventing workflow orchestration.

**Philosophy: Use n8n's proven orchestration, build ReviewBot's unique review logic.**
