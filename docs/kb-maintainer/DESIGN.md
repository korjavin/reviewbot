# KB Maintainer — Design

> **Parent**: [ARCHITECTURE.md](../ARCHITECTURE.md)
> **Status**: Implemented · February 2026

---

## 1. Purpose

KB Maintainer is a small Go service that watches the `./intels` directory and
keeps it in sync with an **AnythingLLM** workspace via the AnythingLLM REST API.

It replaces the earlier plan of building a custom Knowledge Base service (see
[knowledgebase/DESIGN.md](../knowledgebase/DESIGN.md)). AnythingLLM provides
document storage, vector embedding, and semantic search out of the box; KB
Maintainer's only job is to keep the workspace up to date as intel files change.

---

## 2. How It Works

```
./intels/*.md
      │
      │  inotify (fsnotify) + periodic full resync
      ▼
kb-maintainer (Go service)
      │
      │  POST /api/v1/document/upload          ← upload raw markdown
      │  POST /api/v1/workspace/{slug}/update-embeddings  ← embed in workspace
      ▼
AnythingLLM (http://anythingllm:3001)
      │
      ▼
Workspace "intels"  (queryable by n8n / pipeline agents)
```

### Sync logic

1. **Startup**: load persisted state (`/state/kb-maintainer.json`), then run a
   full sync of all `.md` files in `/intels`.
2. **File watcher** (`fsnotify`): on `Create` or `Write` events, upload the
   changed file and embed it in the workspace.  On `Remove` / `Rename`, remove
   it from the workspace.
3. **Periodic resync** (default every 5 min): full directory scan catches any
   changes that inotify may have missed (e.g. volume mounts in Docker that don't
   propagate kernel events).
4. **State file**: maps each filename to `{hash, doc_location, uploaded_at}`.
   A file is skipped on resync if its MD5 hash has not changed.

### Startup retry

KB Maintainer calls `EnsureWorkspace` in a retry loop with exponential backoff
(5 s → 10 s → … → 60 s cap, deadline 5 min) so it gracefully waits for
AnythingLLM to become healthy before the first sync.

---

## 3. Configuration

All configuration is via environment variables.

| Variable | Default | Description |
|---|---|---|
| `ANYTHINGLLM_URL` | `http://anythingllm:3001` | AnythingLLM base URL |
| `ANYTHINGLLM_API_KEY` | — **required** | Bearer token from AnythingLLM Settings → API Keys |
| `ANYTHINGLLM_WORKSPACE` | `intels` | Workspace slug (created if absent) |
| `INTELS_DIR` | `/intels` | Directory to watch (bind-mounted from host `./intels`) |
| `STATE_PATH` | `/state/kb-maintainer.json` | Persistent sync state file |
| `SYNC_INTERVAL` | `5m` | Periodic full-resync interval (Go duration string) |

---

## 4. AnythingLLM API calls used

| Method | Endpoint | Purpose |
|--------|----------|---------|
| `GET` | `/api/v1/workspace/{slug}` | Check workspace exists |
| `POST` | `/api/v1/workspace/new` | Create workspace if missing |
| `POST` | `/api/v1/document/upload` | Upload markdown file |
| `POST` | `/api/v1/workspace/{slug}/update-embeddings` | Add / remove doc from workspace |

All requests carry `Authorization: Bearer <ANYTHINGLLM_API_KEY>`.

---

## 5. File Layout

```
services/kb-maintainer/
├── main.go       — entry point, wires watcher + ticker + signal handler
├── config.go     — env-var config loading
├── client.go     — AnythingLLM HTTP client
├── sync.go       — SyncFile / DeleteFile / FullSync logic
├── state.go      — JSON state persistence + MD5 hashing
├── go.mod
├── go.sum
└── Dockerfile    — multi-stage build, non-root user
```

---

## 6. Docker / Compose

The service is declared in `docker-compose.yml` under the name `kb-maintainer`.
Important mounts:

| Container path | Source | Mode |
|---|---|---|
| `/intels` | `${INTELS_DIR:-./intels}` on host | `ro` (read-only) |
| `/state` | named volume `kb-maintainer-state` | read-write |

`depends_on: anythingllm` ensures Docker starts AnythingLLM first (the startup
retry loop handles the case where AnythingLLM is still initialising).

---

## 7. CI/CD

The image is built by `.github/workflows/deploy.yml` as a separate
`docker/build-push-action` step using `context: ./services/kb-maintainer`.

Image name: `ghcr.io/{owner}/reviewbot-kb-maintainer:{sha}`

The deploy workflow patches the image tag in `docker-compose.yml` on the
`deploy` branch and triggers a Portainer redeploy, same as the main reviewbot
service.

---

## 8. Intel file format

Intel files are plain Markdown with optional YAML front-matter. The front-matter
is preserved and uploaded as-is; AnythingLLM will include it in the embedded
text. Example:

```markdown
---
id: github-actions-secret-leakage
title: GitHub Actions — Secret Leakage via Environment Variables
severity: high
tags: [github-actions, secrets, ci-cd]
taxonomy: security/ci-cd/github-actions
---

# GitHub Actions — Secret Leakage …
```

---

## 9. Adding Intel Files in Production (Portainer + Podman)

The `/intels` mount inside the container is **read-only**. New files must be
written on the **host** side of the bind mount.

### Find the host path

```bash
podman inspect kb-maintainer \
  --format '{{range .Mounts}}{{if eq .Destination "/intels"}}{{.Source}}{{end}}{{end}}'
```

This prints the real host path (e.g. `/data/compose/7/intels`).

### Write a new intel file

```bash
INTELS=$(podman inspect kb-maintainer \
  --format '{{range .Mounts}}{{if eq .Destination "/intels"}}{{.Source}}{{end}}{{end}}')

cat > $INTELS/my-new-intel.md << 'EOF'
---
title: My New Intel
severity: high
tags: [example]
---
# My New Intel
...
EOF
```

kb-maintainer picks up the new file via inotify within milliseconds.  Check:

```bash
podman logs -f kb-maintainer
```

### Recommended: pin the host path

To avoid hunting for the Portainer-generated path, set `INTELS_DIR` to a fixed
absolute path in your stack env vars (e.g. `INTELS_DIR=/opt/reviewbot/intels`),
create the directory on the host, and redeploy. The path is then predictable
across redeployments.

---

## 10. Limitations & Future Work

- Only top-level `.md` files in `INTELS_DIR` are watched (no subdirectory recursion).
- Documents are never deleted from the AnythingLLM document store; they are only
  removed from the workspace embedding index.  Use the AnythingLLM UI to purge
  old documents if needed.
- No retry on individual file upload failures within a sync run — the next
  periodic resync will retry them.
