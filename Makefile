.PHONY: help build run test fmt vet clean install-deps dev

help:
	@echo "IrisFlair API - Available commands:"
	@echo "  make install-deps  - Download Go dependencies"
	@echo "  make build         - Build the application binary"
	@echo "  make run           - Run the server"
	@echo "  make dev           - Run in development mode (with auto-reload)"
	@echo "  make test          - Run tests"
	@echo "  make fmt           - Format code"
	@echo "  make vet           - Lint code"
	@echo "  make clean         - Remove build artifacts"

install-deps:
	go mod download
	go mod tidy

build:
	go build -o app main.go
	@echo "✓ Built successfully: ./app"

run: build
	./app

dev:
	@command -v air > /dev/null || (echo "Installing air for hot reload..." && go install github.com/cosmtrek/air@latest)
	air

test:
	go test -v ./...

fmt:
	go fmt ./...
	@echo "✓ Code formatted"

vet:
	go vet ./...
	@echo "✓ Code linted"

clean:
	rm -f app
	go clean
	@echo "✓ Cleaned build artifacts"
