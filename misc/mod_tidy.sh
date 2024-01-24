#!/usr/bin/env bash

# This script finds and tidies all go.mod
# files recursively from the repository root

set -e # exit on error

# CD into the repo root
cd ..

# Find all go.mod files
gomods=$(find . -type f -name go.mod)

# Calculate sums for all go.mod files
sums=$(shasum $gomods)

# Tidy each go.mod file
for modfile in $gomods; do
  dir=$(dirname "$modfile")

  # Run go mod tidy in the directory
  (cd "$dir" && go mod tidy -v) || exit 1
done

# Verify the sums
echo "$sums" | shasum -c