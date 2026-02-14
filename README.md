# ReviewBot - GitHub App for Automated Interactions

ReviewBot is a GitHub App designed to automate interactions with pull requests and issues. It demonstrates how to handle webhooks, authenticate as a GitHub App, and use the GitHub API to perform actions like commenting, reacting, and creating pull requests.

## Features

- **Ping Event**: Responds to GitHub App ping events to verify connectivity.
- **Issue Comments**:
  - Detects comments containing `@reviewbot` (ignores bot comments).
  - Reacts with an "eyes" (👀) emoji.
  - Replies with a "Pong!" message quoting a snippet of the original comment.
- **Pull Request Automation**:
  - Automatically triggered when a new Pull Request is opened (ignores bot PRs).
  - Clones the repository to a temporary directory.
  - Creates a new branch (`reviewbot/review-pr-{prNumber}`).
  - Adds a timestamped file (`YYYY-MM-DD.txt`) to the branch.
  - Commits and pushes the new branch.
  - Opens a **new** Pull Request from this branch targeting the default branch.
  - Comments on the **original** Pull Request with a link to the newly created review PR.

## Architecture

The application is written in Go and uses the following key components:
- **Web Server**: Standard library `net/http` server.
- **GitHub API Client**: `google/go-github` for API interactions.
- **Authentication**: `bradleyfalzon/ghinstallation` handles JWT creation and installation token management.
- **Git Operations**: Uses `os/exec` to run `git` commands for cloning, committing, and pushing.

## Prerequisites

- **Go**: Version 1.25 or higher.
- **Docker**: Optional, for containerized deployment.
- **Git**: Installed and available in the system PATH (required for PR automation).

## Configuration

ReviewBot is configured via environment variables. Create a `.env` file (see `.env.example`) or set these variables in your environment.

| Variable | Required | Description |
|---|---|---|
| `GITHUB_APP_ID` | Yes | The ID of your GitHub App. |
| `GITHUB_PRIVATE_KEY_PATH` | Yes* | Path to the private key PEM file. |
| `GITHUB_PRIVATE_KEY` | Yes* | Raw content of the private key PEM (alternative to path). |
| `GITHUB_WEBHOOK_SECRET` | Yes | The secret used to secure webhooks. |
| `GITHUB_CLIENT_ID` | No | OAuth Client ID (if using OAuth flow). |
| `GITHUB_CLIENT_SECRET` | No | OAuth Client Secret (if using OAuth flow). |
| `PORT` | No | Server port (default: 8080). |
| `BASE_URL` | No | Public URL for OAuth redirects. |

\* *One of `GITHUB_PRIVATE_KEY_PATH` or `GITHUB_PRIVATE_KEY` must be provided.*

## Installation & Running

### Local Development

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/korjavin/reviewbot.git
    cd reviewbot
    ```

2.  **Install dependencies:**
    ```bash
    go mod download
    ```

3.  **Run the application:**
    ```bash
    # Set environment variables first or use a .env loader
    export GITHUB_APP_ID=...
    export GITHUB_PRIVATE_KEY_PATH=...
    export GITHUB_WEBHOOK_SECRET=...

    go run main.go
    ```

### Docker

1.  **Build and run with Docker Compose:**
    ```bash
    docker-compose up --build
    ```
    Ensure your `.env` file is present in the root directory.

## GitHub App Setup

To use ReviewBot, you must register a GitHub App with the following permissions and events:

### Repository Permissions
- **Contents**: Read & Write (to commit and push changes).
- **Issues**: Read & Write (to comment on issues/PRs).
- **Pull Requests**: Read & Write (to create and comment on PRs).
- **Metadata**: Read-only (mandatory).

### Subscribe to Events
- **Pull request**: To trigger on opened PRs.
- **Issue comment**: To trigger on `@reviewbot` mentions.

### Webhook URL
Set the Webhook URL to your public endpoint (e.g., `https://your-domain.com/webhook`). For local development, use a proxy like Smee or Ngrok.

## Development

To test webhooks locally:

1.  **Using Smee.io**:
    - Create a channel on [smee.io](https://smee.io).
    - Install the client: `npm install --global smee-client`.
    - Run the client: `smee -u https://smee.io/your-channel -t http://localhost:8080/webhook`.

2.  **Using Ngrok**:
    - Run `ngrok http 8080`.
    - Use the generated HTTPS URL as your GitHub App's Webhook URL.

## Testing

Run the test suite using standard Go tooling:

```bash
go test ./...
```

## Project Structure

- `main.go`: Application entry point and server setup.
- `internal/config/`: Configuration loading and validation.
- `internal/github/`: GitHub client factory and webhook handler logic.
- `internal/handler/`: Event handlers for Ping, Issue Comment, and Pull Request events.
- `internal/git/`: Helper functions for executing Git commands.
- `internal/middleware/`: HTTP middleware (e.g., logging).
- `internal/oauth/`: OAuth callback handler.
