---
layout: post
title: "Building the Intelligence Database: Offline Processing & Continuous Imports"
date: 2026-02-21 12:00:00 -0000
categories: architecture intelligence
excerpt: "How we gather security intelligence from papers and OWASP, structure it, and keep it fresh."
---

## The Foundation: Human-Curated Intelligence

AnythingLLM's RAG system works best when seeded with high-quality documents. We don't start with empty workspaces.

Instead, we're building **structured security intelligence**:

```
Papers
├── Academic research
├── Security whitepapers
└── Industry analyses

OWASP Reports
├── OWASP Top 10
├── API Security Guidelines
├── Threat Modeling Checklists
└── Secure Coding Practices

Custom Analysis
├── Repo-specific patterns
├── Architecture notes
└── Vulnerability mappings
```

This curated KB gives agents:
- ✅ Authoritative security knowledge
- ✅ Structured taxonomies
- ✅ Actionable guidance
- ✅ Low noise, high signal

## Stage 1: Offline Processing

We process papers and reports **offline** into structured markdown:

```
Input: "OWASP API Security Top 10.pdf"
  ↓
Processing:
├─ Extract key vulnerabilities
├─ Add structured frontmatter (title, severity, tags)
├─ Create clear examples
├─ Link to remediation strategies
  ↓
Output: YAML frontmatter + Markdown content

---
title: Broken Authentication (API1:2019)
severity: critical
tags: [authentication, owasp-api, top10]
references:
  - https://owasp.org/www-project-api-security/
  - https://...
---

# Broken Authentication

Authentication mechanisms are often incorrectly implemented...
```

### Example: AI Red-Teaming Intel

We have curated intelligence packs like `intels/ai-redteaming-intel-pack/`:

```
ai-redteaming-intel-pack/
├── jailbreak-techniques.md      (collected attack patterns)
├── mitigation-strategies.md     (defenses)
├── evaluation-frameworks.md     (how to test)
├── threat-models.md             (taxonomy)
└── case-studies.md              (real incidents)
```

Each document has **structured frontmatter**:

```yaml
---
title: Prompt Injection via User Input
severity: high
category: ai-security
tags:
  - prompt-injection
  - input-validation
  - llm-attacks
  - mitigation
---
```

**Why structured?**
- Agents can filter by severity, tags, category
- Embedding search becomes more precise
- Easy to audit and update
- Future tooling can analyze the corpus

## Stage 2: Continuous Imports (kb-maintainer)

Once documents are created, **kb-maintainer** keeps AnythingLLM in sync:

```
intels/ directory (on host)
    ↓
inotify watches for changes
    ↓
kb-maintainer service
    ├→ Detects new/modified files
    ├→ Computes MD5 (skip unchanged)
    ├→ Uploads to AnythingLLM
    └→ Updates state.json
```

### The Service in Action

```go
// services/kb-maintainer/main.go
func main() {
    // Watch for file changes
    watcher := NewFileWatcher(intelsDir)

    // Sync on startup
    sync.FullSync()

    // Continuous monitoring
    for event := range watcher.Events {
        switch event.Op {
        case CREATE, MODIFY:
            sync.SyncFile(event.Name)
        case DELETE:
            sync.DeleteFile(event.Name)
        }
    }
}
```

### Non-Intrusive Syncing

The kb-maintainer service:
- 🟢 Tracks state in `state.json` (MD5 hashes)
- 🟢 Only uploads changed files
- 🟢 Retries failed uploads
- 🟢 Handles large batches efficiently
- 🟢 Runs periodically (5-minute full-sync)

This means:
- Drop new intel files into `/intels`
- kb-maintainer picks them up automatically
- No manual uploads or admin work
- Knowledge base stays fresh

## How It Flows

```
Researcher creates intel:
    intels/new-vulnerability.md
        ↓
kb-maintainer detects:
    File created event
        ↓
Extracts frontmatter:
    title, severity, tags
        ↓
Uploads to AnythingLLM:
    Workspace: intels
        ↓
Claude agents can now:
    Search & reference the intel
        ↓
Reviews become smarter:
    With new knowledge available
```

## Multi-Workspace Strategy

We use AnythingLLM workspaces strategically:

```
├── workspace: "intels"
│   ├── OWASP guidelines
│   ├── Security papers
│   ├── Best practices
│   └── Vulnerability taxonomies
│   └── Used by: All executors
│
├── workspace: "repo-123"
│   ├── Architecture docs
│   ├── Code patterns
│   ├── Review history
│   └── Used by: Reviewers for repo-123
│
└── workspace: "repo-456"
    └── Separate context for different repo
```

**Benefits:**
- Universal intel (OWASP, papers) shared across repos
- Per-repo context stays isolated
- Agents pick the right workspace for context
- No token waste on irrelevant documents

## Future Evolution

As the intelligence database grows:

1. **Structured querying** - Tag-based filters, severity levels
2. **Embedding optimization** - Code-specific models
3. **Version control** - Track intel changes over time
4. **Feedback loops** - Improve intelligence based on review outcomes
5. **Automated ingestion** - Scan websites, papers for new content
6. **Cross-linking** - Connect related intel documents

## The Workflow

```
Continuous Pipeline (in n8n):

Every 24 hours:
├→ Check for new OWASP releases
├→ Fetch academic papers (RSS, APIs)
├→ Process (extract, structure, format)
├→ Add to intels/
├→ kb-maintainer syncs automatically
└→ Agents have fresh knowledge
```

## Why This Approach?

| Alternative | Problem | Our Approach | Benefit |
|---|---|---|---|
| Unstructured docs | Hard to search, noisy | Frontmatter taxonomy | Precise, filterable |
| Manual uploads | Slow, error-prone | Auto-import service | Fresh, continuous |
| All-in-one workspace | Token waste, noise | Workspace separation | Focused context |
| Third-party KB | Vendor lock-in | Self-hosted AnythingLLM | Full control |

## Philosophy

We're building intelligence, not just storing documents:

1. **Human curation** - Not all content is equal
2. **Structured metadata** - Enables smart search
3. **Continuous updates** - Knowledge base grows
4. **Workspace isolation** - Efficient token use
5. **Open standards** - Easy to export, audit, evolve

The result: agents with access to curated, searchable, continuously-updated security knowledge.

**Philosophy: Build intelligence incrementally, structure it thoughtfully, keep it fresh automatically.**
