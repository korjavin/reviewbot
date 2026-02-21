---
layout: post
title: "MCP AnythingLLM: Seamless KB Access for Agents"
date: 2026-02-21 11:00:00 -0000
categories: architecture integration
excerpt: "How Model Context Protocol (MCP) connects our agents to AnythingLLM knowledge bases."
---

## The Challenge: Agents Need KB Access

Our Claude/Gemini agents need to:
1. Receive code to review
2. Query AnythingLLM for relevant knowledge
3. Use that context to inform their analysis

Without tight integration, agents need to:
- Make direct HTTP calls to AnythingLLM REST API
- Parse responses, format contexts
- Manage rate limiting, error handling
- Custom code for each executor

This leads to duplication and tight coupling.

## Enter Model Context Protocol (MCP)

**MCP** is an open standard for connecting AI models to resources:

> MCP standardizes how applications can provide context and tools to AI models.

Instead of agents calling APIs directly, MCP acts as a standardized interface:

```
Claude Agent
    ↓
MCP Client (built into executor)
    ↓
MCP Servers (we write/configure these)
    ├→ AnythingLLM Server
    ├→ GitHub Server
    └→ ... other tools
```

## Our MCP Server: `mcp-anythingllm`

We built a lightweight MCP server that exposes AnythingLLM as tools:

```go
// mcp-anythingllm/server.go
type Server struct {
    client *anythingllm.Client
}

// Agents call these via MCP
func (s *Server) SearchWorkspace(query string) []Document
func (s *Server) GetDocument(id string) Document
func (s *Server) ListWorkspaces() []Workspace
func (s *Server) ChatWithContext(query string, workspace string) string
```

### How It Works

In a Claude executor:

```go
executor := NewClaudeExecutor()

// Automatically available to Claude as tools
executor.AddMCPServer("anythingllm",
    "mcp-anythingllm",
    []string{"localhost:3000"},
)

// Claude can now use these naturally:
// "Search for JWT vulnerability patterns in workspace 'auth-service'"
// "Get the full document about OAuth2 implementation"
// etc.
```

Claude accesses the KB **natively** through MCP, without custom code.

## Architecture

```
Claude Executor Container
├── Claude API client
├── MCP Client
│   └→ mcp-anythingllm Server
│       └→ AnythingLLM HTTP API
│           └→ Vector DB + Documents
│
├── GitHub integration (also via MCP)
└── Output to n8n webhook
```

### Key Benefits

| Aspect | Benefit |
|--------|---------|
| **Standard interface** | Claude, Gemini, and future agents use same tools |
| **No custom API code** | MCP handles transport, serialization |
| **Tool discovery** | Agents introspect available tools automatically |
| **Consistent error handling** | MCP standardizes failures |
| **Future-proof** | When we add new KB systems, just add MCP server |

## Example: Agent Exploration

When a Claude agent reviews code:

```
Claude: "I need to understand the authentication patterns"
  ↓
(Internally uses MCP)
  ↓
mcp-anythingllm.SearchWorkspace("authentication patterns")
  ↓
Returns: [doc1.md (Session auth), doc2.md (JWT), doc3.md (OAuth)]
  ↓
Claude: "Found 3 relevant docs. Analyzing for security..."
```

No explicit API calls. No boilerplate. Just agents using tools.

## Not Exposed Externally

Important design choice: **mcp-anythingllm is internal only**.

- 🔴 NOT exposed to GitHub
- 🔴 NOT exposed to n8n directly
- 🔴 Only accessible within executor containers

This keeps:
- ✅ Security boundaries clean
- ✅ KB access controlled (through containers)
- ✅ Executor isolation intact
- ✅ Simple deployment model (no extra services needed)

## What's Next?

With MCP, we can easily add:
- **GitHub MCP server** - Agents browse repos natively
- **Vector search MCP** - Direct embedding queries
- **Custom analysis MCP** - Domain-specific tools
- **Feedback loop MCP** - Agents improve KB

All without modifying Claude/Gemini code.

## Why This Matters

MCP exemplifies our philosophy:
1. Use proven standards (MCP is Anthropic's standard)
2. Minimize custom code (MCP handles the plumbing)
3. Enable composition (multiple MCP servers work together)
4. Future-proof (built on open protocol)

The result: Claude agents that feel native to our platform, with seamless KB access and tool discovery.

**Philosophy: Use standards (MCP) to glue components together.**
