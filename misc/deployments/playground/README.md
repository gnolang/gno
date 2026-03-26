# Playground Deployment Configs

This directory contains deployment configurations for play.gno.land.

## Files

- `gnoland1.0.env` — environment for the gnoland1.0 mainnet deployment

## Usage

```bash
# Build for gnoland1.0
cd contribs/playground
make build-release TAG=chain/gnoland1.0

# Or use the deploy script
../../misc/deployments/playground/deploy.sh chain/gnoland1.0
```
