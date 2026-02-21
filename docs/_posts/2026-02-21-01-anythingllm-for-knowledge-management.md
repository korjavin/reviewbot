---
layout: post
title: "We're not building a knowledge base. We're reusing one."
date: 2026-02-21 08:00:00 -0000
categories: architecture
excerpt: "Every AI security reviewer has the same problem: context is expensive. We solved it without writing a single line of vector database code."
---

Here's the dirty secret of AI code review: it's stupidly expensive to do it right.

To catch a real vulnerability, the AI needs to understand the whole picture — how auth is wired, where secrets flow, what third-party libraries are doing. That means feeding it massive context every single time. For any repo with real codebase size, you're burning 50-100k tokens per review. Run it daily? Weekly? Across 20 repos? The math gets ugly fast.

So we asked the obvious question: **why re-read everything from scratch on every run?**

---

## The insight: treat context like a database

Codebases don't change completely between reviews. What if we stored what the AI already knows — and only retrieved the relevant parts when needed?

That's exactly what RAG (Retrieval Augmented Generation) is for. Index documents once, query them semantically when you need them. Instead of dumping an entire codebase into a prompt, you pull the 2-3 most relevant pieces of context. Same quality, fraction of the tokens.

We could have built that system ourselves. Chunking strategy, embedding pipeline, vector store, retrieval ranking, storage layer, admin UI. Probably 6-8 weeks of solid engineering.

We didn't.

---

## AnythingLLM already exists

[AnythingLLM](https://anythingllm.com) is a mature, self-hosted RAG platform that ships with everything we needed out of the box:

- Document storage and embedding
- Semantic + full-text hybrid search
- Per-workspace separation (so repo A's context doesn't pollute repo B's)
- A clean REST API for programmatic access
- A web UI to browse and chat with indexed documents

We run it ourselves, so nothing leaves our infrastructure. And because it's API-driven, our agents talk to it natively — they search for relevant patterns, retrieve specific documents, and use that context to produce more accurate reviews.

Here's what it looks like when we're exploring an indexed knowledge base:

![AnythingLLM chat interface showing a conversation about security intel]({{ '/assets/images/anythingllm-chat-example.png' | relative_url }})

That's not a prototype — it's the actual system. We can chat with our security knowledge base, ask it questions, verify what's indexed. The same workspace that feeds our automated reviewers.

---

## What we actually get

For any repo we've reviewed before, the second run costs a fraction of the first. Context is already indexed. The agent queries what's relevant, skips what's not.

Over time, each workspace becomes a living record of everything we know about that codebase — past findings, architecture notes, dependency analysis. The more we review, the smarter the next review gets.

That's the compounding effect we're building toward. Not just cheaper reviews — reviews that get better as they accumulate context.

We got all of that by choosing not to reinvent what already exists.
