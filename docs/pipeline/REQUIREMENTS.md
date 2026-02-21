# Pipeline Engine — Requirements

> **Parent**: [ARCHITECTURE.md](../ARCHITECTURE.md)  
> **Status**: Draft · February 2026

---

## 1. Purpose

The Pipeline Engine orchestrates the multi-step, multi-agent workflow that analyses a repository and produces CI/CD security checks. It is built on **[n8n](https://n8n.io)** (self-hosted) extended with custom nodes specific to ReviewBot.

---

## 2. Workflow Overview

```
Trigger
  → Checkout
  → Profiling Agent
  → Knowledge Base Lookup (parallel per topic)
  → Review Agents (parallel per Intel)
  → Check Generator Agent
  → PR Creator
```

Each step is implemented as one or more **n8n nodes**. Custom nodes are compiled as an npm package and mounted into the n8n container.

---

## 3. Custom n8n Nodes

### Node: `Checkout Repository`
- **Inputs**: `repoUrl` (string), `ref` (branch/SHA, optional), `depth` (int, default 1)
- **Outputs**: `workspacePath` (string), `repoMetadata` (object: language, stars, topics)
- **Behaviour**: Clones the repository into a temp directory inside the pipeline container. Authenticates via the GitHub App token injected by the KB service.

### Node: `Search Repository`
- **Inputs**: `workspacePath`, `pattern` (glob or regex), `fileExtensions` (optional)
- **Outputs**: `matches` (array of `{file, lineNumber, content}`)
- **Behaviour**: Runs `ripgrep` or equivalent against the workspace.

### Node: `Knowledge Base Query`
- **Inputs**: `type` (enum: `tags | taxonomy | semantic | hybrid`), `query` (string or tags array), `taxonomyPath` (optional), `topK` (int, default 10)
- **Outputs**: `intels` (array of Intel objects)
- **Behaviour**: Calls the Knowledge Base gRPC/REST API. Configured via environment variable `KB_URL`.

### Node: `Agent Invoke`
- **Inputs**: `model` (enum: `gemini-flash | claude-haiku | ollama/<model>`), `systemPrompt`, `userPrompt`, `temperature`, `maxTokens`
- **Outputs**: `response` (string), `tokenUsage` (object)
- **Behaviour**: Thin wrapper over the chosen LLM API. Uses a unified interface so the model can be swapped via workflow configuration. Supports structured output (JSON schema).

### Node: `Generate CI Check`
- **Inputs**: `findings` (array of confirmed vulnerability findings), `ciSystem` (enum: `github-actions | gitlab-ci | jenkins`), `workspacePath`
- **Outputs**: `files` (array of `{path, content}` — files to add/modify)
- **Behaviour**: Calls the Check Generator Agent (Agent Invoke node internally) and returns generated CI files.

### Node: `Create Pull Request`
- **Inputs**: `workspacePath`, `repoUrl`, `baseBranch`, `files` (array), `prTitle`, `prBody`
- **Outputs**: `prUrl`, `prNumber`
- **Behaviour**: Commits `files` to a new branch (`reviewbot/checks-{timestamp}`) and opens a PR via the GitHub API.

---

## 4. Pipeline Stages (Detailed)

### Stage 1 — Trigger

Accepted trigger sources:
- GitHub webhook on PR opened (via existing `github-app` service)
- Manual HTTP trigger (for testing or re-analysis)
- Scheduled trigger (weekly re-scan of enrolled repositories)

Trigger payload must include: `repoUrl`, `ref`, `installationId`.

### Stage 2 — Profiling Agent

**Goal**: Understand what the repository contains.

System prompt template:
```
You are a security-focused code analyser. Given the file tree and a selection of key files from a repository, identify:
1. Primary programming languages and frameworks.
2. Public-facing services (web, gRPC, TCP, etc.).
3. CI/CD configuration present (GitHub Actions, Dockerfile, etc.).
4. Any dependency manifests (go.mod, package.json, requirements.txt).
5. Return a JSON array of topic slugs relevant to security review, chosen from the provided taxonomy.
```

Output: `string[]` of topic slugs to look up in the KB.

### Stage 3 — Knowledge Base Lookup

For each topic slug returned by the Profiling Agent:
- Call `Knowledge Base Query` node with `type=taxonomy` and `taxonomyPath=<slug>`.
- Merge and de-duplicate results.
- Optionally run a `type=semantic` query with the profiling agent's full output text to catch cross-cutting Intels.

### Stage 4 — Review Agents (Fan-out)

For each Intel returned:
- Spawn a Review Agent in parallel (n8n `SplitInBatches` + async).
- Each agent receives: the Intel document, the relevant code snippets (from `Search Repository`), and a structured output schema.
- Agent returns: `{confirmed: bool, evidence: string, severity: string}`.
- Confirmed findings are collected into a `findings[]` array.

### Stage 5 — Check Generator Agent

Single agent receives all confirmed findings and generates CI files. The prompt instructs it to:
- Write idiomatic GitHub Actions YAML (or other target CI).
- Each generated check must be minimal and fast (no full AI invocation — just static analysis tools, `grep`, linters).
- Include a comment in each check citing the Intel `id` and `title` that motivated it.

### Stage 6 — PR Creation

A PR is opened against the target repository's default branch. The PR description includes:
- A summary of findings (severity breakdown).
- A table listing each generated check and the Intel that triggered it.
- Instructions for the repository owner on how to configure/skip specific checks.

---

## 5. Non-Functional Requirements

| ID | Requirement |
|----|-------------|
| NFR-1 | n8n runs as a standard Docker container; custom nodes are injected via volume mount |
| NFR-2 | Custom nodes are written in TypeScript (n8n standard) |
| NFR-3 | The full pipeline must complete within 10 minutes for repositories ≤ 500 files |
| NFR-4 | Fan-out Review Agents: maximum 20 concurrent LLM calls at once (configurable) |
| NFR-5 | All LLM calls have a configurable timeout (default 60 s) and are retried once on failure |
| NFR-6 | Pipeline execution logs are stored and accessible from the n8n UI for debugging |
| NFR-7 | Sensitive values (API keys, GitHub tokens) are stored in n8n credentials vault, never in workflow JSON |

---

## 6. Environment Variables

| Variable | Description |
|----------|-------------|
| `KB_URL` | Knowledge Base service URL (gRPC or REST) |
| `KB_API_KEY` | API key for KB mutating operations |
| `GITHUB_APP_ID` | GitHub App ID |
| `GITHUB_PRIVATE_KEY` | PEM key for GitHub App auth |
| `LLM_PROVIDER` | Default LLM provider (`gemini`, `claude`, `ollama`) |
| `LLM_MODEL` | Default model name |
| `LLM_API_KEY` | API key for the chosen provider |
| `MAX_CONCURRENT_AGENTS` | Max parallel Review Agent calls (default: 10) |

---

## 7. Out of Scope (v1)

- Real-time progress streaming to the GitHub PR comment
- Cost tracking per pipeline run
- Support for GitLab, Bitbucket (GitHub only for now)
- Custom per-repository taxonomy overrides
