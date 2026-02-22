# ── Build stage ───────────────────────────────────────────────────────────────
FROM golang:1.22-bookworm AS builder

WORKDIR /build
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o gemini-runner .

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM debian:bookworm-slim

# Install git + Node.js (for @google/gemini-cli).
RUN apt-get update && apt-get install -y --no-install-recommends \
        git \
        ca-certificates \
        curl \
    && curl -fsSL https://deb.nodesource.com/setup_22.x | bash - \
    && apt-get install -y nodejs \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Create a non-root user.
RUN useradd -m -u 1000 -s /bin/bash gemini

# Create an app directory for binaries so they aren't shadowed by home volume mounts.
RUN mkdir -p /app && chown gemini:gemini /app

USER gemini

# Set npm to install globally into /app/.npm-global
ENV NPM_CONFIG_PREFIX=/app/.npm-global
ENV PATH="/app/.npm-global/bin:/app/bin:$PATH"

# Install the Google Gemini CLI.
RUN npm install -g @google/gemini-cli

COPY --from=builder /build/gemini-runner /usr/local/bin/gemini-runner

EXPOSE 8080

CMD ["gemini-runner"]
