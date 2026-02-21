# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.22-bookworm AS builder

WORKDIR /build
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o claude-runner .

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM debian:bookworm-slim

# Install git + Node.js (for anythingllm-mcp-server MCP tool).
RUN apt-get update && apt-get install -y --no-install-recommends \
        git \
        ca-certificates \
        curl \
    && curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Create a non-root user. Claude Code refuses --dangerously-skip-permissions
# when running as root, so all claude operations must run as a normal user.
RUN useradd -m -u 1000 -s /bin/bash claude

# Switch to the claude user for all subsequent steps.
# Set npm prefix to user-local dir so `npm install -g` doesn't need root.
USER claude
ENV NPM_CONFIG_PREFIX=/home/claude/.npm-global
ENV PATH="/home/claude/.npm-global/bin:/home/claude/.local/bin:$PATH"

# Install anythingllm-mcp-server (goes to ~/.npm-global/bin) and Claude Code CLI
# (native installer puts binary at ~/.local/bin/claude).
RUN npm install -g anythingllm-mcp-server \
    && curl -fsSL https://claude.ai/install.sh | bash

COPY --from=builder /build/claude-runner /usr/local/bin/claude-runner

EXPOSE 8080

CMD ["claude-runner"]
