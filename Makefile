.PHONY: build schema clean test

# Default target
all: schema build

run:
	go run ./src $(ARGS)


# Build the angel binary
build:
	@echo "Building angel..."
	go build -o ./angel ./src

# Generate JSON schema from Config struct
schema:
	@echo "Generating JSON schema..."
	mkdir -p build
	go run -tags tools ./tools/schema.go
	@echo "Schema generated: angel-config-schema.json"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -f angel
	rm -f angel-config-schema.json

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod tidy


# Release build (optimized)
release: schema
	@echo "Building angel (release mode)..."
	go build -ldflags="-s -w" -o angel src/*.go
