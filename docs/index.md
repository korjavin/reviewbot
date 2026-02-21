---
layout: default
title: ReviewBot
---

# ReviewBot

Security code review is broken. It's expensive, slow, and it doesn't scale — especially for teams without a dedicated security engineer on every project.

We're building an AI-powered reviewer that gets smarter over time. Not just a model that reads your diff, but a system that remembers what it already knows about your codebase, queries a structured knowledge base of security intelligence, and delivers findings that actually reflect context — not just pattern matching on the current PR.

Our guiding principle: **don't build what already exists**. We assemble proven components — AnythingLLM for knowledge management, n8n for orchestration, Claude and Gemini for reasoning — and focus our engineering on what's genuinely new: the intelligence layer, the context accumulation, the review quality.

---

{% assign sorted_posts = site.posts | sort: 'date' | reverse %}
{% for post in sorted_posts %}
<div class="card">
  <h2><a href="{{ post.url | relative_url }}">{{ post.title }}</a></h2>
  <div class="post-meta">
    <time datetime="{{ post.date | date: '%Y-%m-%dT%H:%M:%SZ' }}">{{ post.date | date: "%B %d, %Y" }}</time>
  </div>
  <p>{{ post.excerpt | strip_html | truncatewords: 35 }}</p>
  <a href="{{ post.url | relative_url }}">Read →</a>
</div>
{% endfor %}
