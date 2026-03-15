# govDAO Scripts

Scripts for govDAO members to manage the gnoland1 chain. If you're a validator operator (valoper), you can ignore this directory.

All scripts default to `GNOKEY_NAME=moul`, `CHAIN_ID=gnoland1`, and `REMOTE=https://rpc.betanet.testnets.gno.land:443`. Override via env vars.

```bash
./add-validator-from-valopers.sh ADDR        # add a validator registered at r/gnops/valopers
./add-validator.sh ADDR PUBKEY [POWER]       # add a validator with explicit pub_key
./rm-validator.sh ADDR                       # remove a validator
./extend-govdao-t1.sh                       # add 6 T1 members to govDAO (one-time bootstrap)
```
