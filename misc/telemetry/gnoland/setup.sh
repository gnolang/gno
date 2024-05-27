#!/bin/bash

# Initialize the node config
/gnoland config init --config-path /gnoroot/config/config.toml

# Set the block time to 1s
/gnoland config set --config-path /gnoroot/config/config.toml consensus.timeout_commit 1s

# Set the listen address
/gnoland config set --config-path /gnoroot/config/config.toml rpc.laddr tcp://0.0.0.0:26657

# Enable the metrics
/gnoland config set --config-path /gnoroot/config/config.toml telemetry.enabled true

# Set the metrics exporter endpoint
/gnoland config set --config-path /gnoroot/config/config.toml telemetry.exporter_endpoint collector:4317

# Start the Gnoland node (lazy will init the genesis.json and secrets)
/gnoland start --lazy