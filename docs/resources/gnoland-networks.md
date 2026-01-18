# Gno networks

## Network configurations

| Network | RPC Endpoint                             | Chain ID  |
|---------|------------------------------------------|-----------|
| Staging | https://rpc.gno.land:443                 | `staging` |
| Test10  | https://rpc.test10.testnets.gno.land:443 | `test10`  |

### WebSocket endpoints

All networks follow the same pattern for websocket connections:

```shell
wss://<rpc-endpoint:port>/websocket
```

## Staging Environments

Staging is an always-up-to-date staging testnet that allows for using
the latest version of Gno, gno.land, and TM2. By utilizing the power of Docker
& the [tx-archive](https://github.com/gnolang/gno/tree/master/contribs/tx-archive) tool, the Staging
can run the latest code from the master branch on the [Gno monorepo](https://github.com/gnolang/gno),
while preserving most/all the previous transaction data.

The Staging chain allows for quick iteration on the latest version of Gno - without
having to make a hard/soft fork.

Below is a diagram demonstrating how the Staging chain works:
```
                    +----------------------------------+
                    |       Staging     running        |  < ----+
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

Specifically, Staging behaves like a normal network until a change is detected
in the `master` branch in the Gno monorepo. At this point, the Staging chain archives
on-chain data using the [tx-archive](https://github.com/gnolang/gno/tree/master/contribs/tx-archive)
tool, saving all transactions that happened on it thus far.

It then pulls the latest changes from the `master` branch, and inserts all
previously archived transactions into the genesis of the newly deployed chain.
After genesis has been replayed, the chain continues working as normal.

### Using the Staging network

The Staging network deployment can be found at [gno.land](https://gno.land), while
the exposed RPC endpoints can be found on `https://rpc.gno.land:443`.

#### A warning note

While allowing for quick iteration on the most up-to-date software, the Staging chain
has some drawbacks:
- If a breaking change happens on `master`, transactions that used the previous version of
  Gno will fail to be replayed, meaning **data will be lost**.
- Since transactions are archived and replayed during genesis,
  block height & timestamp cannot be relied upon.

#### Deploying to Staging

There are two ways to deploy code to Staging:

1. *automatic* - all packages in found in the `examples/gno.land/{p,r}/` directory in the [Gno monorepo](https://github.com/gnolang/gno) get added to the
   new genesis each cycle,
2. *permissionless* - this includes replayed transactions with `addpkg`, and
   new transactions you can issue with `gnokey maketx addpkg`.

Since the packages in `examples/gno.land/{p,r}` are deployed first,
permissionless deployments get superseded when packages with identical `pkgpath`
get merged into `examples/`.

The above mechanism is also how the `examples/` on Staging get collaboratively
iterated upon, which is its main mission.

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

### Staging chain

The Staging chain is an always up-to-date rolling testnet. It is meant to be used as
a nightly build of the Gno tech stack. The home page of [gno.land](https://gno.land)
is the `gnoweb` render of the Staging testnet.

- **Persistence of state:**
  - State is kept on a best-effort basis
  - Transactions that are affected by breaking changes will be discarded
- **Timeliness of code:**
  - Packages & realms which are available in the `examples/` folder on the
    [Gno monorepo](https://github.com/gnolang/gno) exist on Staging in
    matching state - they are refreshed with every new commit to the `master`
    branch.
- **Intended purpose**
  - Providing access the latest version of Gno for fast development & demoing
- **Versioning strategy**:
  - Staging infrastructure is managed within the
    [`misc/loop`](https://github.com/gnolang/gno/tree/master/misc/loop) folder in the
    monorepo

### Test10

The latest Gno.land testnet, released on the 18th of December, 2025.

- **Persistence of state:**
  - State is fully persisted unless there are breaking changes in a new release,
    where persistence partly depends on implementing a migration strategy
- **Timeliness of code:**
  - Pre-deployed packages and realms are at release tag [chain/test10.0](https://github.com/gnolang/gno/releases/tag/chain%2Ftest10.0)
- **Intended purpose**
  - Running a full node, testing validator coordination, deploying stable Gno
    dApps, creating tools that require persisted state & transaction history

This testnet introduces major changes to the codebase, such as the `std` package 
split, private realms, the storage fee collector, and more. 

### TestX

These testnets are deprecated and currently serve as archives of previous progress.

### Test9 (archive)

Test9 is the testnet released on the 14th of October, 2025.

### Test8 (archive)

Test8 is the testnet released on the 5th of September, 2025.

### Test7 (archive)

Test7 is the testnet released on the 25th of July, 2025.

### Test6 (archive)

Test6 enables token locking, implements the interrealm specification, GovDAO V3 and more.

Launch date: 23rd of June 2025

### Test5 (archive)

Test5 a permanent multi-node testnet. It bumped the validator set from 7 to 17
nodes, introduced GovDAO V2, and added lots of bug fixes and quality of life
improvements. Archived data for test5 can be
found [here](https://github.com/gnolang/tx-exports/tree/main/test5.gno.land).

Test5 was launched in November 2024.

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
