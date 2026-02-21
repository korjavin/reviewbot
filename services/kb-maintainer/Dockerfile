FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /kb-maintainer .

FROM alpine:3.21
WORKDIR /app
RUN apk add --no-cache ca-certificates

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    mkdir -p /intels /state && \
    chown -R appuser:appuser /app /intels /state

USER appuser

# /intels  — bind-mount the intel markdown files (read-only)
# /state   — volume for persistent sync state
VOLUME ["/intels", "/state"]

COPY --from=builder /kb-maintainer /app/kb-maintainer

ENTRYPOINT ["/app/kb-maintainer"]
