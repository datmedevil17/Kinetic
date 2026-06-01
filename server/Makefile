## ── Dev ─────────────────────────────────────────────────────────────────────
run:
	go run ./cmd/server

# Hot-reload dev server (install: go install github.com/air-verse/air@latest)
dev:
	air

## ── Build ────────────────────────────────────────────────────────────────────
build:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/kinetic ./cmd/server

## ── Quality ──────────────────────────────────────────────────────────────────
lint:
	golangci-lint run ./...

test:
	go test ./... -v

## ── Docker ───────────────────────────────────────────────────────────────────
docker-build:
	docker build -t kinetic-server .

docker-run:
	docker run --env-file .env -p 8080:8080 kinetic-server

.PHONY: run dev build lint test docker-build docker-run
