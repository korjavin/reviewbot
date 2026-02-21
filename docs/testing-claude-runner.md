# Testing Claude Runner

This document covers two testing paths:

1. **n8n workflow** — trigger a review via the n8n UI (end-to-end test)
2. **Docker exec** — log in, explore, and test Claude + MCP manually inside the container

---

## Prerequisites

Make sure the stack is running:

```bash
docker compose up -d claude-runner anythingllm n8n
```

Verify claude-runner is healthy:

```bash
curl http://localhost:8080/health   # if port-forwarded locally
# OR from inside n8n container:
docker exec -it n8n wget -qO- http://claude-runner:8080/health
```

Expected response: `{"status":"ok"}`

---

## 1. n8n Workflow Test

### Import the workflow

1. Open n8n UI → **+** → **Import from file**
2. Select `n8n/schemas/claude-runner-test.json`
3. The workflow has these nodes:

```
Manual Trigger → Test Parameters → Call Claude Runner → Format Result → Review Output
```

### Configure and run

1. Open the **Test Parameters** node and edit:
   - `owner` / `repo` — GitHub repo to review
   - `pr_number` — PR number for the prompt context
   - `github_token` — personal access token (needs `repo` scope to clone private repos; public repos work without auth but still need a non-empty token value)
   - `clone_url` — full HTTPS URL, e.g. `https://github.com/korjavin/reviewbot.git`

2. (Optional) Set `GITHUB_TEST_TOKEN` as an n8n environment variable to avoid hardcoding in the node.

3. Click **Execute Workflow** (▶ button).

4. Once done, open **Format Result** → **Output** to read the review. The `review` field contains the full Claude-generated markdown.

> **Timeout note**: claude-runner may take up to 10 minutes on large repos. The workflow node already has an 11-minute timeout configured.

### Health-check node

The **Health Check (optional)** node is disconnected. Right-click it → **Execute node** to ping claude-runner independently.

---

## 2. Docker Exec — Manual Claude + MCP Testing

### Enter the container

```bash
docker exec -it -u claude claude-runner bash
```

You land in a Debian environment as the `claude` user (`/home/claude`) with:
- `claude` CLI (from `@anthropic-ai/claude-code`)
- `anythingllm-mcp-server` (global npm package)
- `git`

### First-time: authenticate Claude

Claude Code needs your Anthropic account or API key. Inside the container:

```bash
# Option A — API key (recommended for headless/server use)
export ANTHROPIC_API_KEY=sk-ant-...
# This is already injected via docker-compose env, so it should work out of the box.

# Option B — interactive browser login (if you need OAuth)
claude auth
# Follow the URL printed in the terminal on your laptop.
```

Auth is persisted in the `claude-auth` Docker volume mounted at `/home/claude`.
You only need to do this **once**.

### Verify Claude works

```bash
echo "Say hello in one sentence." | claude -p "Hello test"
# or
claude -p "What is 2+2?" --dangerously-skip-permissions
```

### Check the MCP config

On startup `claude-runner` seeds `~/.claude.json` (i.e. `/home/claude/.claude.json`) with the AnythingLLM MCP config:

```bash
cat ~/.claude/settings.json
```

Expected output:
```json
{
  "mcpServers": {
    "anythingllm": {
      "command": "anythingllm-mcp-server",
      "env": {
        "ANYTHINGLLM_API_KEY": "<your-key>",
        "ANYTHINGLLM_BASE_URL": "http://anythingllm:3001"
      }
    }
  }
}
```

### Test MCP connectivity manually

Start the MCP server directly to confirm it can reach AnythingLLM:

```bash
ANYTHINGLLM_API_KEY=$ANYTHINGLLM_API_KEY \
ANYTHINGLLM_BASE_URL=$ANYTHINGLLM_URL \
  anythingllm-mcp-server
# Press Ctrl+C after a few seconds — if it starts without errors, the connection works.
```

### Run a full review inside the container

Clone a repo and invoke claude with the MCP server active:

```bash
# Clone a public repo for testing
git clone --depth=1 https://github.com/korjavin/reviewbot.git /tmp/test-repo
cd /tmp/test-repo

# Run claude with MCP (AnythingLLM KB available as tools)
claude -p "Review this repo for security issues. Use the AnythingLLM MCP tools to query the knowledge base first." \
  --dangerously-skip-permissions
```

Claude will:
1. List available MCP tools (you'll see `anythingllm__*` tools)
2. Query the AnythingLLM workspace
3. Read files in the repo
4. Output a markdown review

### Test via curl (without n8n)

From **your laptop** (if you port-forward or expose claude-runner):

```bash
# Port-forward temporarily
docker port claude-runner 8080   # check if exposed

# OR exec into another container on the same network, e.g. n8n:
docker exec -it n8n sh -c '
  wget -qO- --post-data='"'"'{
    "clone_url": "https://github.com/korjavin/reviewbot.git",
    "github_token": "dummy-for-public-repo",
    "owner": "korjavin",
    "repo": "reviewbot",
    "pr_number": 1
  }'"'"' \
  --header="Content-Type: application/json" \
  http://claude-runner:8080/review
'
```

Or add a temporary port mapping in `docker-compose.yml` for local testing:
```yaml
claude-runner:
  ports:
    - "18080:8080"   # ← add this, then `docker compose up -d claude-runner`
```

Then from your laptop:
```bash
curl -s -X POST http://localhost:18080/review \
  -H "Content-Type: application/json" \
  -d '{
    "clone_url": "https://github.com/korjavin/reviewbot.git",
    "github_token": "YOUR_PAT_HERE",
    "owner": "korjavin",
    "repo": "reviewbot",
    "pr_number": 1
  }' | jq .review -r
```

---

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---|---|---|
| `{"status":"ok"}` not returned | container not running | `docker compose up -d claude-runner` |
| `clone failed` | bad token or repo URL | check `github_token` and `clone_url` |
| `review failed` | Claude auth issue | `docker exec -it claude-runner claude auth` |
| MCP tools not showing | settings.json missing | delete volume and restart: `docker volume rm reviewbot_claude-auth && docker compose up -d claude-runner` |
| Timeout in n8n | large repo / slow Claude | increase node timeout or use a smaller repo |

---

## What to look for in a successful MCP test

In the Claude output you should see:
```
Using tool: anythingllm__query_workspace
Tool result: { "textResponse": "...", "sources": [...] }
```

This confirms:
- Claude Code is running ✅
- The AnythingLLM MCP server is connected ✅
- The knowledge base is being queried ✅
