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

# Install anythingllm-mcp-server and Claude Code CLI as the claude user
# so the binaries land in /home/claude/.local/bin (where the installer puts them).
USER claude
RUN npm install -g anythingllm-mcp-server \
    && curl -fsSL https://claude.ai/install.sh | bash

# Add the claude user's local bin to PATH.
ENV PATH="/home/claude/.local/bin:$PATH"

# The Go binary needs to run as claude too (it spawns claude and writes ~/.claude.json).
COPY --from=builder /build/claude-runner /usr/local/bin/claude-runner

EXPOSE 8080

CMD ["claude-runner"]
