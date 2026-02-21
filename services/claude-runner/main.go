package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const defaultReviewTimeout = 10 * time.Minute

// ReviewRequest is sent by n8n to trigger a code review.
type ReviewRequest struct {
	CloneURL    string `json:"clone_url"`
	GitHubToken string `json:"github_token"`
	Owner       string `json:"owner"`
	Repo        string `json:"repo"`
	PRNumber    int    `json:"pr_number"`
	// Optional: override the default prompt.
	Prompt string `json:"prompt,omitempty"`
}

// ReviewResponse is returned to n8n after the review completes.
type ReviewResponse struct {
	Review string `json:"review"`
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})))

	// Seed ~/.claude.json with MCP config on first run.
	seedClaudeConfig()

	mux := http.NewServeMux()
	mux.HandleFunc("POST /review", handleReview)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: defaultReviewTimeout + 30*time.Second,
		IdleTimeout:  60 * time.Second,
	}

	slog.Info("claude-runner starting", "port", port)
	if err := srv.ListenAndServe(); err != nil {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}

func handleReview(w http.ResponseWriter, r *http.Request) {
	var req ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Owner == "" || req.Repo == "" || req.GitHubToken == "" {
		http.Error(w, "owner, repo, and github_token are required", http.StatusBadRequest)
		return
	}

	slog.Info("review job started", "owner", req.Owner, "repo", req.Repo, "pr", req.PRNumber)

	tmpDir, err := os.MkdirTemp("", "claude-runner-*")
	if err != nil {
		slog.Error("temp dir creation failed", "err", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer os.RemoveAll(tmpDir)
	slog.Debug("temp dir created", "dir", tmpDir)

	// Clone the repo with a shallow depth to keep it fast.
	cloneURL := fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git",
		req.GitHubToken, req.Owner, req.Repo)
	cloneURLSafe := fmt.Sprintf("https://x-access-token:***@github.com/%s/%s.git", req.Owner, req.Repo)
	slog.Debug("cloning repo", "url", cloneURLSafe, "dir", tmpDir)
	gitCmd := exec.Command("git", "clone", "--depth=1", cloneURL, ".")
	gitCmd.Dir = tmpDir
	if out, err := gitCmd.CombinedOutput(); err != nil {
		sanitized := strings.ReplaceAll(string(out), req.GitHubToken, "***")
		slog.Error("git clone failed", "output", sanitized, "err", err)
		http.Error(w, "clone failed", http.StatusInternalServerError)
		return
	} else {
		slog.Debug("git clone succeeded", "output", strings.ReplaceAll(string(out), req.GitHubToken, "***"))
	}

	prompt := req.Prompt
	if prompt == "" {
		prompt = buildPrompt(req.Owner, req.Repo, req.PRNumber)
	}
	slog.Debug("prompt", "text", prompt)

	ctx, cancel := context.WithTimeout(r.Context(), defaultReviewTimeout)
	defer cancel()

	// Run claude in non-interactive print mode.
	// --dangerously-skip-permissions: required for headless use (no TTY to confirm tool calls).
	cmdArgs := []string{"-p", prompt, "--dangerously-skip-permissions"}
	slog.Debug("running claude", "args", cmdArgs, "dir", tmpDir)
	claudeCmd := exec.CommandContext(ctx, "/app/bin/claude", cmdArgs...)
	claudeCmd.Dir = tmpDir
	// Prevent update checks and telemetry in headless mode.
	claudeCmd.Env = append(os.Environ(), "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1")

	// Wire Claude's output to both our buffer (to return to n8n) and standard out (for real-time docker logs)
	var outBuf bytes.Buffer
	multiWriter := io.MultiWriter(&outBuf, os.Stdout)
	claudeCmd.Stdout = multiWriter
	claudeCmd.Stderr = multiWriter

	err = claudeCmd.Run()
	out := outBuf.Bytes()

	if err != nil {
		// Log the full output so we can see what claude actually complained about.
		slog.Error("claude run failed",
			"err", err,
			"owner", req.Owner, "repo", req.Repo, "pr", req.PRNumber,
			"output", string(out),
		)
		http.Error(w, "review failed", http.StatusInternalServerError)
		return
	}

	slog.Info("review complete",
		"owner", req.Owner, "repo", req.Repo, "pr", req.PRNumber,
		"review_bytes", len(out),
	)
	slog.Debug("review preview", "first_200_chars", func() string {
		s := string(out)
		if len(s) > 200 {
			return s[:200] + "..."
		}
		return s
	}())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ReviewResponse{Review: string(out)})
}

