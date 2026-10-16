# govDAO Scripts

Shared scripts for govDAO governance operations. These scripts require `GNOKEY_NAME`, `CHAIN_ID`, and `REMOTE` to be set.

`set-halt.sh` also reads the version the node at `REMOTE` runs from its `/status` (with `curl`, or `wget` when `curl` is missing) and refuses a `MIN_VERSION` that version already meets: such a floor gates nothing after the halt and refuses every restart of the running binary before it (see [Restart gates](../../gno.land/cmd/gnoland/UPGRADES.md#restart-gates)). It reads one node, directly as `gnokey` does (proxy variables are ignored), and prints the version it compared: behind a load balancer, check the other nodes when the network runs mixed versions. A `REMOTE` it cannot read stops the script; cancelling (height `0`) reads nothing.

**Don't call these scripts directly** — use the deployment wrapper instead:

```bash
# From the deployment directory:
./govdao-exec.sh                                     # list available commands
./govdao-exec.sh add-validator-from-valopers ADDR    # v2: add a validator registered at r/gnops/valopers
./govdao-exec.sh add-validator ADDR PUBKEY [POWER]   # v2: add a validator with explicit pub_key
./govdao-exec.sh rm-validator ADDR                   # v2: remove a validator
./govdao-exec.sh add-validator-v0 OPADDR [POWER]     # v0: add validator by operator-address (test-13+)
./govdao-exec.sh rm-validator-v0 OPADDR              # v0: remove validator by operator-address (test-13+)
./govdao-exec.sh register-valoper MONIKER DESC TYPE OPADDR PUBKEY   # operator self-register profile (NOT govDAO-signed)
./govdao-exec.sh register-user USERNAME ADDR         # govDAO-grant a custom username for ADDR
./govdao-exec.sh extend-govdao-t1                    # add 6 T1 members to govDAO (one-time bootstrap)
./govdao-exec.sh unrestrict-account ADDR [ADDR...]   # allow address(es) to transfer ugnot
./govdao-exec.sh restrict-account ADDR [ADDR...]     # re-restrict account(s) from transferring ugnot
./govdao-exec.sh set-cla URL                         # set/update CLA document via govDAO proposal
./govdao-exec.sh set-valoper-minfee AMOUNT           # update valoper registration minimum fee
./govdao-exec.sh set-halt HEIGHT MIN_VERSION          # schedule a coordinated halt (HEIGHT 0 cancels)
```

The `-v0` validator commands route through `r/sys/validators/v0` (operator-keyed,
post-VALOPLAN2). Use them on chains running v0 (test-13 onward). On chains still
running v2 (gnoland1 pre-hardfork), use the unsuffixed `add-validator` / `rm-validator`.

Each deployment wrapper (e.g., `misc/deployments/mainnet.gno.land/govdao-exec.sh`; a couple of older deployments name it `govdao`) sets the correct chain ID, RPC endpoint, and default key name.
