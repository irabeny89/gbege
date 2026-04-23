.PHONY: all test coverage changelog setup-hooks

# Default target
all: test

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
