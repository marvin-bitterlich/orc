.PHONY: install build test clean

# Build and install ORC binary globally
install:
	go build -o $(GOPATH)/bin/orc ./cmd/orc
	@echo "✓ ORC installed to $(GOPATH)/bin/orc"

# Build ORC binary locally
build:
	go build -o orc ./cmd/orc
	@echo "✓ ORC binary built: ./orc"

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -f orc
	go clean
