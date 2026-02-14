package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/korjavin/reviewbot/internal/config"
)

func TestWebhookRejectsInvalidSignature(t *testing.T) {
	cfg := &config.Config{
		AppID:         123,
		PrivateKey:    "test-key",
		WebhookSecret: "test-secret",
	}

	handler := NewWebhookHandler(cfg)

	body := `{"zen":"test","hook_id":1}`
	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", "sha256=invalidsignature")
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for invalid signature, got %d", rr.Code)
	}
}

func TestWebhookAcceptsValidPing(t *testing.T) {
	secret := "test-secret"
	cfg := &config.Config{
		AppID:         123,
		PrivateKey:    "test-key",
		WebhookSecret: secret,
	}

	handler := NewWebhookHandler(cfg)

	body := `{"zen":"Keep it logically awesome.","hook_id":1}`
	sig := computeHMAC([]byte(body), []byte(secret))

	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", "sha256="+sig)
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 for valid ping, got %d", rr.Code)
	}
}

func TestWebhookMissingSignature(t *testing.T) {
	cfg := &config.Config{
		AppID:         123,
		PrivateKey:    "test-key",
		WebhookSecret: "test-secret",
	}

	handler := NewWebhookHandler(cfg)

	body := `{"zen":"test"}`
	req := httptest.NewRequest("POST", "/webhook", strings.NewReader(body))
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("Content-Type", "application/json")
	// No signature header

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 for missing signature, got %d", rr.Code)
	}
}

func computeHMAC(message, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return hex.EncodeToString(mac.Sum(nil))
}
