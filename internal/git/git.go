package git

import (
	"fmt"
	"os/exec"
	"strings"
)

// Clone performs a shallow clone using HTTPS with an access token.
func Clone(dir, token, owner, repo string) error {
	url := fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git", token, owner, repo)
	err := run(dir, "git", "clone", "--depth=1", url, ".")
	if err != nil {
		return sanitizeErr(err, token)
	}
	return nil
}

// CreateBranch creates and checks out a new branch.
func CreateBranch(dir, branch string) error {
	return run(dir, "git", "checkout", "-b", branch)
}

// Add stages files.
func Add(dir string, paths ...string) error {
	args := append([]string{"add"}, paths...)
	return run(dir, "git", args...)
}

// Commit creates a commit with the given message.
func Commit(dir, message string) error {
	if err := run(dir, "git", "-c", "user.name=reviewbot[bot]", "-c", "user.email=reviewbot[bot]@users.noreply.github.com", "commit", "-m", message); err != nil {
		return err
	}
	return nil
}

// Push pushes the current branch to origin.
func Push(dir, token, owner, repo, branch string) error {
	url := fmt.Sprintf("https://x-access-token:%s@github.com/%s/%s.git", token, owner, repo)
	err := run(dir, "git", "push", url, branch)
	if err != nil {
		return sanitizeErr(err, token)
	}
	return nil
}

func run(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w\n%s", name, strings.Join(args, " "), err, string(out))
	}
	return nil
}

// sanitizeErr strips the token from error messages to avoid leaking it in logs.
func sanitizeErr(err error, token string) error {
	return fmt.Errorf("%s", strings.ReplaceAll(err.Error(), token, "***"))
}
