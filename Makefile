.PHONY: build test clean run install deps fmt lint build-all integration-test help

# Build the iperf3-go server
build:
	go build -o iperf3-go main.go

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f iperf3-go iperf3-go.exe
	go clean

# Run the server with default settings
run: build
	./iperf3-go -v

# Run the server on a different port
run-port: build
	./iperf3-go -v -p 8080

# Install dependencies
deps:
	go mod tidy

# Format code
fmt:
	go fmt ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Build for different platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o iperf3-go-linux-amd64 main.go
	GOOS=windows GOARCH=amd64 go build -o iperf3-go-windows-amd64.exe main.go
	GOOS=darwin GOARCH=amd64 go build -o iperf3-go-darwin-amd64 main.go
	GOOS=darwin GOARCH=arm64 go build -o iperf3-go-darwin-arm64 main.go

# Test with actual iperf3 client (requires iperf3 to be installed)
integration-test: build
	@echo "Starting server in background..."
	@./iperf3-go -v & \
	SERVER_PID=$$!; \
	sleep 2; \
	echo "Running iperf3 client test..."; \
	iperf3 -c localhost -t 5 || echo "iperf3 client test failed or not available"; \
	kill $$SERVER_PID

# Quick test with netcat (if iperf3 not available)
quick-test: build
	@echo "Starting server in background..."
	@./iperf3-go -v & \
	SERVER_PID=$$!; \
	sleep 2; \
	echo "Testing connection with netcat..."; \
	echo "test" | nc localhost 5201 || echo "Connection test failed"; \
	kill $$SERVER_PID

# Install iperf3 client for testing (Ubuntu/Debian)
install-iperf3-ubuntu:
	sudo apt-get update && sudo apt-get install -y iperf3

# Install iperf3 client for testing (CentOS/RHEL)
install-iperf3-centos:
	sudo yum install -y iperf3

# Show help
help:
	@echo "Available targets:"
	@echo "  build              - Build the iperf3-go server"
	@echo "  test               - Run Go tests"
	@echo "  clean              - Clean build artifacts"
	@echo "  run                - Build and run server with verbose output"
	@echo "  run-port           - Run server on port 8080"
	@echo "  deps               - Install/update dependencies"
	@echo "  fmt                - Format Go code"
	@echo "  lint               - Run linter (requires golangci-lint)"
	@echo "  build-all          - Build for multiple platforms"
	@echo "  integration-test   - Test with real iperf3 client"
	@echo "  quick-test         - Quick connection test with netcat"
	@echo "  install-iperf3-*   - Install iperf3 client for testing"
	@echo "  help               - Show this help message"