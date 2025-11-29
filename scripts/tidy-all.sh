#!/usr/bin/env bash

set -u  # Exit on undefined variables

# Check if go command is available
if ! command -v go >/dev/null 2>&1; then
    echo "Error: go command not found. Please install Go."
    exit 1
fi

# Change to the project root directory
cd "$(dirname "$0")/.." || {
    echo "Error: Failed to change to project root directory."
    exit 1
}

# Check if internal directory exists
if [ ! -d "internal" ]; then
    echo "Error: internal directory not found."
    exit 1
fi

# Loop through each subdirectory in internal/
for dir in internal/*/; do
    if [ -d "$dir" ] && [ -f "${dir}go.mod" ]; then
        echo "Tidying ${dir%/}"
        if (cd "$dir" && go mod tidy); then
            echo "Successfully tidied ${dir%/}"
        else
            echo "Error: Failed to tidy ${dir%/}"
        fi
    fi
done
