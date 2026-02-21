# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.22-bookworm AS builder

WORKDIR /build
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o claude-runner .

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM debian:bookworm-slim

# Install git + Node.js (for anythingllm-mcp-server MCP tool) + Claude Code CLI.
# Claude Code now ships as a native binary — no Node required for claude itself.
# Node is still needed for anythingllm-mcp-server.
RUN apt-get update && apt-get install -y --no-install-recommends \
        git \
        ca-certificates \
        curl \
    && curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && npm install -g anythingllm-mcp-server \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* \
    && curl -fsSL https://claude.ai/install.sh | bash

# The native Claude installer puts the binary at ~/.local/bin/claude.
# Add it to PATH so the Go runner and any shell scripts can invoke `claude`.
ENV PATH="/root/.local/bin:$PATH"

COPY --from=builder /build/claude-runner /usr/local/bin/claude-runner

EXPOSE 8080

CMD ["claude-runner"]
