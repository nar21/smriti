# Project variables
BINARY_NAME=smriti
PKG=./...
GOFILES=$(shell find . -name '*.go' -type f -not -path "./vendor/*")

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	go build -o bin/$(BINARY_NAME) .

# Run the app
.PHONY: run
run: build
	./bin/$(BINARY_NAME)


# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning..."
	rm -rf bin
