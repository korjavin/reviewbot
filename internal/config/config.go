package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	AppID          int64
	PrivateKeyPath string
	PrivateKey     string // raw PEM contents (alternative to path)
	WebhookSecret  string
	ClientID       string
	ClientSecret   string
	Port           string
	BaseURL        string
	// N8NWebhookURL is the n8n webhook endpoint that receives review jobs.
	// When set, @reviewbot mentions are forwarded there instead of replying inline.
	N8NWebhookURL string
	// SharedReposDir is the root of the shared volume where repos are checked out.
	// Must be accessible to both reviewbot and claude-runner containers.
	// Default: /shared/repos
	SharedReposDir string
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

	// Normalize PEM key: Portainer and other UIs strip newlines when pasting
	// multi-line env vars, turning the PEM into a single space-separated line.
	if privateKey != "" {
		privateKey = normalizePEM(privateKey)
	}

	webhookSecret := os.Getenv("GITHUB_WEBHOOK_SECRET")
	if webhookSecret == "" {
		return nil, fmt.Errorf("GITHUB_WEBHOOK_SECRET is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	sharedReposDir := os.Getenv("SHARED_REPOS_DIR")
	if sharedReposDir == "" {
		sharedReposDir = "/shared/repos"
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
		N8NWebhookURL:  os.Getenv("N8N_WEBHOOK_URL"),
		SharedReposDir: sharedReposDir,
	}, nil
}

// normalizePEM fixes PEM keys that lost their newlines (e.g. pasted into
// Portainer env vars). It reconstructs proper PEM format with 64-char lines.
func normalizePEM(key string) string {
	key = strings.TrimSpace(key)

	// If the key already has real newlines, leave it alone.
	if strings.Contains(key, "\n") {
		return key
	}

	// Replace literal \n escape sequences (e.g. from JSON-encoded env vars).
	if strings.Contains(key, `\n`) {
		return strings.ReplaceAll(key, `\n`, "\n")
	}

	// Handle space-separated PEM (Portainer style):
	// "-----BEGIN RSA PRIVATE KEY----- MIIE... -----END RSA PRIVATE KEY-----"
	const beginPrefix = "-----BEGIN "
	const endPrefix = "-----END "

	// Find the header and footer.
	beginIdx := strings.Index(key, beginPrefix)
	if beginIdx == -1 {
		return key
	}
	headerEnd := strings.Index(key[beginIdx:], "-----")
	if headerEnd == -1 {
		return key
	}
	// Find the closing ----- of the header line.
	afterBegin := beginIdx + len(beginPrefix)
	closeDashes := strings.Index(key[afterBegin:], "-----")
	if closeDashes == -1 {
		return key
	}
	headerEndPos := afterBegin + closeDashes + 5
	header := key[beginIdx:headerEndPos]

	endIdx := strings.LastIndex(key, endPrefix)
	if endIdx == -1 {
		return key
	}
	footer := key[endIdx:]

	// Extract the base64 body between header and footer.
	body := strings.TrimSpace(key[headerEndPos:endIdx])
	// Remove all spaces from body.
	body = strings.ReplaceAll(body, " ", "")

	// Rebuild with 64-char lines.
	var lines []string
	lines = append(lines, header)
	for i := 0; i < len(body); i += 64 {
		end := i + 64
		if end > len(body) {
			end = len(body)
		}
		lines = append(lines, body[i:end])
	}
	lines = append(lines, footer)

	return strings.Join(lines, "\n")
}
