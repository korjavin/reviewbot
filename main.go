package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/korjavin/reviewbot/internal/config"
	ghwebhook "github.com/korjavin/reviewbot/internal/github"
	"github.com/korjavin/reviewbot/internal/middleware"
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
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"app":"reviewbot","status":"ok"}`))
	})

	handler := middleware.Logging(mux)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("ReviewBot server starting on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
