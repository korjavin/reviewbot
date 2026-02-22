package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ── Checkout ──────────────────────────────────────────────────────────────────

// CheckoutRequest is POSTed by n8n to clone a repo into the shared volume.
type CheckoutRequest struct {
	Owner       string `json:"owner"`
	Repo        string `json:"repo"`
	GitHubToken string `json:"github_token"`
	PRNumber    int    `json:"pr_number"`
	BaseBranch  string `json:"base_branch"` // optional; detected from clone if empty
}

// CheckoutResponse is returned after a successful checkout.
type CheckoutResponse struct {
	RepoPath      string `json:"repo_path"`
	Branch        string `json:"branch"`
	DefaultBranch string `json:"default_branch"`
}

// HandleCheckout clones a GitHub repo into an isolated directory on the shared
// volume and creates a dedicated review branch. Each call produces a fresh
// directory named {owner}-{repo}-pr{N}-{timestamp} so concurrent reviews of
// the same repo never interfere.
//
// POST /git/checkout
func HandleCheckout(sharedReposDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CheckoutRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.Owner == "" || req.Repo == "" || req.GitHubToken == "" {
			http.Error(w, "owner, repo, and github_token are required", http.StatusBadRequest)
			return
		}

		ts := time.Now().UTC().Format("20060102T150405Z")
		dirName := fmt.Sprintf("%s-%s-pr%d-%s", req.Owner, req.Repo, req.PRNumber, ts)
		repoDir := filepath.Join(sharedReposDir, dirName)

		if err := os.MkdirAll(repoDir, 0o755); err != nil {
			log.Printf("ERROR: creating repo dir %s: %v", repoDir, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		cloneURL := fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git",
			req.GitHubToken, req.Owner, req.Repo)

		log.Printf("Cloning %s/%s into %s", req.Owner, req.Repo, repoDir)
		if out, err := gitRun(repoDir, "git", "clone", "--depth=50", cloneURL, "."); err != nil {
			sanitized := strings.ReplaceAll(string(out), req.GitHubToken, "***")
			log.Printf("ERROR: git clone failed: %s: %v", sanitized, err)
			os.RemoveAll(repoDir)
			http.Error(w, "clone failed", http.StatusInternalServerError)
			return
		}

		// Detect the default branch from the cloned HEAD.
		defaultBranch := req.BaseBranch
		if defaultBranch == "" {
			if out, err := gitRun(repoDir, "git", "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
				defaultBranch = strings.TrimSpace(string(out))
			}
		}
		if defaultBranch == "" {
			defaultBranch = "main"
		}

		// Create the review branch.
		var branchName string
		if req.PRNumber > 0 {
			branchName = fmt.Sprintf("reviewbot-pr%d-%s", req.PRNumber, ts)
		} else {
			branchName = fmt.Sprintf("reviewbot-%s", ts)
		}
		if _, err := gitRun(repoDir, "git", "checkout", "-b", branchName); err != nil {
			log.Printf("ERROR: creating branch %s: %v", branchName, err)
			os.RemoveAll(repoDir)
			http.Error(w, "branch creation failed", http.StatusInternalServerError)
			return
		}

		log.Printf("Checkout complete: %s/%s → %s (branch: %s)", req.Owner, req.Repo, repoDir, branchName)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CheckoutResponse{
			RepoPath:      repoDir,
			Branch:        branchName,
			DefaultBranch: defaultBranch,
		})
	}
}

// ── Create PR ─────────────────────────────────────────────────────────────────

// CreatePRRequest is POSTed by n8n after all CI checks have been committed.
type CreatePRRequest struct {
	RepoPath    string `json:"repo_path"`    // absolute path on shared volume
	Owner       string `json:"owner"`
	Repo        string `json:"repo"`
	GitHubToken string `json:"github_token"`
	Branch      string `json:"branch"`
	Base        string `json:"base"`  // target branch; defaults to "main"
	Title       string `json:"title"` // PR title
	Body        string `json:"body"`  // PR description (markdown)
}

// CreatePRResponse is returned after the PR is created.
type CreatePRResponse struct {
	PRURL    string `json:"pr_url"`
	PRNumber int    `json:"pr_number"`
}

// HandleCreatePR pushes the review branch and opens a GitHub Pull Request.
//
// POST /git/create-pr
func HandleCreatePR() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req CreatePRRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}
		if req.RepoPath == "" || req.Owner == "" || req.Repo == "" ||
			req.GitHubToken == "" || req.Branch == "" {
			http.Error(w, "repo_path, owner, repo, github_token, and branch are required", http.StatusBadRequest)
			return
		}
		if req.Base == "" {
			req.Base = "main"
		}
		if req.Title == "" {
			req.Title = fmt.Sprintf("reviewbot: security CI checks for %s/%s", req.Owner, req.Repo)
		}

		// Push the branch — embed the token in the push URL to avoid storing it.
		pushURL := fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git",
			req.GitHubToken, req.Owner, req.Repo)
		log.Printf("Pushing branch %s to %s/%s", req.Branch, req.Owner, req.Repo)
		if out, err := gitRun(req.RepoPath, "git", "push", pushURL, req.Branch); err != nil {
			sanitized := strings.ReplaceAll(string(out), req.GitHubToken, "***")
			log.Printf("ERROR: git push failed: %s: %v", sanitized, err)
			http.Error(w, "push failed", http.StatusInternalServerError)
			return
		}

		// Create the PR via GitHub REST API.
		type prPayload struct {
			Title string `json:"title"`
			Body  string `json:"body"`
			Head  string `json:"head"`
			Base  string `json:"base"`
		}
		payload, _ := json.Marshal(prPayload{
			Title: req.Title,
			Body:  req.Body,
			Head:  req.Branch,
			Base:  req.Base,
		})

		apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/pulls", req.Owner, req.Repo)
		apiReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, apiURL, bytes.NewReader(payload))
		if err != nil {
			log.Printf("ERROR: building PR request: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		apiReq.Header.Set("Authorization", "Bearer "+req.GitHubToken)
		apiReq.Header.Set("Accept", "application/vnd.github+json")
		apiReq.Header.Set("X-GitHub-Api-Version", "2022-11-28")
		apiReq.Header.Set("Content-Type", "application/json")

		apiResp, err := http.DefaultClient.Do(apiReq)
		if err != nil {
			log.Printf("ERROR: GitHub PR creation failed: %v", err)
			http.Error(w, "GitHub API error", http.StatusBadGateway)
			return
		}
		defer apiResp.Body.Close()

		var ghPR struct {
			Number  int    `json:"number"`
			HTMLURL string `json:"html_url"`
		}
		if err := json.NewDecoder(apiResp.Body).Decode(&ghPR); err != nil || ghPR.Number == 0 {
			log.Printf("ERROR: parsing GitHub PR response (status %d): %v", apiResp.StatusCode, err)
			http.Error(w, fmt.Sprintf("GitHub API returned %d", apiResp.StatusCode), http.StatusBadGateway)
			return
		}

		log.Printf("Created PR #%d: %s", ghPR.Number, ghPR.HTMLURL)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(CreatePRResponse{
			PRURL:    ghPR.HTMLURL,
			PRNumber: ghPR.Number,
		})
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

// gitRun executes a git command in dir and returns combined output.
func gitRun(dir string, name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	return cmd.CombinedOutput()
}
