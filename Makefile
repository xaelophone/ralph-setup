.PHONY: build install clean test

# Build the rwatch binary
build:
	go build -o rwatch ./cmd/rwatch

# Install to $GOPATH/bin
install:
	go install ./cmd/rwatch

# Clean build artifacts
clean:
	rm -f rwatch
	go clean

# Run tests
test:
	go test ./...

# Build for all platforms
build-all:
	GOOS=darwin GOARCH=amd64 go build -o rwatch-darwin-amd64 ./cmd/rwatch
	GOOS=darwin GOARCH=arm64 go build -o rwatch-darwin-arm64 ./cmd/rwatch
	GOOS=linux GOARCH=amd64 go build -o rwatch-linux-amd64 ./cmd/rwatch
	GOOS=linux GOARCH=arm64 go build -o rwatch-linux-arm64 ./cmd/rwatch

# Development: build and run
dev: build
	./rwatch --monitor-only