// buildPrompt returns the default review prompt for a PR.
func buildPrompt(owner, repo string, prNumber int) string {
	return fmt.Sprintf(`You are a code reviewer for GitHub repository %s/%s, reviewing PR #%d.

IMPORTANT: Before using any AnythingLLM MCP tools, you MUST first call the
initialize_anythingllm tool with no arguments — it will auto-configure from
environment variables. Skip this step and all other anythingllm tools will fail.

After initializing, use the AnythingLLM MCP server tools to query the knowledge
base for relevant security guidelines, common vulnerability patterns, and code
quality standards that apply to this type of codebase.

Then review the code in the current directory. Examine the file structure and key source files. Focus on:
1. Security vulnerabilities (injections, auth issues, data exposure, unsafe operations)
2. Code quality and maintainability problems
3. Violations of best practices found in the knowledge base
4. Obvious bugs or logic errors

Output a concise, actionable code review in GitHub-flavored markdown. Reference specific file paths and line numbers where relevant. If the code looks good overall, say so and note any minor suggestions.`, owner, repo, prNumber)
}

// seedClaudeConfig writes MCP server config into ~/.claude.json (user scope).
// The native Claude Code installer reads user-scoped MCP servers from ~/.claude.json
// (NOT from ~/.claude/settings.json which was the old format).
//
// We merge our entries into the existing file so auth tokens (also in ~/.claude.json)
// written by `claude auth` are preserved.
//
// We also pre-approve all anythingllm MCP tools via permissions.allow so that
// headless `claude -p` runs never get stuck on "Do you want to proceed?" prompts.
func seedClaudeConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("could not determine home dir, skipping MCP seed", "err", err)
		return
	}

	cfgPath := filepath.Join(home, ".claude.json")

	// Read existing config (may contain auth tokens written by `claude auth`).
	existing := map[string]any{}
	if raw, err := os.ReadFile(cfgPath); err == nil {
		if err := json.Unmarshal(raw, &existing); err != nil {
			slog.Warn("~/.claude.json is not valid JSON, will overwrite", "err", err)
			existing = map[string]any{}
		}
	}

	// ── MCP server registration ────────────────────────────────────────────────
	servers, _ := existing["mcpServers"].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}

	baseURL := os.Getenv("ANYTHINGLLM_URL")
	if baseURL == "" {
		baseURL = "http://anythingllm:3001"
	}

	servers["anythingllm"] = map[string]any{
		"type":    "stdio",
		"command": "anythingllm-mcp-server",
		"env": map[string]string{
			// Inject actual environment variable values directly into the config.
			// Relying on Claude to expand ${VAR} seems to fail in some versions.
			"ANYTHINGLLM_API_KEY":  os.Getenv("ANYTHINGLLM_API_KEY"),
			"ANYTHINGLLM_BASE_URL": baseURL,
		},
	}
	existing["mcpServers"] = servers

	// ── Pre-approve all anythingllm MCP tool calls ─────────────────────────────
	// Without this, `claude -p` in headless mode halts on "Do you want to proceed?"
	// for every MCP tool invocation, even when --dangerously-skip-permissions is set.
	// The pattern "mcp__anythingllm__*" allows all tools from the anythingllm server.
	perms, _ := existing["permissions"].(map[string]any)
	if perms == nil {
		perms = map[string]any{}
	}
	allow, _ := perms["allow"].([]any)
	// Only add the wildcard if not already present.
	found := false
	for _, v := range allow {
		if s, ok := v.(string); ok && s == "mcp__anythingllm__*" {
			found = true
			break
		}
	}
	if !found {
		allow = append(allow, "mcp__anythingllm__*")
	}
	perms["allow"] = allow
	existing["permissions"] = perms

	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		slog.Warn("could not marshal ~/.claude.json", "err", err)
		return
	}

	if err := os.WriteFile(cfgPath, data, 0600); err != nil {
		slog.Warn("could not write ~/.claude.json", "err", err)
		return
	}

	slog.Info("seeded ~/.claude.json with AnythingLLM MCP config and permissions")
}
