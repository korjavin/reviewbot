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

# Create a non-root user.
RUN useradd -m -u 1000 -s /bin/bash claude

# Create an app directory for binaries so they aren't shadowed by the
# /home/claude Docker volume mount in docker-compose.yml.
RUN mkdir -p /app && chown claude:claude /app

USER claude

# Set npm to install globally into /app/.npm-global
ENV NPM_CONFIG_PREFIX=/app/.npm-global
ENV PATH="/app/.npm-global/bin:/app/.local/bin:$PATH"

# Install anythingllm-mcp-server into /app/.npm-global
# Also temporarily set HOME=/app so the Claude Code installer script
# places the `claude` binary in `/app/.local/bin/claude` rather than `/home/claude`.
RUN npm install -g anythingllm-mcp-server \
    && HOME=/app curl -fsSL https://claude.ai/install.sh | bash

COPY --from=builder /build/claude-runner /usr/local/bin/claude-runner

EXPOSE 8080

CMD ["claude-runner"]
