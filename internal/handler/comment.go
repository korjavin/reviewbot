package handler

import (
	"context"
	"fmt"
	"log"
	"strings"

	gh "github.com/google/go-github/v68/github"
)

func HandleIssueComment(newClient ClientFactory, event *gh.IssueCommentEvent) {
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

	// Add "eyes" reaction to the original comment.
	_, _, err = client.Reactions.CreateIssueCommentReaction(ctx, owner, repoName, commentID, "eyes")
	if err != nil {
		log.Printf("ERROR: adding reaction: %v", err)
	}

	// Build a snippet (first 100 chars).
	snippet := commentBody
	if len(snippet) > 100 {
		snippet = snippet[:100] + "..."
	}

	body := fmt.Sprintf("Pong! You mentioned me in: \"%s\"", snippet)
	reply := &gh.IssueComment{Body: &body}
	_, _, err = client.Issues.CreateComment(ctx, owner, repoName, issueNumber, reply)
	if err != nil {
		log.Printf("ERROR: replying to comment in issue #%d: %v", issueNumber, err)
		return
	}

	log.Printf("Replied to mention in issue #%d in %s/%s", issueNumber, owner, repoName)
}
