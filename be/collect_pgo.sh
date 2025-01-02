#!/bin/bash

# Directory to store the profiles
OUTPUT_DIR="profiles"
mkdir -p "$OUTPUT_DIR"

# URL to download the profile
URL="http://103.82.133.178:8080/debug/pprof/profile?seconds=30"

# Maximum number of files to keep
MAX_FILES=100

while true; do
  # Generate a timestamped file name
  TIMESTAMP=$(date +%Y%m%d_%H%M%S)
  FILE="$OUTPUT_DIR/profile_$TIMESTAMP.prof"

  # Download the profile
  wget -q -O "$FILE" "$URL"

  # Check if the file is non-empty
  if [ -s "$FILE" ]; then
    echo "Saved: $FILE"
  else
    echo "Empty file, skipping: $FILE"
    rm -f "$FILE"
  fi

  # Ensure only the most recent $MAX_FILES are kept
  FILE_COUNT=$(ls "$OUTPUT_DIR" | wc -l)
  if [ "$FILE_COUNT" -gt "$MAX_FILES" ]; then
    FILES_TO_DELETE=$((FILE_COUNT - MAX_FILES))
    ls -t "$OUTPUT_DIR" | tail -n "$FILES_TO_DELETE" | xargs -I {} rm -f "$OUTPUT_DIR/{}"
    echo "Deleted $FILES_TO_DELETE old files."
  fi
done
