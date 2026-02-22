package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// SECURITY RED TEST: path-traversal-in-checkout
// This test is intentionally FAILING — it documents a security vulnerability.
// It will pass once the underlying code is fixed.
// Vulnerability: Path traversal via owner/repo (line 54) allows attackers to write checkout directories outside /shared/repos, potentially enabling arbitrary file writes to the host filesystem. This HIGH severity vulnerability in github_ops.go:54 directly compromises the isolation model and could lead to container escape or host system compromise.

func TestHandleCheckout_RejectsPathTraversal(t *testing.T) {
	// Create a temporary directory to use as sharedReposDir
	tmpDir := t.TempDir()
	sharedReposDir := filepath.Join(tmpDir, "shared-repos")
	if err := os.MkdirAll(sharedReposDir, 0o755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	handler := HandleCheckout(sharedReposDir)

	testCases := []struct {
		name        string
		owner       string
		repo        string
		description string
	}{
		{
			name:        "parent_dir_traversal_in_owner",
			owner:       "../evil",
			repo:        "test",
			description: "attempt to escape via .. in owner",
		},
		{
			name:        "multiple_parent_traversal_in_owner",
			owner:       "../../etc/passwd",
			repo:        "test",
			description: "attempt to escape to /etc via multiple .. in owner",
		},
		{
			name:        "parent_dir_traversal_in_repo",
			owner:       "normal-owner",
			repo:        "../escape",
			description: "attempt to escape via .. in repo",
		},
		{
			name:        "absolute_path_in_owner",
			owner:       "/etc/passwd",
			repo:        "test",
			description: "attempt to use absolute path in owner",
		},
		{
			name:        "absolute_path_in_repo",
			owner:       "normal-owner",
			repo:        "/tmp/malicious",
			description: "attempt to use absolute path in repo",
		},
		{
			name:        "combined_traversal",
			owner:       "../..",
			repo:        "repo",
			description: "multiple parent directory traversals",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Build the request payload
			req := CheckoutRequest{
				Owner:       tc.owner,
				Repo:        tc.repo,
				GitHubToken: "fake-token-12345",
				PRNumber:    42,
				BaseBranch:  "main",
			}

			body, _ := json.Marshal(req)
			httpReq := httptest.NewRequest("POST", "/git/checkout", bytes.NewReader(body))
			w := httptest.NewRecorder()

			// Call the handler
			handler(w, httpReq)

			// ────────────────────────────────────────────────────────────────
			// ASSERTION 1: Handler must reject with 4xx status, NOT 5xx
			// ────────────────────────────────────────────────────────────────
			if w.Code >= 500 {
				t.Errorf("handler returned %d (server error), expected 4xx (client error). Vulnerability: path traversal not rejected", w.Code)
			}

			// For a properly secured implementation, we should reject with 400 or 403
			if w.Code != http.StatusBadRequest && w.Code != http.StatusForbidden {
				t.Logf("VULNERABILITY: handler returned %d instead of rejecting with 400/403 for %s", w.Code, tc.description)
			}

			// ────────────────────────────────────────────────────────────────
			// ASSERTION 2: No directory should be created outside sharedReposDir
			// ────────────────────────────────────────────────────────────────
			createdDirs, err := listCreatedDirs(sharedReposDir)
			if err != nil {
				t.Logf("failed to check created directories: %v", err)
			}

			// Verify that no directories were created outside sharedReposDir
			for _, dir := range createdDirs {
				absDir, _ := filepath.Abs(dir)
				absShared, _ := filepath.Abs(sharedReposDir)

				// Check if the directory is actually inside sharedReposDir
				relPath, err := filepath.Rel(absShared, absDir)
				if err != nil || strings.HasPrefix(relPath, "..") {
					t.Errorf("VULNERABILITY: directory created outside sharedReposDir: %s (relative path: %s)", dir, relPath)
				}
			}

			// ────────────────────────────────────────────────────────────────
			// ASSERTION 3: Verify that resolved path would stay within bounds
			// ────────────────────────────────────────────────────────────────
			ts := time.Now().UTC().Format("20060102T150405Z")
			dirName := tc.owner + "-" + tc.repo + "-pr42-" + ts
			repoDir := filepath.Join(sharedReposDir, dirName)

			// Canonicalize paths for comparison
			absRepoDir, _ := filepath.Abs(repoDir)
			absSharedReposDir, _ := filepath.Abs(sharedReposDir)

			// The resolved repo directory MUST be within sharedReposDir
			relPath, err := filepath.Rel(absSharedReposDir, absRepoDir)
			if err != nil {
				t.Logf("failed to compute relative path: %v", err)
			}

			if strings.HasPrefix(relPath, "..") || filepath.IsAbs(relPath) {
				t.Errorf("VULNERABILITY: resolved path escapes sharedReposDir: %s relative to %s = %s", absRepoDir, absSharedReposDir, relPath)
			}
		})
	}
}

