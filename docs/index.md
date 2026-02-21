---
layout: default
title: ReviewBot Architecture & Design
---

# ReviewBot: Intelligent Code Review with RAG Context

A GitHub App that performs intelligent security code reviews by maintaining knowledge bases and leveraging Claude/Gemini agents for deep context-aware analysis.

## Our Philosophy: Don't Reinvent, Reuse

Instead of building custom solutions for common problems, we integrate proven, mature technologies:

- **Knowledge Management** → AnythingLLM (RAG + Storage + UI)
- **Pipeline Orchestration** → n8n (Workflow automation)
- **Code Intelligence** → Claude/Gemini (Agent reasoning)
- **Integration Layer** → MCP + Custom executors (Lightweight coordination)

---

## Latest Posts

{% assign sorted_posts = site.posts | sort: 'date' | reverse %}
{% for post in sorted_posts %}
<div class="card">
  <h2><a href="{{ post.url | relative_url }}">{{ post.title }}</a></h2>
  <div class="post-meta">
    <time datetime="{{ post.date | date: '%Y-%m-%dT%H:%M:%SZ' }}">{{ post.date | date: "%B %d, %Y" }}</time>
  </div>
  <p>{{ post.excerpt | strip_html | truncatewords: 30 }}</p>
  <p><a href="{{ post.url | relative_url }}">Read more →</a></p>
</div>
{% endfor %}

---

## Quick Links

- [📋 Full Architecture]({{ '/docs/ARCHITECTURE' | relative_url }})
- [🏗️ KB Maintainer Design]({{ '/docs/kb-maintainer-design' | relative_url }})
- [🔗 GitHub Repository](https://github.com/iv/reviewbot)
- [📚 Intel Database]({{ '/intels' | relative_url }})

---

## Project Structure

```
ReviewBot/
├── main.go                    # GitHub App entry point
├── internal/                  # Core logic
│   ├── config/
│   ├── github/
│   └── handler/
├── services/kb-maintainer/    # Knowledge base sync service
├── intels/                    # Security intelligence markdown files
└── docs/                      # This documentation
```

Each component is designed to be independently deployable and replaceable, following the "don't reinvent" philosophy.
