package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	AppID            int64
	PrivateKeyPath   string
	PrivateKey       string // raw PEM contents (alternative to path)
	WebhookSecret    string
	ClientID         string
	ClientSecret     string
	Port             string
	BaseURL          string
}

func Load() (*Config, error) {
	appIDStr := os.Getenv("GITHUB_APP_ID")
	if appIDStr == "" {
		return nil, fmt.Errorf("GITHUB_APP_ID is required")
	}
	appID, err := strconv.ParseInt(appIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("GITHUB_APP_ID must be a number: %w", err)
	}

	privateKeyPath := os.Getenv("GITHUB_PRIVATE_KEY_PATH")
	privateKey := os.Getenv("GITHUB_PRIVATE_KEY")
	if privateKeyPath == "" && privateKey == "" {
		return nil, fmt.Errorf("GITHUB_PRIVATE_KEY_PATH or GITHUB_PRIVATE_KEY is required")
	}

	webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return nil, fmt.Errorf("GITHUB_WEBHOOK_SECRET is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		AppID:          appID,
		PrivateKeyPath: privateKeyPath,
		PrivateKey:     privateKey,
		WebhookSecret:  webhookSecret,
		ClientID:       os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret:   os.Getenv("GITHUB_CLIENT_SECRET"),
		Port:           port,
		BaseURL:        os.Getenv("BASE_URL"),
	}, nil
}
