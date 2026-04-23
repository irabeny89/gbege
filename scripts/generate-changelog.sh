#!/bin/bash

# generate-changelog.sh
# This script generates the project changelog using cocogitto.

if ! command -v cog &> /dev/null; then
    echo "Error: cocogitto (cog) is not installed."
    echo "Please install it from: https://github.com/cocogitto/cocogitto"
    exit 1
fi

echo "Generating CHANGELOG.md..."
cog changelog > CHANGELOG.md
echo "Changelog successfully generated at CHANGELOG.md."
