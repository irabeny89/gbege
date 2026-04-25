# Variables
BINARY_NAME=gbege

# Declare targets that aren't files
.PHONY: all test coverage changelog setup-hooks env run build run-build clean

# Default target runs when you type just `make`
all: test-race build

# Run all tests
test:
	go test -v ./...

# Run tests with the race detector
test-race:
	go test -race ./...

# Generate a coverage report
coverage:
	go test -v -cover -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Setup git hooks
setup-hooks:
	chmod +x scripts/setup-hooks.sh
	./scripts/setup-hooks.sh

# Generate the changelog
changelog:
	chmod +x scripts/generate-changelog.sh
	./scripts/generate-changelog.sh

env:
	export $(grep -v '^#' .env | xargs)
	@echo ".env loaded successfully"

run: env
	go run main.go

build: env
	go build -o $(BINARY_NAME) main.go

run-build:
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)