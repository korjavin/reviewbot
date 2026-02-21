---
layout: post
title: "Pipelines without a pipeline team"
date: 2026-02-21 09:00:00 -0000
categories: architecture
excerpt: "Running AI agents across dozens of repos means orchestrating a lot of moving parts. We didn't write a custom orchestrator. We used n8n."
---

Running a security review isn't one API call. It's a sequence: receive a webhook, pull repo context, query the knowledge base, trigger one or more AI agents, aggregate their output, post results back to GitHub. If any step fails, retry it. If an agent times out, alert someone.

Writing that orchestration from scratch is a significant chunk of work. And the moment requirements change — new agent, different retry logic, new notification channel — you're back in the code.

We've seen teams burn months on this kind of glue. We didn't want to.

---

## n8n as the backbone

[n8n](https://n8n.io) is an open-source workflow automation platform. Think Zapier, but self-hosted, with full programmatic power and a visual editor. It's what we use to wire every step of the review pipeline together.

Here's what a real pipeline looks like:

![n8n pipeline showing the review workflow with connected nodes]({{ '/assets/images/n8n-pipeline-example.png' | relative_url }})

Each node is a step. Connections show data flow. You can see exactly where a run succeeded or failed, inspect the input and output of any step, and change behavior without touching code.

---

## Why it works for us

**Webhooks in, webhooks out.** GitHub sends us an event, n8n receives it. Our agents POST their results back when they're done. n8n handles the rest.

**Visual debugging.** When something breaks mid-pipeline, we don't grep through logs. We open the execution history, click the failed node, and see exactly what went in and what came out.

**No redeploy to change behavior.** Want to add a Slack notification when a critical vulnerability is found? Two minutes in the editor, no code push. This matters a lot in early stages when we're iterating fast on review strategies.

**Retry and error handling built in.** Transient failures — API timeouts, rate limits — are handled with configurable retry logic. No custom code needed.

---

## The tradeoff we accepted

n8n adds a service to our stack. We run it ourselves, which means one more thing to operate. For a team of our size, that's a reasonable cost for what we get back: the ability to change pipeline logic without engineering time.

As the review system matures — more agent types, more repo-specific strategies, feedback loops — having a visual, configurable orchestration layer means those changes stay fast and cheap.

We build the intelligence. n8n handles the plumbing.
