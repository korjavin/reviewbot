package github

import (
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	gh "github.com/google/go-github/v68/github"
	"github.com/korjavin/reviewbot/internal/config"
)

func NewInstallationTransport(cfg *config.Config, installationID int64) (*ghinstallation.Transport, error) {
	if cfg.PrivateKeyPath != "" {
		return ghinstallation.NewKeyFromFile(http.DefaultTransport, cfg.AppID, installationID, cfg.PrivateKeyPath)
	}
	return ghinstallation.New(http.DefaultTransport, cfg.AppID, installationID, []byte(cfg.PrivateKey))
}

func NewInstallationClient(cfg *config.Config, installationID int64) (*gh.Client, error) {
	transport, err := NewInstallationTransport(cfg, installationID)
	if err != nil {
		return nil, fmt.Errorf("creating github transport: %w", err)
	}
	return gh.NewClient(&http.Client{Transport: transport}), nil
}
