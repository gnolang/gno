#!/usr/bin/env bash

# This script finds and tidies all go.mod
# files recursively from the repository root and
# optionally verifies the last echo based on an
# environment variable VERIFY_MOD_SUMS

set -e # exit on error

# CD into the repo root
cd ..

# Check for the verify argument
verify=${VERIFY_MOD_SUMS:-false}

# Find all go.mod files
gomods=$(find . -type f -name go.mod)

if $verify; then
  # Calculate sums for all go.mod files
  sums=$(shasum $gomods)
fi

# Tidy each go.mod file
for modfile in $gomods; do
  dir=$(dirname "$modfile")

  # Run go mod tidy in the directory
  echo "Running \`go -C $dir mod tidy -v\`"
  go -C "$dir" mod tidy -v || exit 1
done

# Optionally verify the sums
if $verify; then
  echo "Verifying sums..."
  echo "$sums" | shasum -c
else
  echo "Skipping sum verification"
fi
