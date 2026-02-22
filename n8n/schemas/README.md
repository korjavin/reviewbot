# n8n Workflow Schemas

This directory contains versioned n8n workflow definitions that can be imported directly into the n8n UI.

## Workflows

| File | Description |
|------|-------------|
| [`inbox_handler.json`](./inbox_handler.json) | **Main pipeline** — triggered by @reviewbot mentions, runs full security review, opens PR |
| [`claude-runner-test.json`](./claude-runner-test.json) | Manual test — calls claude-runner directly (dev/debug) |
| [`poc-review-pipeline.json`](./poc-review-pipeline.json) | Legacy POC — Gemini-based comment reviewer (superseded) |

---

## `inbox_handler.json` — Main Pipeline

**15-node workflow triggered by @reviewbot mentions:**

```
Webhook
  └── Extract Fields
        └── Checkout Repo      (POST reviewbot /git/checkout)
              └── Sanitize     (no-op pass, future security gate)
                    └── Post Acknowledgement  (GitHub comment: "I'm on it")
                          └── Build Understanding Prompt
                                └── General Understanding   (claude-runner)
                                      └── Build Find Checks Prompt
                                            └── Find CI Checks   (claude-runner)
                                                  └── Parse Checks  (→ 1 item per check)
                                                        └── Loop Over Checks (SplitInBatches)
                                                              ├── [done] → Create PR  (POST reviewbot /git/create-pr)
                                                              │               └── Post PR Link Comment
                                                              └── [batch] → Build Implement Prompt
                                                                              └── Implement Check  (claude-runner)
                                                                                      └── (loop back)
```

### Pipeline steps

| Step | Node | What it does |
|------|------|--------------|
| 1 | Checkout Repo | Clones repo to isolated directory on shared volume, creates `reviewbot-pr{N}-{ts}` branch |
| 2 | Sanitize | No-op for now; future: malicious file scan, rate limiting |
| 3 | Post Acknowledgement | Immediate "I'm working on it" comment so user gets feedback fast |
| 4 | General Understanding | claude-runner analyzes the repo, checks AnythingLLM KB for prior analysis, stores findings |
| 5 | Find CI Checks | claude-runner identifies 2-3 high-value CI checks not already present |
| 6 | Implement Checks (loop) | claude-runner implements each check one at a time, commits with "Why:" message |
| 7 | Create PR | reviewbot pushes the branch and opens a GitHub PR |
| 8 | Post PR Link | Comment on original thread with link to the new PR |

### Trigger payload (sent by reviewbot Go service)

```json
{
  "owner":          "korjavin",
  "repo":           "myproject",
  "clone_url":      "https://github.com/korjavin/myproject.git",
  "default_branch": "main",
  "pr_number":      42,
  "comment_body":   "@reviewbot please review",
  "comment_id":     12345
}
```

Authorization header: `Bearer <github-installation-token>`

### Webhook URL (after import)

```
POST https://n8n.your-domain.com/webhook/reviewbot-inbox
```

Set `N8N_WEBHOOK_URL=https://n8n.your-domain.com/webhook/reviewbot-inbox` in your `.env`.

### Environment variables used

| Variable | Source | Description |
|---|---|---|
| `ANYTHINGLLM_URL` | docker-compose env | `http://anythingllm:3001` |
| `ANYTHINGLLM_API_KEY` | n8n env / `.env` | AnythingLLM API key |

Internal service URLs (Docker network, no env vars needed):
- `http://reviewbot:8080` — for `/git/checkout` and `/git/create-pr`
- `http://claude-runner:8080` — for `/review`

---

## `claude-runner-test.json` — Manual Test

**5-node workflow for developer testing:**

```
Manual Trigger → Test Parameters → Call Claude Runner → Format Result → Review Output
```

Use this to test the claude-runner service directly without triggering a real GitHub webhook.

### How to use

1. Import into n8n
2. Open **Test Parameters** node and set `owner`, `repo`, `pr_number`, `github_token`, `clone_url`, `prompt`
3. Click **Execute Workflow**
4. Check **Format Result** output for the review markdown

See [`docs/testing-claude-runner.md`](../../docs/testing-claude-runner.md) for full instructions.

---

## `poc-review-pipeline.json` — Legacy POC (superseded)

Original 7-node POC pipeline using Gemini REST API. Superseded by `inbox_handler.json`.
Kept for reference.
