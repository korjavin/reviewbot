package handler

import (
	"context"
	"fmt"
	"log"

	gh "github.com/google/go-github/v68/github"
)

func HandlePullRequest(newClient ClientFactory, event *gh.PullRequestEvent) {
	if event.GetAction() != "opened" {
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
	prNumber := event.GetPullRequest().GetNumber()
	prTitle := event.GetPullRequest().GetTitle()

	body := fmt.Sprintf("Pong! I see a new PR: **%s**", prTitle)

	ctx := context.Background()
	comment := &gh.IssueComment{Body: &body}
	_, _, err = client.Issues.CreateComment(ctx, owner, repoName, prNumber, comment)
	if err != nil {
		log.Printf("ERROR: commenting on PR #%d: %v", prNumber, err)
		return
	}

	log.Printf("Commented on PR #%d in %s/%s", prNumber, owner, repoName)
}
