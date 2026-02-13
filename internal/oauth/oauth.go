package oauth

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/korjavin/reviewbot/internal/config"
)

func NewCallbackHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		installationID := r.URL.Query().Get("installation_id")
		setupAction := r.URL.Query().Get("setup_action")

		log.Printf("OAuth callback: installation_id=%s, setup_action=%s", installationID, setupAction)

		if code == "" {
			// Installation without OAuth (e.g. setup_action=install with no code).
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, successHTML("ReviewBot installed successfully!"))
			return
		}

		if cfg.ClientID == "" || cfg.ClientSecret == "" {
			log.Printf("OAuth code received but CLIENT_ID/CLIENT_SECRET not configured, skipping token exchange")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, successHTML("ReviewBot installed successfully!"))
			return
		}

		// Exchange code for user access token.
		token, err := exchangeCode(cfg.ClientID, cfg.ClientSecret, code)
		if err != nil {
			log.Printf("ERROR: exchanging OAuth code: %v", err)
			http.Error(w, "OAuth token exchange failed", http.StatusInternalServerError)
			return
		}

		log.Printf("OAuth token obtained (scoped, starts with %s...)", token[:8])

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, successHTML("ReviewBot installed and authorized successfully!"))
	}
}

func exchangeCode(clientID, clientSecret, code string) (string, error) {
	data := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {code},
	}

	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response: %w", err)
	}

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		Scope       string `json:"scope"`
		Error       string `json:"error"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parsing response: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("github error: %s", result.Error)
	}

	return result.AccessToken, nil
}

func successHTML(message string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><title>ReviewBot</title></head>
<body style="font-family:sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;">
<div style="text-align:center;">
<h1>%s</h1>
<p>You can close this window.</p>
</div>
</body>
</html>`, message)
}
