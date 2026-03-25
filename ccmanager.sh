#!/bin/bash
# ccmanager - Launch script for Claude Code session manager

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="$SCRIPT_DIR/ccmanager"

if [ ! -f "$BINARY" ]; then
    echo "Binary not found. Building..."
    cd "$SCRIPT_DIR" && go build -o ccmanager ./cmd/ccmanager
fi

exec "$BINARY" "$@"
