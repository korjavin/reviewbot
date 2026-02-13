package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/korjavin/reviewbot/internal/config"
	ghwebhook "github.com/korjavin/reviewbot/internal/github"
	"github.com/korjavin/reviewbot/internal/oauth"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /webhook", ghwebhook.NewWebhookHandler(cfg))
	mux.HandleFunc("GET /callback", oauth.NewCallbackHandler(cfg))
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("ReviewBot server starting on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
