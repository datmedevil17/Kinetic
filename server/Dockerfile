# ── Stage 1: Build ────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS builder

# Install git for private module downloads (not needed now but good practice)
RUN apk add --no-cache git ca-certificates

WORKDIR /app

# Download dependencies first (Docker layer cache: only re-downloads when go.mod changes)
COPY go.mod go.sum ./
RUN go mod download

# Copy source and build a stripped, statically-linked binary
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /kinetic \
    ./cmd/server

# ── Stage 2: Run ──────────────────────────────────────────────────────────────
# Use scratch (zero-byte base image) for minimum attack surface and image size
FROM scratch

# TLS certificates — needed for HTTPS calls to Supabase
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# The compiled binary
COPY --from=builder /kinetic /kinetic

EXPOSE 8080

ENTRYPOINT ["/kinetic"]
