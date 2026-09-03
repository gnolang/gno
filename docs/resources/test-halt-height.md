# Test GovDAO Halt Height Feature

Single-validator manual test plan for the governance-based chain halt mechanism.

For where this fits in a chain upgrade, see [Chain upgrades](chain-upgrades.md).
For how the mechanism works internally (params, arming, the two startup checks),
see [`gno.land/adr/pr5368_govdao_halt_height.md`](../../gno.land/adr/pr5368_govdao_halt_height.md).

## Cleanup / Starting Fresh

To reset and start the test from scratch:

```bash
# Stop the node (Ctrl-C if running), then delete the data directory.
rm -rf ./testnode
```

This removes all chain state, config, and secrets. The next `gnoland start --lazy`
will regenerate everything from scratch.

## Prerequisites

### Build two gnoland binaries with distinct versions

The startup checks compare `tm2/pkg/version.Version` against the governance
`halt_min_version` param. Comparison only understands the
`chain/gnoland<major>.<minor>` format; anything else falls back to exact string
equality. The `gno.land/Makefile` build targets *do* inject a version, but it is
derived from `git describe` (e.g. `master.12345+abc1234`), which never parses as
a chain version — so you must pass the version explicitly.

```bash
# "Old" binary — simulates the currently running chain software.
go build -ldflags "-X github.com/gnolang/gno/tm2/pkg/version.Version=chain/gnoland1.0" \
    -o ./build/gnoland-v1.0 ./gno.land/cmd/gnoland

# "New" binary — simulates the upgrade target.
go build -ldflags "-X github.com/gnolang/gno/tm2/pkg/version.Version=chain/gnoland1.1" \
    -o ./build/gnoland-v1.1 ./gno.land/cmd/gnoland

# Also build gnokey (for submitting transactions).
go build -o ./build/gnokey ./gno.land/cmd/gnokey
```

### Add the test1 key to gnokey

The default genesis balances file funds the `test1` account. Import it using
its well-known seed phrase.

```bash
printf '%s\n\n\n' \
    'source bonus chronic canvas draft south burst lottery vacant surface solve popular case indicate oppose farm nothing bullet exhibit title speed wink action roast' \
    | ./build/gnokey add test1 --recover --insecure-password-stdin
```

Verify the address matches the funded genesis account:

```bash
./build/gnokey list
# Should show: test1 -> g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5
```

### Start the node and bootstrap GovDAO membership

```bash
# Start with --lazy to auto-generate genesis, secrets, and config.
# --skip-genesis-sig-verification is required because the default genesis txs
# are signed by a key that differs from the lazy-generated validator key.
./build/gnoland-v1.0 start --lazy --data-dir ./testnode \
    --genesis ./testnode/genesis.json \
    --skip-genesis-sig-verification
```

The default GovDAO loader creates tiers and sets the DAO impl, but adds **no
members** and leaves `AllowedDAOs` empty. When `AllowedDAOs` is empty, any
caller can interact with the DAO. We exploit this to bootstrap `test1` as a T1
member via MsgRun.

Write a bootstrap script (`/tmp/bootstrap_govdao.gno`):

```gno
package main

import (
    "gno.land/r/gov/dao/v3/memberstore"
)

func main(cur realm) {
    // Add test1 as a T1 member (supermajority power).
    // The loader already set the DAO impl; AllowedDAOs is empty so any
    // caller is permitted. We just need to register the member.
    err := memberstore.Get(0, cur).SetMember(memberstore.T1,
        address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
        memberstore.NewMember(3),
    )
    if err != nil {
        panic(err)
    }
}
```

Submit it (in a second terminal while the node is running):

```bash
./build/gnokey maketx run \
    -gas-fee 1000000ugnot -gas-wanted 100000000 \
    -broadcast -chainid dev \
    test1 /tmp/bootstrap_govdao.gno
```

**Expected**: `OK!` — test1 is now a T1 GovDAO member.

Note the current block height from the node logs and choose a **halt height**
comfortably in the future (e.g. current height + 40). `WillSetParam` rejects a
halt height that is not strictly greater than the height at which the proposal
executes, so leave room for the vote and execution txs. The examples below use
`50`; substitute your own value consistently.

## Propose and execute a halt via GovDAO

**Goal**: Verify that a governance proposal can schedule a chain halt.

### Create the halt proposal

Write a Gno script (`/tmp/propose_halt.gno`). Adjust the halt height as needed:

```gno
package main

import (
    "gno.land/r/gov/dao"
    "gno.land/r/sys/params"
)

func main(cur realm) {
    preq := params.NewSetHaltRequest(cross(cur), 50, "chain/gnoland1.1")
    dao.MustCreateProposal(cross(cur), preq)
}
```

