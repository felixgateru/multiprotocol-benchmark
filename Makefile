.PHONY: help build run clean install cross-compile deps setup

# Default target
help:
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  run            - Build and run the benchmark"
	@echo "  clean          - Remove built binaries"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  cross-compile  - Build binaries for multiple platforms"
	@echo "  deps           - Download and tidy dependencies"
	@echo "  setup          - Setup environment configuration"

build:
	@echo "Building multiprotocol-benchmark..."
	go build -o multiprotocol-benchmark .


run: build
	@echo "Running benchmark..."
	./multiprotocol-benchmark

clean:
	@echo "Cleaning..."
	rm -f multiprotocol-benchmark
	rm -rf dist/
	rm -f test_results_*.json

install:
	@echo "Installing to GOPATH/bin..."
	go install .

cross-compile:
	@echo "Cross-compiling for multiple platforms..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 go build -o dist/multiprotocol-benchmark-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build -o dist/multiprotocol-benchmark-linux-arm64 .
	GOOS=darwin GOARCH=amd64 go build -o dist/multiprotocol-benchmark-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o dist/multiprotocol-benchmark-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o dist/multiprotocol-benchmark-windows-amd64.exe .
	@echo "Binaries built in dist/"

deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

setup:
	@if [ ! -f .env ]; then \
		echo "Creating .env file from .env.example..."; \
		cp .env.example .env; \
		echo "Please edit .env with your configuration"; \
	else \
		echo ".env file already exists"; \
	fi
