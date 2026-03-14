package handler

import (
	"bytes"
	"log"
	"strings"
	"testing"

	gh "github.com/google/go-github/v68/github"
)

func TestHandlePullRequest(t *testing.T) {
	tests := []struct {
		name           string
		action         string
		userType       string
		wantLog        bool
		expectedLogMsg string
	}{
		{
			name:     "action not opened",
			action:   "closed",
			userType: "User",
			wantLog:  false,
		},
		{
			name:     "user is bot",
			action:   "opened",
			userType: "Bot",
			wantLog:  false,
		},
		{
			name:           "valid opened PR",
			action:         "opened",
			userType:       "User",
			wantLog:        true,
			expectedLogMsg: "PR opened: owner/repo #1 — pipeline triggered only by @reviewbot mention",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			oldOutput := log.Writer()
			log.SetOutput(&buf)
			defer log.SetOutput(oldOutput)

			event := &gh.PullRequestEvent{
				Action: &tt.action,
				PullRequest: &gh.PullRequest{
					Number: gh.Ptr(1),
					User: &gh.User{
						Type: &tt.userType,
					},
				},
				Repo: &gh.Repository{
					Name: gh.Ptr("repo"),
					Owner: &gh.User{
						Login: gh.Ptr("owner"),
					},
				},
			}

			HandlePullRequest(nil, nil, event)

			gotLog := buf.String()
			if tt.wantLog {
				if gotLog == "" {
					t.Errorf("expected log output, got none")
				}
				if tt.expectedLogMsg != "" && !strings.Contains(gotLog, tt.expectedLogMsg) {
					t.Errorf("expected log message to contain %q, got %q", tt.expectedLogMsg, gotLog)
				}
			} else {
				if gotLog != "" {
					t.Errorf("expected no log output, got %q", gotLog)
				}
			}
		})
	}
}
