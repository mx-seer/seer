.PHONY: build build-ce build-pro dev test clean build-frontend build-all-ce build-all-pro

# Build frontend
build-frontend:
	cd web && bun run build

# Build CE binary
build-ce: build-frontend
	mkdir -p bin
	go build -o bin/seer ./cmd/seer

# Build Pro binary
build-pro: build-frontend
	mkdir -p bin
	go build -tags pro -o bin/seer-pro ./cmd/seer

# Default build (CE)
build: build-ce

# Development (frontend + backend separate)
dev:
	cd web && bun run dev &
	go run ./cmd/seer

# Run Go server only (for development without frontend)
dev-server:
	go run ./cmd/seer

# Run tests
test:
	go test -cover ./...
	cd web && bun run test:e2e

# Run Go tests only
test-go:
	go test -cover ./...

# Clean build artifacts
clean:
	rm -rf bin/ dist/ web/build data

# Build all platforms (CE)
build-all-ce: build-frontend
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -o dist/seer-linux-amd64 ./cmd/seer
	GOOS=linux GOARCH=arm64 go build -o dist/seer-linux-arm64 ./cmd/seer
	GOOS=darwin GOARCH=amd64 go build -o dist/seer-darwin-amd64 ./cmd/seer
	GOOS=darwin GOARCH=arm64 go build -o dist/seer-darwin-arm64 ./cmd/seer
	GOOS=windows GOARCH=amd64 go build -o dist/seer-windows-amd64.exe ./cmd/seer

# Build all platforms (Pro)
build-all-pro: build-frontend
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -tags pro -o dist/seer-pro-linux-amd64 ./cmd/seer
	GOOS=linux GOARCH=arm64 go build -tags pro -o dist/seer-pro-linux-arm64 ./cmd/seer
	GOOS=darwin GOARCH=amd64 go build -tags pro -o dist/seer-pro-darwin-amd64 ./cmd/seer
	GOOS=darwin GOARCH=arm64 go build -tags pro -o dist/seer-pro-darwin-arm64 ./cmd/seer
	GOOS=windows GOARCH=amd64 go build -tags pro -o dist/seer-pro-windows-amd64.exe ./cmd/seer
