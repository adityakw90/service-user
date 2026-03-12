#!/bin/bash
# Create config.yaml.example from config.yaml with sensitive values redacted

set -e

INPUT_FILE="${1:-config.yaml}"
OUTPUT_FILE="${2:-config.yaml.example}"

# Check if input file exists
if [ ! -f "$INPUT_FILE" ]; then
    echo "Error: $INPUT_FILE not found"
    exit 1
fi

# Create output file with sensitive values replaced
sed -e 's/password:.*$/password: "<YOUR_DB_PASSWORD>"/g' \
    -e 's/secret:.*$/secret: "<YOUR_SECRET>"/g' \
    -e 's/key:.*$/key: "<YOUR_API_KEY>"/g' \
    -e 's/secret_key:.*$/secret_key: "<YOUR_SECRET_KEY>"/g' \
    -e 's/api_key:.*$/api_key: "<YOUR_API_KEY>"/g' \
    -e 's/jwt_secret:.*$/jwt_secret: "<YOUR_JWT_SECRET>"/g' \
    -e 's/client_secret:.*$/client_secret: "<YOUR_OAUTH_CLIENT_SECRET>"/g' \
    -e 's/access_token:.*$/access_token: "<YOUR_ACCESS_TOKEN>"/g' \
    -e 's/refresh_token:.*$/refresh_token: "<YOUR_REFRESH_TOKEN>"/g' \
    "$INPUT_FILE" > "$OUTPUT_FILE"

echo "Created $OUTPUT_FILE with sensitive values redacted"