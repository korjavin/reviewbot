package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	gh "github.com/google/go-github/v68/github"
)

// reviewJob is the payload sent to the n8n webhook for async review processing.
type reviewJob struct {
	Owner       string `json:"owner"`
	Repo        string `json:"repo"`
	PRNumber    int    `json:"pr_number"`
	CommentBody string `json:"comment_body"`
	CommentID   int64  `json:"comment_id"`
}

func HandleIssueComment(newClient ClientFactory, newTransport TransportFactory, n8nWebhookURL string, event *gh.IssueCommentEvent) {
	if event.GetAction() != "created" {
		return
	}

	// Ignore comments from bots to avoid infinite loops.
	if event.GetComment().GetUser().GetType() == "Bot" {
		return
	}

	commentBody := event.GetComment().GetBody()
	if !strings.Contains(strings.ToLower(commentBody), "@reviewbot") {
		return
	}

	installationID := event.GetInstallation().GetID()
	client, err := newClient(installationID)
	if err != nil {
		log.Printf("ERROR: creating client for installation %d: %v", installationID, err)
		return
	}

	repo := event.GetRepo()
	owner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()
	issueNumber := event.GetIssue().GetNumber()
	commentID := event.GetComment().GetID()

	ctx := context.Background()

	// Acknowledge with 👀 immediately so the user knows the bot is on it.
	_, _, err = client.Reactions.CreateIssueCommentReaction(ctx, owner, repoName, commentID, "eyes")
	if err != nil {
		log.Printf("ERROR: adding reaction: %v", err)
	}

	if n8nWebhookURL == "" {
		// No n8n configured — fall back to a placeholder comment.
		body := "Review requested. No pipeline configured yet (`N8N_WEBHOOK_URL` not set)."
		reply := &gh.IssueComment{Body: &body}
		_, _, err = client.Issues.CreateComment(ctx, owner, repoName, issueNumber, reply)
		if err != nil {
			log.Printf("ERROR: replying to comment in issue #%d: %v", issueNumber, err)
		}
		return
	}

	// Get a short-lived installation token for the runner to clone with.
	transport, err := newTransport(installationID)
	if err != nil {
		log.Printf("ERROR: creating transport for installation %d: %v", installationID, err)
		return
	}
	token, err := transport.Token(ctx)
	if err != nil {
		log.Printf("ERROR: getting installation token: %v", err)
		return
	}

	job := reviewJob{
		Owner:       owner,
		Repo:        repoName,
		PRNumber:    issueNumber,
		CommentBody: commentBody,
		CommentID:   commentID,
	}

	if err := dispatchToN8N(n8nWebhookURL, token, job); err != nil {
		log.Printf("ERROR: dispatching review job to n8n for issue #%d: %v", issueNumber, err)
		// Let the user know something went wrong.
		body := fmt.Sprintf("Sorry, I couldn't start the review pipeline: `%v`", err)
		reply := &gh.IssueComment{Body: &body}
		_, _, _ = client.Issues.CreateComment(ctx, owner, repoName, issueNumber, reply)
		return
	}

	log.Printf("Dispatched review job to n8n for %s/%s issue #%d", owner, repoName, issueNumber)
}

// dispatchToN8N POSTs a review job to the n8n webhook URL.
// The GitHub token is sent as a Bearer header so n8n can forward it to
// the claude-runner without embedding it in the JSON body.
func dispatchToN8N(webhookURL, githubToken string, job reviewJob) error {
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, webhookURL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+githubToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("n8n returned %d", resp.StatusCode)
	}
	return nil
}
