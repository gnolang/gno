#!/bin/bash


# Initialize the node config
gnoland config init --config-path /gnoroot/gnoland-data/config/config.toml

# Set the block time to 1s
gnoland config set --config-path /gnoroot/gnoland-data/config/config.toml consensus.timeout_commit 1s

# Set the listen address
gnoland config set --config-path /gnoroot/gnoland-data/config/config.toml rpc.laddr tcp://0.0.0.0:26657

# Enable the metrics
gnoland config set --config-path /gnoroot/gnoland-data/config/config.toml telemetry.enabled true

# Set the metrics exporter endpoint
# If you want to use otel-collector setup this
# gnoland config set --config-path /gnoroot/gnoland-data/config/config.toml telemetry.exporter_endpoint collector:4317

# To not use otel-collector, your can use prometheus flag `--enable-feature=otlp-write-receiver` and push metrics directly to it
gnoland config set --config-path /gnoroot/gnoland-data/config/config.toml telemetry.exporter_endpoint http://prometheus:9090/api/v1/otlp/v1/metrics


# Start the Gnoland node (lazy will init the genesis.json and secrets)
gnoland start --lazy --log-level=error
