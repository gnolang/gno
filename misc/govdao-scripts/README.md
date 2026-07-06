# govDAO Scripts

Shared scripts for govDAO governance operations. These scripts require `GNOKEY_NAME`, `CHAIN_ID`, and `REMOTE` to be set.

**Don't call these scripts directly** — use the deployment wrapper instead:

```bash
# From the deployment directory:
./govdao                                     # list available commands
./govdao add-validator-from-valopers ADDR    # v2: add a validator registered at r/gnops/valopers
./govdao add-validator ADDR PUBKEY [POWER]   # v2: add a validator with explicit pub_key
./govdao rm-validator ADDR                   # v2: remove a validator
./govdao add-validator-v3 OPADDR [POWER]     # v3: add validator by operator-address (test-13+)
./govdao rm-validator-v3 OPADDR              # v3: remove validator by operator-address (test-13+)
./govdao register-valoper MONIKER DESC TYPE OPADDR PUBKEY   # operator self-register profile (NOT govDAO-signed)
./govdao register-user USERNAME ADDR         # govDAO-grant a custom username for ADDR
./govdao extend-govdao-t1                    # add 6 T1 members to govDAO (one-time bootstrap)
./govdao unrestrict-account ADDR [ADDR...]   # allow address(es) to transfer ugnot
./govdao restrict-account ADDR [ADDR...]     # re-restrict account(s) from transferring ugnot
./govdao set-cla URL                         # set/update CLA document via govDAO proposal
./govdao set-valoper-minfee AMOUNT           # update valoper registration minimum fee
```

The `-v3` validator commands route through `r/sys/validators/v3` (operator-keyed,
post-VALOPLAN2). Use them on chains running v3 (test-13 onward). On chains still
running v2 (gnoland1 pre-hardfork), use the unsuffixed `add-validator` / `rm-validator`.

Each deployment wrapper (e.g., `misc/deployments/gnoland1/govdao`) sets the correct chain ID, RPC endpoint, and default key name.
