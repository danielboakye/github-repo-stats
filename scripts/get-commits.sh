#!/bin/bash

set -e  # Exit immediately if a command exits with a non-zero status
set -o pipefail  # Exit if any command in a pipeline fails

# Default values
DEFAULT_LIMIT=5
DEFAULT_PORT=9000

if [ -z "$1" ]; then
  echo "Usage: $0 <repoName> [limit] [port]"
  echo "<repoName> is required"
  echo "[limit] is optional and will default to $DEFAULT_LIMIT"
  echo "[port] is optional and will default to $DEFAULT_PORT"
  exit 1
fi

REPO_NAME=$1

# Assign arguments to variables with defaults if not provided
LIMIT=${2:-$DEFAULT_LIMIT}
PORT=${3:-$DEFAULT_PORT}


# Example usage of LIMIT and PORT
echo "Using limit: $LIMIT"
echo "Using port: $PORT"

echo "Running curl -s -X GET \"http://localhost:$PORT/v1/commits?repoName=$REPO_NAME&limit=$LIMIT\" -H \"Accept: application/json\""

curl -s -X GET "http://localhost:$PORT/v1/commits?repoName=$REPO_NAME&limit=$LIMIT" -H "Accept: application/json"

echo 
