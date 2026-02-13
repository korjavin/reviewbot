package handler

import (
	"github.com/bradleyfalzon/ghinstallation/v2"
	gh "github.com/google/go-github/v68/github"
)

// ClientFactory creates a GitHub client for the given installation ID.
type ClientFactory func(installationID int64) (*gh.Client, error)

// TransportFactory creates a GitHub App installation transport for token extraction.
type TransportFactory func(installationID int64) (*ghinstallation.Transport, error)
