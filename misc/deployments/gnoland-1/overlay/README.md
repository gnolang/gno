# Upgrade Overlay Scripts

Scripts in this directory are executed during Phase 2 of `generate-genesis.sh`,
in alphabetical order. Each script receives the working genesis path as `$1`.

## Naming convention

Use numbered prefixes to control execution order:

```
01-update-params.sh     # Parameter changes
02-deploy-contracts.sh  # New contract deployments
03-set-cla.sh           # CLA configuration
```

## Example

```bash
#!/usr/bin/env bash
# 01-update-params.sh — Update valoper min_fee to 0
GENESIS="$1"
gnogenesis params set vm:gno.land/r/gnops/valopers min_fee 0 --genesis-path "$GENESIS"
```
