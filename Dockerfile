FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /reviewbot main.go

FROM alpine:3.21
WORKDIR /app
RUN apk add --no-cache ca-certificates git

COPY --from=builder /reviewbot /app/reviewbot

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8080
ENTRYPOINT ["/app/reviewbot"]
