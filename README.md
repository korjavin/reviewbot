# ReviewBot

> **AI-powered security review pipeline for GitHub repositories.**

ReviewBot analyses any GitHub repository for security vulnerabilities, spawns focused AI agents to review the code against a curated knowledge base of security intel, generates lightweight CI checks for confirmed findings, and opens a Pull Request with those checks in the target repo.

Subsequent commits run only the **lightweight generated checks** — the expensive AI pipeline is only invoked again on explicit re-review.

---

## How it works

```
PR opened / manual trigger
        │
        ▼
1.  Checkout target repository
        │
        ▼
2.  Profiler Agent — "what services, frameworks, attack surfaces are here?"
        │
        ▼
3.  Knowledge Base lookup (parallel per topic)
        │
        ▼
4.  Review Agents (parallel, one per intel document)
    "Is this vulnerability present? Evidence?"
        │
        ▼
5.  Check Generator — produces GitHub Actions YAML for confirmed findings
        │
        ▼
6.  PR opened in target repo with generated CI/CD checks
```

---

## Key Components

| Component | Role | Tech |
|-----------|------|------|
| **github-app** | Receives webhooks, authenticates as GitHub App, opens PRs | Go |
| **n8n** | Orchestrates the multi-step review pipeline | n8n (self-hosted) |
| **AnythingLLM** | Knowledge base — document storage, embeddings, hybrid search | AnythingLLM (self-hosted) |
| **kb-maintainer** | Watches `intels/` and keeps AnythingLLM in sync automatically | Go |
| **Agent containers** | Run Gemini / Claude with a job and POST results back asynchronously | Docker |
| **mcp-anythingllm** | MCP server inside agent containers — gives agents native KB access | Go |

→ Full details in [ARCHITECTURE.md](ARCHITECTURE.md)

---

## Repository Layout

```
reviewbot/
├── main.go
├── ARCHITECTURE.md          ← system design & diagrams
├── docker-compose.yml       ← n8n + AnythingLLM services
├── intels/                  ← security intel documents (*.md)
│
├── internal/
│   ├── config/
│   ├── github/              ← webhook handler + API client
│   ├── handler/             ← event handlers
│   ├── git/                 ← git helpers
│   ├── middleware/
│   └── oauth/
│
├── pkg/
│   ├── knowledgebase/       ← KB client library
│   └── pipeline/            ← pipeline types & interfaces
│
├── services/
│   └── kb-maintainer/       ← syncs intels/ → AnythingLLM
│
├── n8n/
│   ├── schemas/             ← n8n workflow definitions
│   └── nodes/               ← custom n8n nodes
│
└── docs/
    └── knowledgebase/
    └── pipeline/
```

---

## Running Locally

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Git

### 1. Start infrastructure services

```bash
docker-compose up -d
```

This starts **n8n** (pipeline orchestrator) and **AnythingLLM** (knowledge base).

### 2. Configure

Copy `.env.example` to `.env` and fill in:

| Variable | Required | Description |
|----------|----------|-------------|
| `GITHUB_APP_ID` | ✓ | GitHub App ID |
| `GITHUB_PRIVATE_KEY_PATH` | ✓* | Path to PEM private key |
| `GITHUB_PRIVATE_KEY` | ✓* | Raw PEM content (alternative) |
| `GITHUB_WEBHOOK_SECRET` | ✓ | Webhook secret |
| `PORT` | — | Server port (default: `8080`) |

*One of `GITHUB_PRIVATE_KEY_PATH` or `GITHUB_PRIVATE_KEY` required.*

### 3. Run the bot

```bash
go run main.go
```

For local webhook testing, proxy with [smee.io](https://smee.io) or ngrok:

```bash
smee -u https://smee.io/your-channel -t http://localhost:8080/webhook
```

---

## GitHub App Permissions

| Permission | Level |
|------------|-------|
| Contents | Read & Write |
| Issues | Read & Write |
| Pull Requests | Read & Write |
| Metadata | Read |

Subscribe to events: **Pull request**, **Issue comment**.

---

## Testing

```bash
go test ./...
```

---

## Intel Documents

Security intelligence lives in `intels/*.md` — tagged, categorized markdown documents covering vulnerability classes. `kb-maintainer` watches this directory and syncs changes to AnythingLLM automatically.

To add new intel: drop a `.md` file into `intels/` and commit. It will be indexed within seconds.
