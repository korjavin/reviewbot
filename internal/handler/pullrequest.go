package handler

import (
	"log"

	gh "github.com/google/go-github/v68/github"
)

// HandlePullRequest is called when a PR is opened.
// The pipeline is triggered only by explicit @reviewbot mentions (issue comments),
// so PR opened events are acknowledged and ignored here.
func HandlePullRequest(_ ClientFactory, _ TransportFactory, event *gh.PullRequestEvent) {
	if event.GetAction() != "opened" {
		return
	}
	// Ignore PRs from bots to avoid infinite loops.
	if event.GetPullRequest().GetUser().GetType() == "Bot" {
		return
	}

	repo := event.GetRepo()
	log.Printf("PR opened: %s/%s #%d — pipeline triggered only by @reviewbot mention",
		repo.GetOwner().GetLogin(), repo.GetName(), event.GetPullRequest().GetNumber())
}
