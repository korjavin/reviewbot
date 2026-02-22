# ReviewBot — AI Security Review Pipeline

AI-powered security review bot. When @reviewbot is mentioned on a GitHub PR or issue, it:
1. Clones the repo, runs claude to understand it and find CI gaps
2. Implements 2-3 targeted CI checks with clear "Why:" commit messages
3. Opens a Pull Request with those checks in the target repository

See `ARCHITECTURE.md` for full system design.

## Zone of Responsibility

- **reviewbot** (this service): GitHub operations — webhook ingestion, auth, `/git/checkout`, `/git/create-pr`
- **claude-runner**: AI agent — runs `claude` CLI with a prompt, returns output; supports `workdir` for pre-cloned repos
- **n8n**: Orchestration — pipeline steps and prompts live here; no redeploy to change behaviour
- **anythingllm**: Knowledge Base — intel library + per-repo analysis documents
- **kb-maintainer**: Keeps `./intels/*.md` synced to AnythingLLM

## Architecture (short)

GitHub webhooks → reviewbot → n8n `inbox_handler` → claude-runner (×3) → reviewbot → GitHub PR

## Build & Run

```bash
# Build
go build -o reviewbot main.go

# Run (requires env vars, see .env.example)
./reviewbot

# Docker
docker compose up --build
```

## Project Structure

```
main.go                          — entry point, HTTP server, endpoint registration
internal/config/config.go        — env-var config (incl. SharedReposDir)
internal/github/client.go        — GitHub App client factory (ghinstallation)
internal/github/webhook.go       — webhook validation, parsing, event routing
internal/handler/handler.go      — ClientFactory / TransportFactory interfaces
internal/handler/comment.go      — @reviewbot mention → dispatch job to n8n
internal/handler/pullrequest.go  — PR opened (no-op; pipeline triggered by mention only)
internal/handler/ping.go         — GitHub ping event handler
internal/handler/github_ops.go   — POST /git/checkout and POST /git/create-pr
internal/git/git.go              — low-level git helpers
internal/middleware/logging.go   — request logging middleware
internal/oauth/oauth.go          — OAuth callback for app installation

services/claude-runner/          — AI agent runner (separate Go module)
  main.go                        — POST /review: runs claude in workdir or clones fresh
  Dockerfile

services/kb-maintainer/          — KB Maintainer (separate Go module)
  main.go / sync.go / state.go   — syncs intels/ → AnythingLLM workspace
  Dockerfile

n8n/schemas/
  inbox_handler.json             — MAIN pipeline (import this into n8n)
  claude-runner-test.json        — manual dev/test workflow
  poc-review-pipeline.json       — legacy POC (superseded)

intels/                          — markdown intel files (synced to AnythingLLM)
docs/                            — design docs
```

## Endpoints

- `POST /webhook` — GitHub webhook receiver (validates HMAC, routes events)
- `GET /callback` — OAuth callback (app installation flow)
- `GET /health` — health check
- `POST /git/checkout` — clone repo to shared volume, create review branch
- `POST /git/create-pr` — push review branch, open GitHub PR

## Environment Variables

### reviewbot

| Variable | Required | Description |
|---|---|---|
| `GITHUB_APP_ID` | yes | GitHub App ID |
| `GITHUB_PRIVATE_KEY_PATH` | yes* | Path to PEM file |
| `GITHUB_PRIVATE_KEY` | yes* | Raw PEM contents (alternative) |
| `GITHUB_WEBHOOK_SECRET` | yes | Webhook HMAC secret |
| `GITHUB_CLIENT_ID` | no | OAuth Client ID |
| `GITHUB_CLIENT_SECRET` | no | OAuth Client Secret |
| `PORT` | no | Server port (default: 8080) |
| `BASE_URL` | no | Public URL for OAuth redirects |
| `N8N_WEBHOOK_URL` | yes | n8n inbox_handler webhook URL |
| `SHARED_REPOS_DIR` | no | Root dir for shared volume checkouts (default: `/shared/repos`) |

*One of `GITHUB_PRIVATE_KEY_PATH` or `GITHUB_PRIVATE_KEY` is required.

### claude-runner

| Variable | Required | Description |
|---|---|---|
| `ANTHROPIC_API_KEY` | yes | Anthropic API key for claude CLI |
| `ANYTHINGLLM_URL` | no | AnythingLLM base URL (default: `http://anythingllm:3001`) |
| `ANYTHINGLLM_API_KEY` | yes | AnythingLLM API key (for MCP config seeded into `~/.claude.json`) |

### KB Maintainer (`services/kb-maintainer`)

| Variable | Required | Description |
|---|---|---|
| `ANYTHINGLLM_URL` | no | AnythingLLM base URL (default: `http://anythingllm:3001`) |
| `ANYTHINGLLM_API_KEY` | yes | Bearer token from AnythingLLM Settings → API Keys |
| `ANYTHINGLLM_WORKSPACE` | no | Workspace slug (default: `intels`, created if absent) |
| `INTELS_DIR` | no | Directory to watch (default: `/intels`) |
| `STATE_PATH` | no | Persistent sync state file (default: `/state/kb-maintainer.json`) |
| `SYNC_INTERVAL` | no | Periodic full-resync interval (default: `5m`) |

## Adding Intel Files (Portainer + Podman)

`/intels` inside the container is read-only. Write files on the host:

```bash
# Find the real host path
INTELS=$(podman inspect kb-maintainer \
  --format '{{range .Mounts}}{{if eq .Destination "/intels"}}{{.Source}}{{end}}{{end}}')

# Drop a new file — kb-maintainer picks it up via inotify
cat > $INTELS/my-intel.md << 'EOF'
---
title: ...
severity: high
tags: [example]
---
# ...
EOF

# Watch it sync
podman logs -f kb-maintainer
```

Tip: set `INTELS_DIR=/opt/reviewbot/intels` in Portainer stack env vars to pin
the host path instead of relying on the auto-generated Portainer compose path.

## Local Development

Use [smee.io](https://smee.io) or [ngrok](https://ngrok.com) to proxy webhooks to localhost:

```bash
# smee
npx smee-client --url https://smee.io/YOUR_CHANNEL --target http://localhost:8080/webhook

# ngrok
ngrok http 8080
```

## Testing

```bash
# Health check
curl http://localhost:8080/health

# Verify: create a PR on a repo with the app installed → bot comments
# Verify: comment with @reviewbot → bot replies + adds 👀 reaction
```
