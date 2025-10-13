.PHONY: build test clean install examples

# Build the kcl2xrd binary
build:
	go build -o bin/kcl2xrd ./cmd/kcl2xrd

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install the binary to GOPATH/bin
install:
	go install ./cmd/kcl2xrd

# Generate example XRDs
examples: build
	./bin/kcl2xrd --input examples/kcl/postgresql.k --group database.example.org --output examples/xrd/postgresql.yaml
	./bin/kcl2xrd --input examples/kcl/validated.k --group example.org --output examples/xrd/validated.yaml
	./bin/kcl2xrd --input examples/kcl/nested-schema.k --group storage.example.org --output examples/xrd/nested-schema.yaml
	./bin/kcl2xrd --input examples/kcl/dynatrace-with-metadata.k --output examples/xrd/dynatrace-with-metadata.yaml
	./bin/kcl2xrd --input examples/kcl/preserve-unknown-fields.k --group config.example.org --output examples/xrd/preserve-unknown-fields.yaml

# Format code
fmt:
	go fmt ./...

# Run linter
lint:
	golangci-lint run

# Run all checks (format, lint, test)
check: fmt test

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build the kcl2xrd binary"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Clean build artifacts"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  examples       - Generate example XRDs"
	@echo "  fmt            - Format code"
	@echo "  lint           - Run linter"
	@echo "  check          - Run format and tests"
	@echo "  help           - Show this help message"