Submit it:

```bash
./build/gnokey maketx run \
    -gas-fee 1000000ugnot -gas-wanted 100000000 \
    -broadcast -chainid dev \
    test1 /tmp/propose_halt.gno
```

**Expected**: `OK!` — proposal ID `0` created.

### Verify the proposal

```bash
./build/gnokey query vm/qrender --data 'gno.land/r/gov/dao:0'
```

**Expected**: Output shows a proposal titled "Set node halt height" with
description mentioning block 50 and version `chain/gnoland1.1`.

### Vote YES

```bash
./build/gnokey maketx call \
    -pkgpath gno.land/r/gov/dao -func MustVoteOnProposalSimple \
    -args 0 -args YES \
    -gas-fee 1000000ugnot -gas-wanted 10000000 \
    -broadcast -chainid dev test1
```

**Expected**: `OK!`

### Execute the proposal

```bash
./build/gnokey maketx call \
    -pkgpath gno.land/r/gov/dao -func ExecuteProposal \
    -args 0 \
    -gas-fee 1000000ugnot -gas-wanted 10000000 \
    -broadcast -chainid dev test1
```

**Expected**: `OK!`

### Verify params are set

```bash
./build/gnokey query params/node:p:halt_height
# Expected: data: "50"

./build/gnokey query params/node:p:halt_min_version
# Expected: data: "chain/gnoland1.1"
```

### The new binary is refused before the halt

**Goal**: Verify the pre-halt startup check (`checkNodeStartupParams`, check 2):
a binary that already meets `halt_min_version` must not run before the chain
has reached `halt_height`.

Stop the node (Ctrl-C) while the chain is still below block 50 and try the new
binary:

```bash
./build/gnoland-v1.1 start --data-dir ./testnode --genesis ./testnode/genesis.json
```

**Expected**: startup fails with

```
binary version "chain/gnoland1.1" is an upgrade intended for halt height 50,
but the chain is at height <N>; please use the previous binary until the halt,
or set skip_upgrade_height = 50 in config.toml if you have already migrated
```

Restart the old binary and let it run to the halt:

```bash
./build/gnoland-v1.0 start --data-dir ./testnode --genesis ./testnode/genesis.json
```

### Observe the halt

Block 50 — the halt height — **is committed**. The EndBlocker of block 50 arms
the halt, and `BaseApp.BeginBlock` panics when the *next* block (51) begins.
The last committed block is therefore 50, not 49.

Watch the logs for the EndBlocker arming the halt at block 50:

```
GovDAO halt height reached, will halt after this block  height=50  halt_height=50
```

then, as block 51 begins, the panic surfacing through the consensus routine's
recover:

```
CONSENSUS FAILURE!!!  err="halt height 50 reached, node shutting down"  stack=...
```

**Result**: the consensus routine stops and the node produces no further
blocks. Note the **process does not exit** — `gnoland start` blocks on its
signal context, not on consensus health, so the RPC server stays up and the
process idles. Stop it with Ctrl-C.

### The old binary is refused after the halt

**Goal**: Verify the post-halt startup check (`checkNodeStartupParams`, check 1):
once the chain has reached `halt_height`, a binary below `halt_min_version`
must not resume it.

```bash
./build/gnoland-v1.0 start --data-dir ./testnode --genesis ./testnode/genesis.json
```

**Expected**: startup fails with

```
binary version "chain/gnoland1.0" does not meet the minimum version
"chain/gnoland1.1" required by governance; please upgrade to a compatible
binary before restarting
```

### New binary resumes after halt

**Goal**: Verify that the upgraded binary (`chain/gnoland1.1`) passes
the startup check and resumes the chain.

```bash
./build/gnoland-v1.1 start --data-dir ./testnode --genesis ./testnode/genesis.json
```

**Expected**: The node starts and the chain continues from block 51 onward.
The EndBlocker arms the halt only on `req.Height == halt_height`, so at height
51 and above it never re-fires and the node does not halt again.

The halt params are **not** cleared on resume — nothing in the node clears
them, by design:

```bash
./build/gnokey query params/node:p:halt_height
# Expected: still data: "50"

./build/gnokey query params/node:p:halt_min_version
# Expected: still data: "chain/gnoland1.1"
```

`halt_min_version` therefore stays in force as a permanent minimum-version
floor: every subsequent restart re-runs check 1 and keeps binaries below
`chain/gnoland1.1` off the chain. To lift the requirement, pass a new GovDAO
proposal with `NewSetHaltRequest(cross(cur), 0, "")`.
