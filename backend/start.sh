#!/bin/bash

# Load environment variables if .env exists
if [ -f .env ]; then
    set -a  # automatically export all variables
    source .env
    set +a  # stop automatically exporting
fi

# Start the Go application
go run cmd/server/main.go