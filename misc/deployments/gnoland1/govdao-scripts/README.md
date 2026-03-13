# govDAO Scripts

Scripts for govDAO members to manage the gnoland1 chain. If you're a validator operator (valoper), you can ignore this directory.

## Scripts

- **add-validator.sh** - Add a validator to the active set via govDAO proposal
- **rm-validator.sh** - Remove a validator from the active set via govDAO proposal

## Usage

```bash
# Add a validator
./add-validator.sh <address> <pub_key> [voting_power]

# Remove a validator
./rm-validator.sh <address>
```

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `GNOKEY_NAME` | `moul` | gnokey key name |
| `CHAIN_ID` | `gnoland1` | Chain ID |
| `REMOTE` | `https://rpc.betanet.testnets.gno.land:443` | RPC endpoint |
| `GAS_WANTED` | `10000000` | Gas limit |
| `GAS_FEE` | `1000000ugnot` | Gas fee |
