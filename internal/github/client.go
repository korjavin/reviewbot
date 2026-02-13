package github

import (
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	gh "github.com/google/go-github/v68/github"
	"github.com/korjavin/reviewbot/internal/config"
)

func NewInstallationClient(cfg *config.Config, installationID int64) (*gh.Client, error) {
	var transport *ghinstallation.Transport
	var err error

	if cfg.PrivateKeyPath != "" {
		transport, err = ghinstallation.NewKeyFromFile(http.DefaultTransport, cfg.AppID, installationID, cfg.PrivateKeyPath)
	} else {
		transport, err = ghinstallation.New(http.DefaultTransport, cfg.AppID, installationID, []byte(cfg.PrivateKey))
	}
	if err != nil {
		return nil, fmt.Errorf("creating github transport: %w", err)
	}

	return gh.NewClient(&http.Client{Transport: transport}), nil
}
