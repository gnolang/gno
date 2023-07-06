# Proof of Contribution

The consensus algorithm used to secure and distribute the gno.land chain.

## Main Concepts

- Many details are yet to be defined, but here are the main ideas:
- The validator set is determined by a DAO acting as an authority, similar to a
  Proof-of-Authority (PoA) system. However, in this case, the authority is a DAO
  called `worxDAO`.
- The `worxDAO` utilizes the evaluation DAO to distribute `^worx` tokens, elect
  new members, and assign membership levels.
- The entire DAO and governance functionalities are managed through smart
  contracts.
- The chain monitors contract changes, such as the `valset`, to configure
  `Tendermint2` accordingly.
- There is no staking mechanism involved in the consensus process.
- Similar to Proof-of-Stake (PoS), chain fees are intended to be distributed.
  However, instead of going to stakers, they will be distributed among
  contributors and validators.
- To ensure efficient distribution, the chain will accumulate all the chain fees
  in a contract-managed bucket.

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
    |  |                             |     |   r/system/valset   |  |              |                                    |
    |  |                             |  +->|                     |--+------+       |  +-------------+                   |
    |  |           worxDAO           |  |  |    validator set    |  |      |       |  |             |                   |
    |  |                             |--+  +---------------------+  |      +-------+->|   Gno SDK   |----------+        |
    |  |   the "Contributors DAO"    |  |  |  r/system/chaincfg  |  |      |       |  |             |          |        |
    |  |                             |  +->|                     |--+------+       |  +-------------+          |        |
    |  |                             |     | chain configuration |  |              |         |                 |        |
    |  +-----------------------------+     +---------------------+  |              |         |                 |        |
    |                 |                    +---------------------+  |              |         v                 v        |
    |                 v                    |  r/system/rewards   |  |              |  +-------------+   +-------------+ |
    |      +--------------------+          |                     |  |              |  |             |   |             | |
    |      |   Evaluation DAO   |          |distribute rewards to|  |              |  |     TM2     |-->|    GnoVM    | |
    |      |                    |          |  contributors and   |  |              |  |             |   |             | |
    |      |Qualification system|          |     validators      |  |              |  +-------------+   +-------------+ |
    |      |to distribute ^worx |          |       +------+      |  |              |         |                 |        |
    |      +--------------------+          |       |Bucket|<- - -|- + -chain fees -|- - - - -                  |        |
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
`r/system` realms, including `valset` and `chaincfg`.

### `r/system/valset`

A realm (smart contract) that enables the `worxDAO` to update the validator set.
Similar to a PoA system, the authority is decentralized in a DAO.

Additionally, this contract is queried by `gno.land` to configure `TM2` when
changes are made to the validator set.

### `r/system/chaincfg`

A governance-backed smart contract that allows for chain configuration through
governance.

It helps prevent unnecessary upgrade campaigns for minor updates.
