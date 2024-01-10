#!/usr/bin/env bash

# This script is meant to orchestrate
# a parallel execution of a gno.land node
# and a backup tool that preserves
# transactions that happen on-chain while
# the node is running. Additionally, the
# script also closes down any hanging process
# if either the node / backup tool fail

set -e # exit on error

# Set up the kill signal callback
teardown() {
  echo "Stopping background processes..."
  kill 0
}

echo "Running local development setup"

# Start the gnoland node (fresh chain), and in parallel
# - start the backup service for transactions
(
  echo "Starting Gno node..."
  make start.gnoland
  teardown
) &
(
  echo "Starting backup..."
  make tx.backup
  teardown
) &

# Trap all kill signals
trap 'teardown' INT

# Wait for all background processes to finish
wait