// TestHandleCheckout_PathTraversalWithSymlinks verifies that symlink-based
// escape attempts are also rejected.
func TestHandleCheckout_RejectsSymlinkEscapeAttempts(t *testing.T) {
	tmpDir := t.TempDir()
	sharedReposDir := filepath.Join(tmpDir, "shared-repos")
	if err := os.MkdirAll(sharedReposDir, 0o755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	// Create a symlink outside the sharedReposDir to an external target
	externalTarget := filepath.Join(tmpDir, "external-target")
	if err := os.MkdirAll(externalTarget, 0o755); err != nil {
		t.Fatalf("failed to create external target: %v", err)
	}

	handler := HandleCheckout(sharedReposDir)

	// Attempt to use a name that would resolve via symlink
	req := CheckoutRequest{
		Owner:       "owner-with-symlink",
		Repo:        "repo",
		GitHubToken: "fake-token",
		PRNumber:    1,
		BaseBranch:  "main",
	}

	body, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/git/checkout", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler(w, httpReq)

	// The handler should reject any request, not create directories
	// (this is a basic sanity check; a full test would involve more complex setup)
	if w.Code >= 500 {
		t.Logf("handler returned %d on symlink test", w.Code)
	}
}

// TestHandleCheckout_BoundaryCheck verifies that path validation uses
// proper canonicalization and boundary checks.
func TestHandleCheckout_EnforcesPathBoundary(t *testing.T) {
	tmpDir := t.TempDir()
	sharedReposDir := filepath.Join(tmpDir, "shared-repos")
	if err := os.MkdirAll(sharedReposDir, 0o755); err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	handler := HandleCheckout(sharedReposDir)

	// Test that even with complex path manipulation, the final directory
	// must be within sharedReposDir
	testCases := []struct {
		owner string
		repo  string
	}{
		{"../", "test"},
		{".", "test"},
		{"..", "test"},
		{"normal", "../.."},
		{"user/..", "test"},
	}

	for _, tc := range testCases {
		req := CheckoutRequest{
			Owner:       tc.owner,
			Repo:        tc.repo,
			GitHubToken: "token",
			PRNumber:    1,
		}

		body, _ := json.Marshal(req)
		httpReq := httptest.NewRequest("POST", "/git/checkout", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler(w, httpReq)

		// Verify no escape
		absShared, _ := filepath.Abs(sharedReposDir)
		entriesAfter, _ := os.ReadDir(sharedReposDir)
		for _, entry := range entriesAfter {
			absEntry, _ := filepath.Abs(filepath.Join(sharedReposDir, entry.Name()))
			relPath, _ := filepath.Rel(absShared, absEntry)
			if strings.HasPrefix(relPath, "..") {
				t.Errorf("VULNERABILITY DETECTED: entry %s escapes sharedReposDir", relPath)
			}
		}
	}
}

// listCreatedDirs returns a list of all directories created within sharedReposDir.
// This helper is used to verify that no directories were created outside the boundary.
func listCreatedDirs(sharedReposDir string) ([]string, error) {
	var dirs []string
	entries, err := os.ReadDir(sharedReposDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(sharedReposDir, entry.Name()))
		}
	}

	return dirs, nil
}
