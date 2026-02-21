# n8n Workflow Schemas

This directory contains versioned n8n workflow definitions that can be imported directly into the n8n UI.

## Workflows

| File | Description |
|------|-------------|
| [`poc-review-pipeline.json`](./poc-review-pipeline.json) | POC end-to-end review pipeline (Gemini via HTTP) |
| [`claude-runner-test.json`](./claude-runner-test.json) | Manual test workflow — calls claude-runner service |

---

## `claude-runner-test.json`

**Manual test workflow — 5 nodes:**

```
Manual Trigger → Test Parameters → Call Claude Runner → Format Result → Review Output
```

### Purpose

Test the `claude-runner` container (the Go HTTP service that runs `claude` + AnythingLLM MCP).
No GitHub webhook needed — trigger from the n8n UI directly.

### How to use

1. Import the file into n8n
2. Open **Test Parameters** node and set `owner`, `repo`, `pr_number`, `github_token`, `clone_url`
3. Click **Execute Workflow**
4. Check **Format Result** output for the review markdown

See [`docs/testing-claude-runner.md`](../../docs/testing-claude-runner.md) for full instructions including `docker exec` and MCP validation.

---

## `poc-review-pipeline.json`

**7-node pipeline:**

```
Webhook → Extract PR Data → Query AnythingLLM KB
       → Prepare LLM Prompt → LLM Security Analysis (Gemini)
       → Parse & Format Comment → Post GitHub PR Comment → Respond OK
```

### How to import

1. Open n8n UI (`https://n8n.your-domain.com`)
2. Click **+** → **Import from file**
3. Select `poc-review-pipeline.json`
4. Activate the workflow

### Webhook URL

After import the webhook is available at:
```
POST https://n8n.your-domain.com/webhook/reviewbot-pipeline
```

### Trigger payload (from reviewbot Go service)

```json
{
  "repoFullName":      "owner/repo",
  "prNumber":          42,
  "ref":               "abc123def456",
  "repoUrl":           "https://github.com/owner/repo",
  "installationToken": "<short-lived GitHub App installation token>"
}
```

### Environment variables used

| Variable | Source | Description |
|---|---|---|
| `ANYTHINGLLM_URL` | docker-compose env | e.g. `http://anythingllm:3001` |
| `ANYTHINGLLM_API_KEY` | n8n env / `.env` | AnythingLLM API key |
| `ANYTHINGLLM_WORKSPACE` | n8n env / `.env` | workspace slug (default: `intels`) |
| `LLM_API_KEY` | n8n env / `.env` | Gemini API key |

### Switching LLM provider

The **"LLM Security Analysis"** node calls Gemini REST API directly.
To use OpenAI-compatible endpoints, update the node's URL and body format:

```
URL: https://api.openai.com/v1/chat/completions
Auth header: Authorization: Bearer <LLM_API_KEY>
Body: { model, messages: [{role, content}], response_format: {type: "json_object"} }
```

### Next steps (production hardening)

- [ ] Add retry logic on LLM timeout (n8n Error Trigger + loop)
- [ ] Fan-out Review Agents per Intel (SplitInBatches node)
- [ ] Generate CI check files (Generate CI Check node)
- [ ] Open PR with generated checks instead of posting comment
- [ ] Replace HTTP Request nodes with custom TypeScript nodes
