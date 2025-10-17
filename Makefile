.PHONY: build schema clean test

# Default target
all: schema build

run:
	go run ./src $(ARGS)


# Build the angel binary
build: schema
	@echo "Building angel..."
	go build -o build/angel ./src

# Generate JSON schema from Config struct
schema:
	@echo "Generating JSON schema..."
	go run -tags tools ./tools/schema.go
	@echo "Schema generated: angel-config-schema.json"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f angel
	rm -f angel-config-schema.json

# Run tests
test:
	@echo "Running tests..."
	go test ./...

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy

# Development build (with race detection)
dev: schema
	@echo "Building angel (dev mode)..."
	go build -race -o angel src/*.go

# Release build (optimized)
release: schema
	@echo "Building angel (release mode)..."
	go build -ldflags="-s -w" -o angel src/*.go
