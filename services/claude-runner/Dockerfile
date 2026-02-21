# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.22-bookworm AS builder

WORKDIR /build
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o claude-runner .

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM debian:bookworm-slim

# Install git + Node.js.
# Node is needed for anythingllm-mcp-server (Claude's MCP tool for the KB).
# Everything is installed at image build time — no runtime npm calls.
RUN apt-get update && apt-get install -y --no-install-recommends \
        git \
        ca-certificates \
        curl \
    && curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && npm install -g @anthropic-ai/claude-code anythingllm-mcp-server \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /build/claude-runner /usr/local/bin/claude-runner

EXPOSE 8080

CMD ["claude-runner"]
