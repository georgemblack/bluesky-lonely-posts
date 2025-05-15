#!/bin/bash

# Check for exactly one argument
if [ "$#" -ne 1 ]; then
  echo "Usage: $0 https://bsky.app/profile/USERNAME"
  exit 1
fi

# Extract the username from the URL
URL="$1"
USERNAME=$(basename "$URL")

# Use CURL to get the DID
RESPONSE=$(curl -s "https://bsky.social/xrpc/com.atproto.identity.resolveHandle?handle=${USERNAME}")

# Extract the "did" field from the JSON response
DID=$(echo "$RESPONSE" | jq -r '.did')

# Check if DID was successfully extracted
if [ "$DID" == "null" ] || [ -z "$DID" ]; then
  echo "Failed to resolve DID for username: $USERNAME"
  exit 1
fi

# Append the DID to the text file
echo "$DID" >> pkg/app/assets/dids.txt

# Deduplicate the file (in-place)
sort -u pkg/app/assets/dids.txt -o pkg/app/assets/dids.txt

echo "Appended and deduplicated DID: $DID"