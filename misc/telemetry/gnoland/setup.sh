#!/bin/bash

# Initialize the node data directories
gnoland start --skip-start

# Set the block time to 1s
gnoland config set --config-path /opt/gno/src/gnoland-data/config/config.toml consensus.timeout_commit 1s

# Set the listen address
gnoland config set --config-path /opt/gno/src/gnoland-data/config/config.toml rpc.laddr tcp://0.0.0.0:26657

# Enable the metrics
gnoland config set --config-path /opt/gno/src/gnoland-data/config/config.toml telemetry.enabled true

# Set the metrics exporter endpoint
gnoland config set --config-path /opt/gno/src/gnoland-data/config/config.toml telemetry.exporter_endpoint collector:4317

# Start the Gnoland node
gnoland start