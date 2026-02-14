package config

import (
	"os"
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
