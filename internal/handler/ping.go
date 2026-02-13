package handler

import (
	"log"

	gh "github.com/google/go-github/v68/github"
)

func HandlePing(event *gh.PingEvent) {
	log.Printf("Ping received, webhook is working. Zen: %s", event.GetZen())
}
