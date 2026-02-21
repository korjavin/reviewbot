---
layout: default
title: ReviewBot — Building a Smarter Code Reviewer
---

# ReviewBot

We're building an automated security code reviewer powered by AI agents — designed to work across many repositories while getting smarter over time, not just blindly re-reading everything on each run.

Our core philosophy: **don't reinvent, reuse**. There are mature, battle-tested systems for knowledge storage, workflow orchestration, and AI reasoning. We integrate them rather than rebuilding from scratch.

---

## Design Decisions

{% assign sorted_posts = site.posts | sort: 'date' | reverse %}
{% for post in sorted_posts %}
<div class="card">
  <h2><a href="{{ post.url | relative_url }}">{{ post.title }}</a></h2>
  <div class="post-meta">
    <time datetime="{{ post.date | date: '%Y-%m-%dT%H:%M:%SZ' }}">{{ post.date | date: "%B %d, %Y" }}</time>
  </div>
  <p>{{ post.excerpt | strip_html | truncatewords: 35 }}</p>
  <a href="{{ post.url | relative_url }}">Read more →</a>
</div>
{% endfor %}
