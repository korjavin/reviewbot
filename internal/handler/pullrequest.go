package handler

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	gh "github.com/google/go-github/v68/github"
	"github.com/korjavin/reviewbot/internal/git"
)

func HandlePullRequest(newClient ClientFactory, newTransport TransportFactory, event *gh.PullRequestEvent) {
	if event.GetAction() != "opened" {
		return
	}

	// Prevent infinite loops: ignore PRs created by bots.
	if event.GetPullRequest().GetUser().GetType() == "Bot" {
		log.Printf("Ignoring PR from bot user %s", event.GetPullRequest().GetUser().GetLogin())
		return
	}

	installationID := event.GetInstallation().GetID()

	transport, err := newTransport(installationID)
	if err != nil {
		log.Printf("ERROR: creating transport for installation %d: %v", installationID, err)
		return
	}

	ctx := context.Background()

	token, err := transport.Token(ctx)
	if err != nil {
		log.Printf("ERROR: getting installation token: %v", err)
		return
	}

	client, err := newClient(installationID)
	if err != nil {
		log.Printf("ERROR: creating client for installation %d: %v", installationID, err)
		return
	}

	repo := event.GetRepo()
	owner := repo.GetOwner().GetLogin()
	repoName := repo.GetName()
	prNumber := event.GetPullRequest().GetNumber()
	defaultBranch := repo.GetDefaultBranch()

	// Create temp directory for clone.
	tmpDir, err := os.MkdirTemp("", "reviewbot-*")
	if err != nil {
		log.Printf("ERROR: creating temp dir: %v", err)
		return
	}
	defer os.RemoveAll(tmpDir)

	// Clone the repo.
	log.Printf("Cloning %s/%s into %s", owner, repoName, tmpDir)
	if err := git.Clone(tmpDir, token, owner, repoName); err != nil {
		log.Printf("ERROR: cloning repo: %v", err)
		return
	}

	// Create a new branch.
	branchName := fmt.Sprintf("reviewbot/review-pr-%d", prNumber)
	if err := git.CreateBranch(tmpDir, branchName); err != nil {
		log.Printf("ERROR: creating branch: %v", err)
		return
	}

	// Write a test file.
	dateStr := time.Now().UTC().Format("2006-01-02")
	fileName := fmt.Sprintf("%s.txt", dateStr)
	filePath := fmt.Sprintf("%s/%s", tmpDir, fileName)
	content := fmt.Sprintf("ReviewBot was here. PR #%d reviewed on %s.\n", prNumber, dateStr)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		log.Printf("ERROR: writing test file: %v", err)
		return
	}

	// Stage, commit, push.
	if err := git.Add(tmpDir, fileName); err != nil {
		log.Printf("ERROR: git add: %v", err)
		return
	}

	commitMsg := fmt.Sprintf("reviewbot: add review file for PR #%d", prNumber)
	if err := git.Commit(tmpDir, commitMsg); err != nil {
		log.Printf("ERROR: git commit: %v", err)
		return
	}

	if err := git.Push(tmpDir, token, owner, repoName, branchName); err != nil {
		log.Printf("ERROR: git push: %v", err)
		return
	}

	log.Printf("Pushed branch %s to %s/%s", branchName, owner, repoName)

	// Create a new PR via the GitHub API.
	newPRTitle := fmt.Sprintf("reviewbot: review for PR #%d", prNumber)
	newPRBody := fmt.Sprintf("Automated review branch for PR #%d.", prNumber)
	newPR, _, err := client.PullRequests.Create(ctx, owner, repoName, &gh.NewPullRequest{
		Title: &newPRTitle,
		Body:  &newPRBody,
		Head:  &branchName,
		Base:  &defaultBranch,
	})
	if err != nil {
		log.Printf("ERROR: creating PR: %v", err)
		return
	}

	log.Printf("Created PR #%d in %s/%s", newPR.GetNumber(), owner, repoName)

	// Comment on the original PR with a link to the new one.
	commentBody := fmt.Sprintf("I've created a review branch: %s", newPR.GetHTMLURL())
	comment := &gh.IssueComment{Body: &commentBody}
	_, _, err = client.Issues.CreateComment(ctx, owner, repoName, prNumber, comment)
	if err != nil {
		log.Printf("ERROR: commenting on PR #%d: %v", prNumber, err)
		return
	}

	log.Printf("Commented on PR #%d with link to PR #%d", prNumber, newPR.GetNumber())
}
