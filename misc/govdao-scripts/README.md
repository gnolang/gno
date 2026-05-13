# govDAO Scripts

Shared scripts for govDAO governance operations. These scripts require `GNOKEY_NAME`, `CHAIN_ID`, and `REMOTE` to be set.

**Don't call these scripts directly** — use the deployment wrapper instead:

```bash
# From the deployment directory:
./govdao                                     # list available commands
./govdao add-validator-from-valopers ADDR    # add a validator registered at r/gnops/valopers
./govdao add-validator ADDR PUBKEY [POWER]   # add a validator with explicit pub_key
./govdao rm-validator ADDR                   # remove a validator
./govdao extend-govdao-t1                    # add 6 T1 members to govDAO (one-time bootstrap)
./govdao unrestrict-account ADDR [ADDR...]   # allow address(es) to transfer ugnot
./govdao restrict-account ADDR [ADDR...]     # re-restrict account(s) from transferring ugnot
./govdao set-cla URL                         # set/update CLA document via govDAO proposal
./govdao set-valoper-minfee AMOUNT           # update valoper registration minimum fee
```

Each deployment wrapper (e.g., `misc/deployments/gnoland1/govdao`) sets the correct chain ID, RPC endpoint, and default key name.
