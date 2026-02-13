# ReviewBot — GitHub App Bot

GitHub App bot that reacts to PR and comment events with ping-pong responses (PoC stage).

## Architecture

GitHub webhooks → Go HTTP server → GitHub API responses

The bot authenticates as a GitHub App installation using JWT + private key (via `ghinstallation`).

## Build & Run

```bash
# Build
go build -o reviewbot main.go

# Run (requires env vars, see .env.example)
./reviewbot

# Docker
docker compose up --build
```

## Project Structure

```
main.go                     — entry point, HTTP server
internal/config/config.go   — configuration from env vars
internal/github/client.go   — GitHub client factory (ghinstallation)
internal/github/webhook.go  — webhook validation, parsing, routing
internal/handler/ping.go    — ping event handler
internal/handler/pullrequest.go — PR opened → comment
internal/handler/comment.go — issue comment with @reviewbot → reply + reaction
internal/oauth/oauth.go     — OAuth callback for app installation
```

## Endpoints

- `POST /webhook` — GitHub webhook receiver
- `GET /callback` — OAuth callback (app installation flow)
- `GET /health` — health check

## Environment Variables

| Variable | Required | Description |
|---|---|---|
| `GITHUB_APP_ID` | yes | GitHub App ID |
| `GITHUB_PRIVATE_KEY_PATH` | yes* | Path to PEM file |
| `GITHUB_PRIVATE_KEY` | yes* | Raw PEM contents (alternative) |
| `GITHUB_WEBHOOK_SECRET` | yes | Webhook HMAC secret |
| `GITHUB_CLIENT_ID` | no | OAuth Client ID |
| `GITHUB_CLIENT_SECRET` | no | OAuth Client Secret |
| `PORT` | no | Server port (default: 8080) |
| `BASE_URL` | no | Public URL for OAuth redirects |

*One of `GITHUB_PRIVATE_KEY_PATH` or `GITHUB_PRIVATE_KEY` is required.

## Local Development

Use [smee.io](https://smee.io) or [ngrok](https://ngrok.com) to proxy webhooks to localhost:

```bash
# smee
npx smee-client --url https://smee.io/YOUR_CHANNEL --target http://localhost:8080/webhook

# ngrok
ngrok http 8080
```

## Testing

```bash
# Health check
curl http://localhost:8080/health

# Verify: create a PR on a repo with the app installed → bot comments
# Verify: comment with @reviewbot → bot replies + adds 👀 reaction
```
