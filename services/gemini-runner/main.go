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
	// Workdir: if set, skip cloning and run gemini in this pre-cloned directory.
	// The directory is NOT deleted after the run (caller manages lifecycle).
	Workdir string `json:"workdir,omitempty"`
	// Prompt: if set, overrides the default review prompt.
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

	slog.Info("gemini-runner starting", "port", port)
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

	if req.Owner == "" || req.Repo == "" {
		http.Error(w, "owner and repo are required", http.StatusBadRequest)
		return
	}

	slog.Info("review job started", "owner", req.Owner, "repo", req.Repo, "pr", req.PRNumber, "workdir", req.Workdir)

	var workDir string

	if req.Workdir != "" {
		// Use the pre-cloned directory — no cloning needed.
		workDir = req.Workdir
		if _, err := os.Stat(workDir); err != nil {
			slog.Error("workdir not found", "dir", workDir, "err", err)
			http.Error(w, "workdir not found", http.StatusBadRequest)
			return
		}
		slog.Debug("using pre-cloned workdir", "dir", workDir)
	} else {
		// No workdir provided — clone into a temp directory.
		if req.GitHubToken == "" {
			http.Error(w, "github_token is required when workdir is not set", http.StatusBadRequest)
			return
		}
		tmpDir, err := os.MkdirTemp("", "gemini-runner-*")
		if err != nil {
			slog.Error("temp dir creation failed", "err", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		defer os.RemoveAll(tmpDir)
		workDir = tmpDir
		slog.Debug("temp dir created", "dir", workDir)

		cloneURL := fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git",
			req.GitHubToken, req.Owner, req.Repo)
		cloneURLSafe := fmt.Sprintf("https://x-access-token:***@github.com/%s/%s.git", req.Owner, req.Repo)
		slog.Debug("cloning repo", "url", cloneURLSafe, "dir", workDir)
		gitCmd := exec.Command("git", "clone", "--depth=1", cloneURL, ".")
		gitCmd.Dir = workDir
		if out, err := gitCmd.CombinedOutput(); err != nil {
			sanitized := strings.ReplaceAll(string(out), req.GitHubToken, "***")
			slog.Error("git clone failed", "output", sanitized, "err", err)
			http.Error(w, "clone failed", http.StatusInternalServerError)
			return
		} else {
			slog.Debug("git clone succeeded", "output", strings.ReplaceAll(string(out), req.GitHubToken, "***"))
		}
	}

	prompt := req.Prompt
	if prompt == "" {
		prompt = buildPrompt(req.Owner, req.Repo, req.PRNumber)
	}
	slog.Debug("prompt", "text", prompt)

	ctx, cancel := context.WithTimeout(r.Context(), defaultReviewTimeout)
	defer cancel()

	// Run gemini in non-interactive print mode.
	// --yolo: auto-approve all tool calls (required for headless use, no TTY).
	cmdArgs := []string{"-p", prompt, "--yolo"}
	slog.Debug("running gemini", "args", cmdArgs, "dir", workDir)
	geminiCmd := exec.CommandContext(ctx, "gemini", cmdArgs...)
	geminiCmd.Dir = workDir
	geminiCmd.Env = append(os.Environ())

	// Wire Gemini's output to both our buffer (to return to n8n) and stdout (for docker logs).
	var outBuf bytes.Buffer
	multiWriter := io.MultiWriter(&outBuf, os.Stdout)
	geminiCmd.Stdout = multiWriter
	geminiCmd.Stderr = multiWriter

	geminiErr := geminiCmd.Run()
	out := outBuf.Bytes()

	if geminiErr != nil {
		slog.Error("gemini run failed",
			"err", geminiErr,
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

Review the code in the current directory. Examine the file structure and key source files. Focus on:
1. Security vulnerabilities (injections, auth issues, data exposure, unsafe operations)
2. Code quality and maintainability problems
3. Violations of best practices
4. Obvious bugs or logic errors

Output a concise, actionable code review in GitHub-flavored markdown. Reference specific file paths and line numbers where relevant. If the code looks good overall, say so and note any minor suggestions.`, owner, repo, prNumber)
}
