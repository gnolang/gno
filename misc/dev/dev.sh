#!/usr/bin/env bash

# Constants and Configurations
TX_ARCHIVE_BINARY="github.com/gnolang/tx-archive/cmd@latest"
DELAY=5

# The restore source directory
restore_dir=$1

# The backup output directory
backup_dir=$2

# Set up the kill signal callback
teardown() {
  echo "Stopping background processes..."
  kill 0
}

# Set up the required tx-archive tool
install_tools() {
  echo "Installing tx-archive binary"
  if go install $TX_ARCHIVE_BINARY; then
    echo "Installation successful"
  else
    echo "Failed to install the binary"
    exit 1
  fi
}

# Helper for checking the local exit code
check_exit_code() {
  local exit_code=$?
  if [ $exit_code -ne 0 ]; then
    echo "Error: Process failed with exit code $exit_code"
    teardown
  fi
}

# Install the required tools
install_tools

# Pull in the latest changes from VC
cd ../..
git checkout master
git pull

# Clean out the blockchain data
cd gno.land && make clean fclean build install

# Start the gnoland node (fresh chain), and in parallel
# - start the restore service for transactions
# - start the backup service for transactions
(
  echo "Starting Gno node..."
  gnoland start
  check_exit_code
) &
(
  # Sleep the restore until the node is fully loaded up
  sleep $DELAY
  echo "Starting restore..."
  cmd restore -legacy -watch -input-path "$restore_dir"
  check_exit_code
) &
(
  # Sleep the restore until the node is fully loaded up
  sleep $DELAY
  echo "Starting backup..."
  cmd backup -legacy -watch -overwrite -output-path "$backup_dir"
  check_exit_code
) &

# Trap all kill signals
trap 'teardown' INT

# Wait for all background processes to finish
wait
