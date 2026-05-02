# Variables
BINARY_NAME=gbege
# default migration file name separator
MIG_SEP:=_
# default migration directory
MIG_DIR:=migrations
# default migration file name
MIG_NAME=latest$(MIG_SEP)Schema
# generate timestamp for migration file name
TS:=$(shell date +%s)

# Load .env if it exists and override variables above
-include .env
# export all variables
export

# Declare targets that aren't files
.PHONY: all mig-dir mig-file test coverage changelog setup-hooks run build run-build clean

# Default target runs when you type just `make`
all: test-race build

# Setup git hooks
setup-hooks:
	chmod +x scripts/setup-hooks.sh
	./scripts/setup-hooks.sh

# Generate the changelog
changelog:
	chmod +x scripts/generate-changelog.sh
	./scripts/generate-changelog.sh

# Generate migration directory
mig-dir: 
	@mkdir -p $(MIG_DIR)

# Create new migration file with name and directory
# pass args e.g. `make mig-file MIG_SEP='-' MIG_DIR='mig' MIG_NAME='name'` to override variables
# simply run `make mig-file` to use MIG_DIR and MIG_SEP from .env
mig-file: mig-dir
	@touch $(MIG_DIR)/$(TS)$(MIG_SEP)$(MIG_NAME).sql
	@echo "Migration file created: $(MIG_DIR)/$(TS)$(MIG_SEP)$(MIG_NAME).sql"

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

# Run the server
run:
	go run ./cmd/app

# Build the binary
build:
	go build -ldflags="-w -s" -o $(BINARY_NAME) ./cmd/app

# Run the binary
run-build:
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)