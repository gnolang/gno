# Gno networks

## Network configurations

| Network     | RPC Endpoint                     | Chain ID      |
|-------------|----------------------------------|---------------|
| Portal Loop | https://rpc.gno.land:443         | `portal-loop` |
| Test5       | https://rpc.test5.gno.land:443   | `test5`       |

### WebSocket endpoints
All networks follow the same pattern for websocket connections:

```shell
wss://<rpc-endpoint:port>/websocket
```

## Staging Environments (Portal Loops)

XXX: tell that portal loop is currently using a custom code but will switch to a gnodev powered alternative, usable by anyone to run a staging

Portal Loop is an always-up-to-date staging testnet that allows for using
the latest version of Gno, gno.land, and TM2. By utilizing the power of Docker
& the [tx-archive](https://github.com/gnolang/tx-archive) tool, the Portal Loop
can run the latest code from the master branch on the [Gno monorepo](https://github.com/gnolang/gno),
while preserving most/all the previous transaction data.

The Portal Loop allows for quick iteration on the latest version of Gno - without
having to make a hard/soft fork.

Below is a diagram demonstrating how the Portal Loop works:
```
                    +----------------------------------+
                    |       Portal Loop running        |  < ----+
                    +----------------------------------+        |
                                     |                          |
                                     |                          |
                                     v                          |
                    +----------------------------------+        |
                    |   Detect changes in 'master'     |        |
                    +----------------------------------+        |
                                     |                          |
                                     |                          |
                                     v                          |
                    +----------------------------------+        |
                    | Archive transaction data & state |        |
                    +----------------------------------+        |
                                     |                          |
                                     |                          |
                                     v                          |
                    +----------------------------------+        |
                    |    Load changes from 'master'    |        |
                    +----------------------------------+        |
                                     |                          |
                                     |                          |
                                     v                          |
                    +----------------------------------+        |
                    |      Replay transaction data     |  ------+
                    +----------------------------------+
```

Specifically, Portal Loop behaves like a normal network until a change is detected
in the `master` branch in the Gno monorepo. At this point, Portal Loop archives
on-chain data using the [tx-archive](https://github.com/gnolang/tx-archive)
tool, saving all transactions that happened on it thus far.

It then pulls the latest changes from the `master` branch, and inserts all
previously archived transactions into the genesis of the newly deployed chain.
After genesis has been replayed, the chain continues working as normal.

### Using the Portal Loop

The Portal Loop deployment can be found at [gno.land](https://gno.land), while
the exposed RPC endpoints can be found on `https://rpc.gno.land:443`.

XXX: list or link to the list of available RPC endpoints.

#### A warning note

While allowing for quick iteration on the most up-to-date software, the Portal Loop
has some drawbacks:
- If a breaking change happens on `master`, transactions that used the previous version of
Gno will fail to be replayed, meaning **data will be lost**.
- Since transactions are archived and replayed during genesis,
block height & timestamp cannot be relied upon.

#### Deploying to the Portal Loop

There are two ways to deploy code to the Portal Loop:

1. *automatic* - all packages in found in the `examples/gno.land/{p,r}/` directory in the [Gno monorepo](https://github.com/gnolang/gno) get added to the
   new genesis each cycle,
2. *permissionless* - this includes replayed transactions with `addpkg`, and
   new transactions you can issue with `gnokey maketx addpkg`.

Since the packages in `examples/gno.land/{p,r}` are deployed first,
permissionless deployments get superseded when packages with identical `pkgpath`
get merged into `examples/`.

The above mechanism is also how the `examples/` on the Portal Loop
get collaboratively iterated upon, which is its main mission.

## Gno Testnets

gno.land testnets are categorized by 4 main points:
- **Persistence of state**
  - Is the state and transaction history persisted?
- **Timeliness of code**
  - How up-to-date are Gno language features and demo packages & realms?
- **Intended purpose**
  - When should this testnet be used?
- **Versioning strategy**
  - How is this testnet versioned?

Below you can find a breakdown of each existing testnet by these categories.

### Portal Loop

Portal Loop is an always up-to-date rolling testnet. It is meant to be used as
a nightly build of the Gno tech stack. The home page of [gno.land](https://gno.land)
is the `gnoweb` render of the Portal Loop testnet.

- **Persistence of state:**
  - State is kept on a best-effort basis
  - Transactions that are affected by breaking changes will be discarded
- **Timeliness of code:**
  - Packages & realms which are available in the `examples/` folder on the
    [Gno monorepo](https://github.com/gnolang/gno) exist on the Portal Loop in
    matching state - they are refreshed with every new commit to the `master`
    branch.
- **Intended purpose**
  - Providing access the latest version of Gno for fast development & demoing
- **Versioning strategy**:
  - Portal Loop infrastructure is managed within the
    [`misc/loop`](https://github.com/gnolang/gno/tree/master/misc/loop) folder in the
    monorepo

### Test5

Test5 a permanent multi-node testnet. It bumped the validator set from 7 to 17
nodes, introduced GovDAO V2, and added lots of bug fixes and quality of life
improvements.

Test5 was launched in November 2024.

- **Persistence of state:**
  - State is fully persisted unless there are breaking changes in a new release,
    where persistence partly depends on implementing a migration strategy
- **Timeliness of code:**
  - Pre-deployed packages and realms are at monorepo commit [2e9f5ce](https://github.com/gnolang/gno/tree/2e9f5ce8ecc90ee81eb3ae41c06bab30ab926150)
- **Intended purpose**
  - Running a full node, testing validator coordination, deploying stable Gno
    dApps, creating tools that require persisted state & transaction history
- **Versioning strategy**:
  - Test5 is to be release-based, following releases of the Gno tech stack.

### TestX

These testnets are deprecated and currently serve as archives of previous progress.

## Test4 (archive)

Test4 is the first permanent multi-node testnet. Archived data for test4 can be
found [here](https://github.com/gnolang/tx-exports/tree/main/test4.gno.land).

Launch date: July 10th 2024
Release commit: [194903d](https://github.com/gnolang/gno/commit/194903db0350ace7d57910e6c34125d3aa9817da)

### Test3 (archive)

The third Gno testnet. Archived data for test3 can be found [here](https://github.com/gnolang/tx-exports/tree/main/test3.gno.land).

Launch date: November 4th 2022
Release commit: [1ca2d97](https://github.com/gnolang/gno/commit/1ca2d973817b174b5b06eb9da011e1fcd2cca575)

### Test2 (archive)

The second Gno testnet. Find archive data [here](https://github.com/gnolang/tx-exports/tree/main/test2.gno.land).

Launch date: July 10th 2022
Release commit: [652dc7a](https://github.com/gnolang/gno/commit/652dc7a3a62ee0438093d598d123a8c357bf2499)

### Test1 (archive)

The first Gno testnet. Find archive data [here](https://github.com/gnolang/tx-exports/tree/main/test1.gno.land).

Launch date: May 6th 2022
Release commit: [797c7a1](https://github.com/gnolang/gno/commit/797c7a132d65534df373c63b837cf94b7831ac6e)
