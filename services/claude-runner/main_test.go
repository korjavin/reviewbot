package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestHandleReview_WorkdirValidation(t *testing.T) {
	// Create a temporary base directory for testing.
	tmpBase, err := os.MkdirTemp("", "allowed-base-*")
	if err != nil {
		t.Fatalf("failed to create temp base: %v", err)
	}
	defer os.RemoveAll(tmpBase)

	absTmpBase, _ := filepath.Abs(tmpBase)
	os.Setenv("ALLOWED_WORKDIR_BASE", absTmpBase)
	defer os.Unsetenv("ALLOWED_WORKDIR_BASE")

	// Create a valid subdirectory.
	validSubdir := filepath.Join(absTmpBase, "repo1")
	if err := os.Mkdir(validSubdir, 0755); err != nil {
		t.Fatalf("failed to create valid subdir: %v", err)
	}

	tests := []struct {
		name           string
		workdir        string
		expectedStatus int
	}{
		{
			name:           "Valid workdir",
			workdir:        validSubdir,
			expectedStatus: http.StatusInternalServerError, // Fails later because claude is missing, but passes validation
		},
		{
			name:           "Outside base directory",
			workdir:        "/etc",
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Path traversal attempt",
			workdir:        filepath.Join(absTmpBase, "../..", "etc"),
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Relative path outside base",
			workdir:        "../../etc",
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(ReviewRequest{
				Owner:   "owner",
				Repo:    "repo",
				Workdir: tt.workdir,
			})

			req := httptest.NewRequest(http.MethodPost, "/review", bytes.NewBuffer(reqBody))
			rr := httptest.NewRecorder()

			handleReview(rr, req)

			if rr.Code != tt.expectedStatus {
				// If it's InternalServerError, it means it passed validation but failed on executing claude/git.
				// This is acceptable for the purpose of testing the validation logic itself if we don't mock everything.
				// However, if we want to be precise:
				if tt.expectedStatus == http.StatusInternalServerError && rr.Code == http.StatusInternalServerError {
					return
				}

				// Special case: "workdir not found" returns 400.
				if tt.workdir == validSubdir && rr.Code == http.StatusInternalServerError {
					return
				}

				t.Errorf("%s: expected status %d, got %d. Body: %s", tt.name, tt.expectedStatus, rr.Code, rr.Body.String())
			}
		})
	}
}
