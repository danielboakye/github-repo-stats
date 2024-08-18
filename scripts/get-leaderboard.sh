#!/bin/bash

set -e  # Exit immediately if a command exits with a non-zero status
set -o pipefail  # Exit if any command in a pipeline fails

# Default values
DEFAULT_LIMIT=5
DEFAULT_PORT=9000

# Assign arguments to variables with defaults if not provided
LIMIT=${1:-$DEFAULT_LIMIT}
PORT=${2:-$DEFAULT_PORT}

# Example usage of LIMIT and PORT
echo "Using limit: $LIMIT"
echo "Using port: $PORT"


URL="http://localhost:$PORT/v1/leaderboard"

if [ ! -z "$LIMIT" ]; then
  URL="$URL?limit=$LIMIT"
fi

echo "Running curl -s -X GET $URL \"Accept: application/json\""

# Make the curl request
curl -s -X GET "$URL" -H "Accept: application/json"

echo
