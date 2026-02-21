package main

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	AnythingLLMURL    string
	AnythingLLMAPIKey string
	WorkspaceSlug     string
	IntelsDir         string
	StatePath         string
	SyncInterval      time.Duration
}

func LoadConfig() (*Config, error) {
	url := os.Getenv("ANYTHINGLLM_URL")
	if url == "" {
		url = "http://anythingllm:3001"
	}

	key := os.Getenv("ANYTHINGLLM_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("ANYTHINGLLM_API_KEY is required")
	}

	workspace := os.Getenv("ANYTHINGLLM_WORKSPACE")
	if workspace == "" {
		workspace = "intels"
	}

	intelsDir := os.Getenv("INTELS_DIR")
	if intelsDir == "" {
		intelsDir = "/intels"
	}

	statePath := os.Getenv("STATE_PATH")
	if statePath == "" {
		statePath = "/state/kb-maintainer.json"
	}

	syncInterval := 5 * time.Minute
	if s := os.Getenv("SYNC_INTERVAL"); s != "" {
		d, err := time.ParseDuration(s)
		if err != nil {
			return nil, fmt.Errorf("invalid SYNC_INTERVAL %q: %w", s, err)
		}
		syncInterval = d
	}

	return &Config{
		AnythingLLMURL:    url,
		AnythingLLMAPIKey: key,
		WorkspaceSlug:     workspace,
		IntelsDir:         intelsDir,
		StatePath:         statePath,
		SyncInterval:      syncInterval,
	}, nil
}
