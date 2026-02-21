package main

import (
	"context"
	"encoding/json"
	"fmt"
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
		Level: slog.LevelInfo,
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

	// Clone the repo with a shallow depth to keep it fast.
	cloneURL := fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git",
		req.GitHubToken, req.Owner, req.Repo)
	gitCmd := exec.Command("git", "clone", "--depth=1", cloneURL, ".")
	gitCmd.Dir = tmpDir
	if out, err := gitCmd.CombinedOutput(); err != nil {
		sanitized := strings.ReplaceAll(string(out), req.GitHubToken, "***")
		slog.Error("git clone failed", "output", sanitized)
		http.Error(w, "clone failed", http.StatusInternalServerError)
		return
	}

	prompt := req.Prompt
	if prompt == "" {
		prompt = buildPrompt(req.Owner, req.Repo, req.PRNumber)
	}

	ctx, cancel := context.WithTimeout(r.Context(), defaultReviewTimeout)
	defer cancel()

	// Run claude in non-interactive print mode.
	// --dangerously-skip-permissions: required for headless use (no TTY to confirm tool calls).
	claudeCmd := exec.CommandContext(ctx, "claude", "-p", prompt, "--dangerously-skip-permissions")
	claudeCmd.Dir = tmpDir
	// Prevent update checks and telemetry in headless mode.
	claudeCmd.Env = append(os.Environ(), "CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1")

	out, err := claudeCmd.Output()
	if err != nil {
		slog.Error("claude run failed", "err", err, "owner", req.Owner, "repo", req.Repo, "pr", req.PRNumber)
		http.Error(w, "review failed", http.StatusInternalServerError)
		return
	}

	slog.Info("review complete",
		"owner", req.Owner, "repo", req.Repo, "pr", req.PRNumber,
		"review_bytes", len(out),
	)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ReviewResponse{Review: string(out)})
}

// buildPrompt returns the default review prompt for a PR.
func buildPrompt(owner, repo string, prNumber int) string {
	return fmt.Sprintf(`You are a code reviewer for GitHub repository %s/%s, reviewing PR #%d.

First, use the AnythingLLM MCP server tools to query the knowledge base for relevant security guidelines, common vulnerability patterns, and code quality standards that apply to this type of codebase.

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
// We merge our mcpServers entry into the existing file if present, so that any
// auth tokens written by `claude auth` (also stored in ~/.claude.json) are preserved.
// The ~/.claude volume mount ensures settings survive container restarts.
func seedClaudeConfig() {
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("could not determine home dir, skipping MCP seed", "err", err)
		return
	}

	cfgPath := filepath.Join(home, ".claude.json")

	anythingllmURL := os.Getenv("ANYTHINGLLM_URL")
	if anythingllmURL == "" {
		anythingllmURL = "http://anythingllm:3001"
	}

	// Read existing config (may contain auth tokens written by `claude auth`).
	existing := map[string]any{}
	if raw, err := os.ReadFile(cfgPath); err == nil {
		if err := json.Unmarshal(raw, &existing); err != nil {
			slog.Warn("~/.claude.json is not valid JSON, will overwrite", "err", err)
			existing = map[string]any{}
		}
	}

	// Check if anythingllm MCP server is already registered.
	if servers, ok := existing["mcpServers"].(map[string]any); ok {
		if _, has := servers["anythingllm"]; has {
			slog.Info("anythingllm MCP server already in ~/.claude.json, skipping seed")
			return
		}
	}

	// Ensure mcpServers map exists.
	servers, _ := existing["mcpServers"].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	servers["anythingllm"] = map[string]any{
		"type":    "stdio",
		"command": "anythingllm-mcp-server",
		"env": map[string]string{
			"ANYTHINGLLM_API_KEY":  os.Getenv("ANYTHINGLLM_API_KEY"),
			"ANYTHINGLLM_BASE_URL": anythingllmURL,
		},
	}
	existing["mcpServers"] = servers

	data, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		slog.Warn("could not marshal ~/.claude.json", "err", err)
		return
	}

	if err := os.WriteFile(cfgPath, data, 0600); err != nil {
		slog.Warn("could not write ~/.claude.json", "err", err)
		return
	}

	slog.Info("seeded ~/.claude.json with AnythingLLM MCP config (user scope)")
}
