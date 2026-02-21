---
layout: post
title: "Security intelligence in depth: structured intels and kb-maintainer"
permalink: /deep/intelligence
---

[← Back to main post]({{ '/blog/2026/02/21/how-we-build-reviewbot/' | relative_url }})

## Document structure

Each intel document is a markdown file with YAML frontmatter. The frontmatter captures enough metadata to make retrieval more precise:

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
references:
  - https://owasp.org/www-project-top-10-for-large-language-model-applications/
---
```

The document body covers the technique: what it is, how it typically manifests in code, what to look for during review, and how to remediate it.

We lean on the embedding and full-text search for retrieval rather than building complex taxonomy-based filtering. The tags and categories exist to improve retrieval quality and help with browsing, not to gate access to content.

## Intel packs

Documents are organized into thematic packs. The current ones:

**ai-redteaming-intel-pack** — covers AI-specific attack vectors: prompt injection patterns, jailbreak techniques, evaluation frameworks for testing model robustness, and real incident case studies.

**owasp-api-security** — structured versions of the OWASP API Security Top 10, formatted for retrieval rather than reading.

**general-web-security** — common web vulnerabilities (XSS, CSRF, SQLi, etc.) in a form useful for code review context.

More packs will be added as we review more types of codebases.

## kb-maintainer

The importer is a small Go service that watches the `intels/` directory and syncs its contents to AnythingLLM.

It uses filesystem events (inotify on Linux) to detect changes:
- **File created**: upload to AnythingLLM, store MD5 in state
- **File modified**: compare MD5 against stored state; upload if changed
- **File deleted**: remove from AnythingLLM, clean up state

State is persisted in a JSON file (`kb-maintainer.json`) so it survives restarts without re-uploading everything.

A periodic full sync (every 5 minutes by default) catches anything that might have been missed due to event delivery issues.

## Adding new intels

Drop a markdown file into the `intels/` directory (or the host path mapped to it). The importer picks it up within a few seconds and uploads it to the configured AnythingLLM workspace. No other steps needed.

For adding intel via SSH when using Portainer or Podman, the process is:

```bash
INTELS=$(podman inspect kb-maintainer \
  --format '{{range .Mounts}}{{if eq .Destination "/intels"}}{{.Source}}{{end}}{{end}}')

cat > $INTELS/my-intel.md << 'EOF'
---
title: ...
severity: high
tags: [example]
---
# Content here
EOF
```

Setting `INTELS_DIR=/opt/reviewbot/intels` in the stack environment variables avoids needing to discover the path dynamically.

## Per-repo workspaces

Beyond the universal intel workspace, ReviewBot maintains per-repo workspaces in AnythingLLM. These accumulate context from each review run:

- Findings from previous reviews
- Architectural notes extracted by agents
- Dependency vulnerability summaries
- Custom patterns noted by reviewers

This means the second review of a codebase starts with more context than the first. The gap narrows over time between what ReviewBot knows about a repo and what a human reviewer who's spent time with it would know.

The per-repo workspaces are populated by the agent executors as part of the review process, not by the kb-maintainer service (which only handles the universal intel).
