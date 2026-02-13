package handler

import gh "github.com/google/go-github/v68/github"

// ClientFactory creates a GitHub client for the given installation ID.
type ClientFactory func(installationID int64) (*gh.Client, error)
