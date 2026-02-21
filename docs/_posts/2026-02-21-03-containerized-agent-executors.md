---
layout: post
title: "Containerized Claude & Gemini Executors: Lightweight Agent Orchestration"
date: 2026-02-21 10:00:00 -0000
categories: architecture agents
excerpt: "How we package Claude and Gemini as independent, containerized services triggered by n8n."
---

## The Problem: Agents Need Lightweight Execution

Claude and Gemini agents excel at code exploration and reasoning, but orchestrating them from n8n creates a dilemma:

🔴 **Heavy orchestration**: Build a complex n8n integration with state management
🔴 **Embedded agents**: Spin up agents inside n8n, blocking the workflow engine
🔴 **Long polling**: Agents take minutes to explore code; n8n shouldn't wait synchronously

We needed a way to:
- ✅ Trigger agents asynchronously
- ✅ Let agents run independently (exploring repos, analyzing code)
- ✅ Receive results back in n8n
- ✅ Scale multiple concurrent reviews

## Our Solution: Containerized Executors

Instead of trying to embed agents in n8n, we package them as independent services:

```
Claude Executor
├── HTTP API endpoint
├── Receives: repo context, query, KB reference
├── Runs: Claude with agentic loop (code exploration)
└── Returns: analysis + findings

Gemini Executor
├── Similar contract to Claude
├── Parallel execution possible
└── Comparison of results

Each runs in its own container, spawned on-demand
```

### Architecture

```
n8n Workflow
    ↓
POST /execute (lightweight request)
    ↓
Executor Container (Docker/Podman)
    ├→ Initialize Claude agent
    ├→ Explore repository
    ├→ Query AnythingLLM for context
    ├→ Perform code analysis
    ├→ Generate report
    └→ POST back to webhook endpoint
        ↓
n8n captures results → updates GitHub PR
```

### Benefits of Containerization

| Aspect | Benefit |
|--------|---------|
| **Independence** | Agents don't block n8n workflows |
| **Scalability** | Spawn multiple executors concurrently |
| **Isolation** | Each agent gets clean environment |
| **Testability** | Run agents locally, same as production |
| **Lightweight trigger** | n8n just sends HTTP request + waits |
| **Resource control** | Limit CPU/memory per executor |

## How n8n Triggers Them

n8n's HTTP node makes this simple:

1. **Send request**
   ```json
   {
     "repo_url": "https://github.com/org/repo",
     "query": "Security vulnerabilities in authentication",
     "kb_workspace": "repo-123",
     "callback_url": "https://n8n.example.com/webhook/results"
   }
   ```

2. **Executor receives it** → Starts agent
3. **Agent completes analysis** → POSTs back to callback
4. **n8n workflow resumes** → Continues with results

### Why Not a Heavy Orchestrator?

We considered:
- Kubernetes + CRDs (overcomplicated for our scale)
- AWS Step Functions (vendor lock-in)
- Custom Go scheduler (reinventing the wheel)

**Containerized executors + webhooks** gives us:
- ✅ Simple, composable architecture
- ✅ Works with any container runtime (Docker, Podman)
- ✅ Easy to run locally and in production
- ✅ No vendor lock-in

## Example: Claude Agent Executor

```go
// services/claude-executor/main.go
func handleExecutionRequest(w http.ResponseWriter, r *http.Request) {
    // 1. Parse request
    req := parseExecutionRequest(r)

    // 2. Initialize Claude agent with MCP
    agent := claude.NewAgent(
        client,
        tools.WithMCP(mcpConfig),    // Access to KB via MCP
        tools.WithGitHub(githubCfg), // Access to repo
    )

    // 3. Run exploration loop
    result := agent.Explore(ctx, req.Query)

    // 4. POST back to callback
    postResults(req.CallbackURL, result)
}
```

The executor is:
- **Stateless**: Can be replicated across containers
- **Focused**: Specializes in agent execution
- **Observable**: Logs its own execution
- **Testable**: Works offline with mock data

## What's Next?

With this architecture, we can:
- **Add specialized executors** (security-focused, performance-focused, etc.)
- **Compare agents** (Claude vs Gemini on same task)
- **Build consensus** (multiple agents + voting on findings)
- **Create feedback loops** (agent corrections based on review outcomes)

All without modifying n8n or rebuilding orchestration logic.

**Philosophy: Use containers for execution, use n8n for coordination.**
