---
layout: post
title: "Structured security intelligence, continuously imported"
date: 2026-02-21 12:00:00 -0000
categories: architecture
excerpt: "An AI reviewer is only as good as what it knows. We're building a curated, structured database of security intelligence — and a pipeline to keep it fresh automatically."
---

An AI with great reasoning but no domain knowledge will miss things. It needs to know what a JWT timing attack looks like, what the OWASP API Top 10 actually means in code, how prompt injection works in LLM-backed applications.

You can bake some of this into prompts, but prompts have limits. What you really want is a knowledge base the agent can query — rich, structured, authoritative, and always up to date.

That's what we're building.

---

## Starting with the right raw material

We process papers, OWASP reports, and security research offline into structured markdown documents. Not just extracted text — genuinely useful intel with metadata that makes it searchable and filterable.

Each document has frontmatter that captures what it's about:

```yaml
---
title: Prompt Injection via Untrusted User Input
severity: high
category: ai-security
tags:
  - prompt-injection
  - llm-attacks
  - input-validation
  - mitigation
---
```

Then the document itself covers the technique in depth — what it is, how it manifests, how to catch it in code review, how to remediate it.

The result is a collection of intel packs organized by domain. The `ai-redteaming-intel-pack`, for example, covers jailbreak techniques, prompt injection patterns, evaluation frameworks, and real incident case studies. Not raw documents — curated, structured, actionable knowledge.

We lean on embedding and full-text search rather than rigid taxonomies, so the tags and categories are there to improve retrieval quality, not to gate access.

---

## Getting it into the system automatically

Once documents exist, our `kb-maintainer` service keeps AnythingLLM in sync without any manual intervention.

It watches the `intels/` directory using filesystem events. The moment a file lands there — new intel, updated document, anything — it picks it up, checks if it's actually changed (MD5 comparison against cached state), and uploads it to the appropriate AnythingLLM workspace.

Drop a file in. It's indexed within seconds. No admin panel, no manual upload step, no deployment needed.

The same service runs a full sync periodically to catch anything that might have slipped through, and handles deletions — when a document is removed from the directory, it's removed from the knowledge base too.

---

## Workspace strategy

We separate our intel into two categories of workspace:

**Universal intel** — OWASP, academic papers, general security patterns. Loaded into a shared workspace that every agent can query regardless of what repo they're reviewing.

**Per-repo context** — findings from previous reviews, architecture notes, dependency analysis specific to one codebase. Isolated so there's no cross-contamination between projects.

This means when an agent reviews a new PR on a repo it's seen before, it has two sources of context: the universal security knowledge, and everything it's already learned about that specific codebase.

---

## The flywheel

Here's what this enables over time: every review adds to the per-repo knowledge base. The next review starts with more context than the last. Findings from six months ago inform analysis today. Patterns that kept appearing in one codebase get noted and resurface when the same code shows up in a different form.

We're not just reviewing code. We're building institutional memory for security analysis — and doing it in a way that compounds automatically.

The intelligence database is how ReviewBot gets smarter without anyone having to make it smarter.
