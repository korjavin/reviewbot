package github

import (
	"log"
	"net/http"

	gh "github.com/google/go-github/v68/github"
	"github.com/korjavin/reviewbot/internal/config"
	"github.com/korjavin/reviewbot/internal/handler"
)

func NewWebhookHandler(cfg *config.Config) http.HandlerFunc {
	newClient := func(installationID int64) (*gh.Client, error) {
		return NewInstallationClient(cfg, installationID)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		payload, err := gh.ValidatePayload(r, []byte(cfg.WebhookSecret))
		if err != nil {
			log.Printf("ERROR: invalid webhook signature: %v", err)
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}

		eventType := gh.WebHookType(r)
		event, err := gh.ParseWebHook(eventType, payload)
		if err != nil {
			log.Printf("ERROR: parsing webhook: %v", err)
			http.Error(w, "parse error", http.StatusBadRequest)
			return
		}

		log.Printf("Received webhook: %s", eventType)

		switch e := event.(type) {
		case *gh.PingEvent:
			handler.HandlePing(e)
		case *gh.PullRequestEvent:
			go handler.HandlePullRequest(newClient, e)
		case *gh.IssueCommentEvent:
			go handler.HandleIssueComment(newClient, e)
		default:
			log.Printf("Unhandled event type: %s", eventType)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}
}
