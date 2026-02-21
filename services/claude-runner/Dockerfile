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
ENV PATH="/app/.npm-global/bin:/app/bin:$PATH"

# Install anythingllm-mcp-server into /app/.npm-global
# The native Claude install script ignores $HOME and installs to the user's
# true home dir (/home/claude/.local/bin/claude). Since that directory is
# shadowed by our docker-compose volume, we explicitly move the binary
# out of the home dir and into /app/bin.
RUN npm install -g anythingllm-mcp-server \
    && curl -fsSL https://claude.ai/install.sh | bash \
    && mkdir -p /app/bin \
    && mv /home/claude/.local/bin/claude /app/bin/claude

COPY --from=builder /build/claude-runner /usr/local/bin/claude-runner

EXPOSE 8080

CMD ["claude-runner"]
