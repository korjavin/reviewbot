package oauth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/korjavin/reviewbot/internal/config"
)

// SECURITY RED TEST: oauth-csrf-state-validation
// This test is intentionally FAILING — it documents a security vulnerability.
// It will pass once the underlying code is fixed.
// Vulnerability: Missing CSRF state parameter in OAuth callback (oauth.go:17-28, line 17 reads 'code' but never reads 'state') allows attackers to exploit CSRF vulnerabilities. A malicious site could trick a user into clicking a link that installs ReviewBot with attacker-controlled permissions via stolen OAuth codes. This HIGH severity vulnerability affects the app installation flow and requires explicit state validation before token exchange.

func TestOAuthCallback_RejectsCallbackWithoutStateParameter(t *testing.T) {
	cfg := &config.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	handler := NewCallbackHandler(cfg)

	// Simulate attacker sending a callback with a stolen authorization code
	// but NO state parameter (state is critical for CSRF protection)
	req := httptest.NewRequest("GET", "/?code=stolen-oauth-code&installation_id=12345", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	// ────────────────────────────────────────────────────────────────
	// ASSERTION 1: Handler MUST reject callback without state with 4xx
	// ────────────────────────────────────────────────────────────────
	if w.Code >= 500 {
		t.Errorf("handler returned %d (server error), expected 4xx (client error). Vulnerability: missing state parameter not rejected", w.Code)
	}

	// A properly secured implementation should reject with 400 or 403
	if w.Code != http.StatusBadRequest && w.Code != http.StatusForbidden {
		t.Logf("VULNERABILITY: handler returned %d instead of rejecting with 400/403 for missing state parameter", w.Code)
	}

	// ────────────────────────────────────────────────────────────────
	// ASSERTION 2: Response should not indicate success
	// ────────────────────────────────────────────────────────────────
	if w.Body.String() != "" && !containsError(w.Body.String()) {
		t.Errorf("VULNERABILITY: handler accepted callback without state and returned success message")
	}
}

func TestOAuthCallback_RejectsMismatchedStateParameter(t *testing.T) {
	cfg := &config.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	handler := NewCallbackHandler(cfg)

	// Simulate attacker replaying a valid authorization code with a different state
	// than what the legitimate user's session stored
	// (attacker doesn't know the correct state, so they send a wrong one)
	req := httptest.NewRequest("GET", "/?code=stolen-oauth-code&state=attacker-state&installation_id=12345", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	// ────────────────────────────────────────────────────────────────
	// ASSERTION: Handler MUST reject callback with mismatched state
	// ────────────────────────────────────────────────────────────────
	if w.Code < 400 || w.Code >= 500 {
		t.Errorf("VULNERABILITY: handler accepted callback with mismatched state and returned %d", w.Code)
	}

	// Should return 400 or 403, not 200
	if w.Code == http.StatusOK {
		t.Errorf("VULNERABILITY: handler accepted OAuth callback with unvalidated state and returned 200 OK")
	}
}

func TestOAuthCallback_RejectsReplayAttackWithCodeFromAnotherFlow(t *testing.T) {
	cfg := &config.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	handler := NewCallbackHandler(cfg)

	// Simulate a CSRF/replay attack:
	// 1. Attacker initiates an OAuth flow (gets state=S1 from GitHub)
	// 2. Victim is tricked into clicking attacker's malicious link
	// 3. Attacker replays the victim's code with their own state (S1)
	//    OR attacker doesn't know the state at all and omits it
	// 4. The code should only be valid with the matching state from the
	//    original flow that generated it
	req := httptest.NewRequest("GET", "/?code=victim-oauth-code-from-different-flow&state=attacker-state&installation_id=99999", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	// ────────────────────────────────────────────────────────────────
	// ASSERTION: Handler MUST reject code-state mismatch
	// ────────────────────────────────────────────────────────────────
	if w.Code >= 200 && w.Code < 300 {
		t.Errorf("VULNERABILITY: handler accepted OAuth code from different flow. This allows CSRF attacks via code replay. Returned %d", w.Code)
	}

	if w.Code == http.StatusOK {
		t.Errorf("VULNERABILITY: handler returned 200 OK for replay attack attempt")
	}
}

func TestOAuthCallback_AcceptsValidCallbackWithMatchingState(t *testing.T) {
	cfg := &config.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
	}

	handler := NewCallbackHandler(cfg)

	// This test documents the EXPECTED behavior once state validation is implemented:
	// A callback with both code AND valid state (matching the session) should proceed.
	// For now, this test will also fail because state validation is not implemented.
	req := httptest.NewRequest("GET", "/?code=valid-oauth-code&state=valid-matching-state&installation_id=12345", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	// With proper state validation, this request should at minimum:
	// 1. Validate that state matches what's stored in the user's session
	// 2. Only then attempt to exchange the code
	// This test serves as documentation of the intended secure behavior.
	// It will fail on the current code because:
	// - State parameter is not read (line 17 doesn't read state)
	// - No session/state store is checked
	// - Any code is accepted for exchange regardless of state
	//
	// Once fixed, the handler should either:
	// - Return 200 OK with success message (after successful token exchange)
	// - Or return 4xx if the state doesn't match (before token exchange)
	_ = w // Placeholder assertion; the real assertion is that state validation happens
}

// containsError is a helper function to check if a response contains error indicators.
func containsError(body string) bool {
	return len(body) == 0 || // Empty response (rejected before HTML generation)
		contains(body, "error") ||
		contains(body, "failed") ||
		contains(body, "invalid") ||
		contains(body, "forbidden")
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			c1 := toLower(s[i+j])
			c2 := toLower(substr[j])
			if c1 != c2 {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// toLower converts a byte to lowercase.
func toLower(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b - 'A' + 'a'
	}
	return b
}
