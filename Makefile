.PHONY: build run clean test install deps fmt lint

# Build the binary
build:
	@echo "Building flux..."
	go build -o flux .

# Run the application
run: deps build
	@echo "Running flux..."
	./flux

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f flux

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Install binary globally
# Installs to $(GOPATH)/bin (typically ~/go/bin)
# Make sure $(GOPATH)/bin is in your PATH
install: build
	@echo "Installing flux to $(GOPATH)/bin..."
	@echo "Ensure $(GOPATH)/bin is in your PATH"
	go install
	@echo "Installation complete! Run 'flux' from terminal."

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Lint code
lint:
	@echo "Linting code..."
	golangci-lint run ./...

.DEFAULT_GOAL := run
