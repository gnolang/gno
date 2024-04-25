---
id: proof-of-contribution
---

# Proof of Contribution

The gno.land chain utilizes a reputation-based consensus mechanism instead of proof-of-stake.
This mechanism emphasizes values and expertise to carry out important tasks like choosing validators and allocating rewards.

Meta issue: [#918](https://github.com/gnolang/gno/issues/918).

Presentation: https://github.com/gnolang/workshops/tree/main/presentations/2023-06-06--buidl-asia--manfred.

## Main Concepts

- Validator set determined by `worxDAO`, a DAO serving as the authority.
- Governance and distribution managed through smart contracts.
- Chain monitors contract changes (e.g., `valset`) to configure `Tendermint2`.
- No staking involved in consensus.
- Chain fees distributed to contributors and validators, not stakers.
- Chain fees accumulated in a contract-managed bucket for efficient distribution.
- Validators likely have equal power (1).
- Validators do not vote like in PoS, but may participate in dedicated governance topics, maybe.

## High-level schema

            ____                   ____         ____                    __       _ __          __  _
           / __ \_________  ____  / __/  ____  / __/  _________  ____  / /______(_) /_  __  __/ /_(_)___  ____  _____
          / /_/ / ___/ __ \/ __ \/ /_   / __ \/ /_   / ___/ __ \/ __ \/ __/ ___/ / __ \/ / / / __/ / __ \/ __ \/ ___/
         / ____/ /  / /_/ / /_/ / __/  / /_/ / __/  / /__/ /_/ / / / / /_/ /  / / /_/ / /_/ / /_/ / /_/ / / / (__  )
        /_/   /_/   \____/\____/_/     \____/_/     \___/\____/_/ /_/\__/_/  /_/_.___/\__,_/\__/_/\____/_/ /_/____/

    +---------------------------------------------------------------+              +------------------------------------+
    |                   gno.land/{p,r} contracts                    |              |              gno.land              |
    |                                                               |              |                                    |
    |  +-----------------------------+     +---------------------+  |              |                                    |
    |  |                             |     |   r/sys/validators  |  |              |                                    |
    |  |                             |  +->|                     |--+------+       |  +-------------+                   |
    |  |           worxDAO           |  |  |    validator set    |  |      |       |  |             |                   |
    |  |                             |--+  +---------------------+  |      +-------+->|   Gno SDK   |----------+        |
    |  |   the "Contributors DAO"    |  |  |    r/sys/config     |  |      |       |  |             |          |        |
    |  |                             |  +->|                     |--+------+       |  +-------------+          |        |
    |  |                             |     | chain configuration |  |              |         |                 |        |
    |  +-----------------------------+     +---------------------+  |              |         |                 |        |
    |                 |                    +---------------------+  |              |         v                 v        |
    |                 v                    |    r/sys/rewards    |  |              |  +-------------+   +-------------+ |
    |      +----------------------+        |                     |  |              |  |             |   |             | |
    |      |    Evaluation DAO    |        | distribute rewards  |  |              |  |     TM2     |-->|    GnoVM    | |
    |      |                      |        | to contributors and |  |              |  |             |   |             | |
    |      | Qualification system |        |     validators      |  |              |  +-------------+   +-------------+ |
    |      | to distribute ^worx  |        |       +------+      |  |              |         |                 |        |
    |      +----------------------+        |       |Bucket|<- - -|- + -chain fees -|- - - - -                  |        |
    |                                      +-------+------+------+  |              |                           |        |
    +---------------------------------------------------------------+              +---------------------------+--------+
                                    ^                                                                          |
                                    |                                                                          |
                                    +---------------user TXs can publish and call contracts--------------------+

## Components

### `gno.land`

The main blockchain powered by the `TM2` engine. It offers permissionless smart
contracts with the `GnoVM` and can self-configure from contracts using the
`GnoSDK`.

### `worxDAO`

The governance entity consisting of contributors, responsible for governing the
`r/sys` realms, including `validators` and `config`.

Meta issue: [#872](https://github.com/gnolang/gno/issues/872).

### `r/sys/validators`

A realm (smart contract) that enables the `worxDAO` to update the validator set.
Similar to a PoA system, the authority is decentralized in a DAO.

Additionally, this contract is queried by `gno.land` to configure `TM2` when
changes are made to the validator set.

### `r/sys/config`

A governance-backed smart contract that allows for chain configuration through
governance.

It helps prevent unnecessary upgrade campaigns for minor updates.

### Evaluation DAO

The system employed by the `worxDAO` to incentivize contributions with `^worx` points.

            +---------------1. propose a contribution-------------+
            |                                                     v
    +--------------+                                     +----------------+
    |              |--------3. improve, negotiate------->|                |
    | contributor  |                                     | Evaluation DAO |
    |              |<-------4. distribute ^worx----------|                |
    +--------------+                                     +----------------+
            ^                                                     |
            +---------------2. review, challenge------------------+
