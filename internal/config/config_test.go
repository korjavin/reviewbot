package config

import (
	"os"
	"strings"
	"testing"
)

func TestLoadRequiresAppID(t *testing.T) {
	clearEnv(t)
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when GITHUB_APP_ID is missing")
	}
}

func TestLoadRequiresPrivateKey(t *testing.T) {
	clearEnv(t)
	t.Setenv("GITHUB_APP_ID", "123")
	t.Setenv("GITHUB_WEBHOOK_SECRET", "secret")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when neither private key option is set")
	}
}

func TestLoadRequiresWebhookSecret(t *testing.T) {
	clearEnv(t)
	t.Setenv("GITHUB_APP_ID", "123")
	t.Setenv("GITHUB_PRIVATE_KEY", "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error when GITHUB_WEBHOOK_SECRET is missing")
	}
}

func TestLoadAppIDMustBeNumeric(t *testing.T) {
	clearEnv(t)
	t.Setenv("GITHUB_APP_ID", "notanumber")
	t.Setenv("GITHUB_PRIVATE_KEY", "test")
	t.Setenv("GITHUB_WEBHOOK_SECRET", "secret")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for non-numeric GITHUB_APP_ID")
	}
}

func TestLoadSuccess(t *testing.T) {
	clearEnv(t)
	t.Setenv("GITHUB_APP_ID", "123")
	t.Setenv("GITHUB_PRIVATE_KEY", "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----")
	t.Setenv("GITHUB_WEBHOOK_SECRET", "secret")
	t.Setenv("PORT", "9090")
	t.Setenv("BASE_URL", "https://example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.AppID != 123 {
		t.Errorf("expected AppID 123, got %d", cfg.AppID)
	}
	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.BaseURL != "https://example.com" {
		t.Errorf("expected BaseURL https://example.com, got %s", cfg.BaseURL)
	}
}

func TestLoadDefaultPort(t *testing.T) {
	clearEnv(t)
	t.Setenv("GITHUB_APP_ID", "123")
	t.Setenv("GITHUB_PRIVATE_KEY", "test")
	t.Setenv("GITHUB_WEBHOOK_SECRET", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
}

func TestLoadPrivateKeyPathAlternative(t *testing.T) {
	clearEnv(t)
	t.Setenv("GITHUB_APP_ID", "123")
	t.Setenv("GITHUB_PRIVATE_KEY_PATH", "/some/path.pem")
	t.Setenv("GITHUB_WEBHOOK_SECRET", "secret")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.PrivateKeyPath != "/some/path.pem" {
		t.Errorf("expected PrivateKeyPath /some/path.pem, got %s", cfg.PrivateKeyPath)
	}
}

func TestNormalizePEMAlreadyValid(t *testing.T) {
	valid := "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA\ntest\n-----END RSA PRIVATE KEY-----"
	result := normalizePEM(valid)
	if result != valid {
		t.Errorf("normalizePEM should not modify valid PEM.\ngot:  %q\nwant: %q", result, valid)
	}
}

func TestNormalizePEMSpaceSeparated(t *testing.T) {
	// Simulates what Portainer does: newlines become spaces.
	spaced := "-----BEGIN RSA PRIVATE KEY----- MIIEpAIBAAKCAQEA dGVzdA== -----END RSA PRIVATE KEY-----"
	result := normalizePEM(spaced)

	if !strings.HasPrefix(result, "-----BEGIN RSA PRIVATE KEY-----\n") {
		t.Errorf("expected header followed by newline, got: %q", result[:60])
	}
	if !strings.HasSuffix(result, "\n-----END RSA PRIVATE KEY-----") {
		t.Errorf("expected footer preceded by newline, got: %q", result[len(result)-60:])
	}
	// Body should have no spaces.
	lines := strings.Split(result, "\n")
	for i, line := range lines {
		if i == 0 || i == len(lines)-1 {
			continue // skip header/footer
		}
		if strings.Contains(line, " ") {
			t.Errorf("body line %d contains spaces: %q", i, line)
		}
	}
}

func TestNormalizePEMEscapedNewlines(t *testing.T) {
	// Simulates JSON-encoded key with literal \n.
	escaped := `-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA\ndGVzdA==\n-----END RSA PRIVATE KEY-----`
	result := normalizePEM(escaped)

	if strings.Contains(result, `\n`) {
		t.Errorf("result still contains literal backslash-n: %q", result)
	}
	lines := strings.Split(result, "\n")
	if len(lines) != 4 {
		t.Errorf("expected 4 lines, got %d: %q", len(lines), result)
	}
}

func TestNormalizePEMBodyLineLength(t *testing.T) {
	// Create a key with a long body that should be split into 64-char lines.
	body := strings.Repeat("A", 200)
	spaced := "-----BEGIN RSA PRIVATE KEY----- " + body + " -----END RSA PRIVATE KEY-----"
	result := normalizePEM(spaced)

	lines := strings.Split(result, "\n")
	// Skip header (first) and footer (last).
	for i := 1; i < len(lines)-1; i++ {
		if len(lines[i]) > 64 {
			t.Errorf("body line %d exceeds 64 chars (%d): %q", i, len(lines[i]), lines[i])
		}
	}
}

func clearEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"GITHUB_APP_ID", "GITHUB_PRIVATE_KEY_PATH", "GITHUB_PRIVATE_KEY",
		"GITHUB_WEBHOOK_SECRET", "GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET",
		"PORT", "BASE_URL",
	} {
		os.Unsetenv(key)
	}
}
