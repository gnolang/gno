# Test GovDAO Halt Height Feature

Single-validator manual test plan for the governance-based chain halt mechanism.

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
`halt_min_version` param. The `gnoland` Makefile target does **not** inject
the version via ldflags by default, so you must pass it explicitly.

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

The default GovDAO loader creates tiers but adds **no members** and leaves
`AllowedDAOs` empty. When `AllowedDAOs` is empty, any caller can interact with
the DAO. We exploit this to bootstrap `test1` as a T1 member via MsgRun.

Write a bootstrap script (`/tmp/bootstrap_govdao.gno`):

```gno
package main

import (
    "gno.land/r/gov/dao/v3/memberstore"
)

func main() {
    // Add test1 as a T1 member (supermajority power).
    // The loader already set the DAO impl; AllowedDAOs is empty so any
    // caller is permitted. We just need to register the member.
    memberstore.Get().SetMember(memberstore.T1,
        address("g1jg8mtutu9khhfwc4nxmuhcpftf0pajdhfvsqf5"),
        &memberstore.Member{InvitationPoints: 3},
    )
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
comfortably in the future (e.g. current height + 40).

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

func main() {
    preq := params.NewSetHaltRequest(50, "chain/gnoland1.1")
    dao.MustCreateProposal(cross, preq)
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

### Observe the halt

The node should panic at BeginBlock of block 50 (the halt height). The block
at halt_height is **never committed** — the last committed block is 49.

Watch the logs for:

```
halt height 50 reached, node shutting down
```

**Result**: The node process exits. Block 49 is the last committed block.

### New binary resumes after halt

**Goal**: Verify that the upgraded binary (`chain/gnoland1.1`) passes
the startup check and resumes the chain.

```bash
./build/gnoland-v1.1 start --data-dir ./testnode --genesis ./testnode/genesis.json
```

**Expected**: The node starts, replays block 50 (the halt height), and the
EndBlocker detects the binary meets the min version. It clears the halt params
from state. The chain continues normally from block 51 onward.

Watch the logs for:

```
binary meets halt min version, clearing halt params  height=50  halt_height=50  ...
```

Verify the halt params have been cleared:

```bash
./build/gnokey query params/node:p:halt_height
# Expected: data: "0"
```
