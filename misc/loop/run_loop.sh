#!/usr/bin/env bash

set -e # exit on error

# Set up the kill signal callback
teardown() {
  echo "Stopping background processes..."
  kill 0
}

# Helper for checking the local exit code
check_exit_code() {
  local exit_code=$?
  if [ $exit_code -ne 0 ]; then
    echo "Error: Process failed with exit code $exit_code"
    teardown
  fi
}

echo "Running local development setup"

# Start the gnoland node (fresh chain), and in parallel
# - start the backup service for transactions
(
  echo "Starting Gno node..."
  make gnoland.start
  check_exit_code
) &
(
  echo "Starting backup..."
  make tx.backup
  check_exit_code
) &

# Trap all kill signals
trap 'teardown' INT

# Wait for all background processes to finish
wait
