# Proof of Contribution

The consensus algorithm used to secure and distribute the gno.land chain.

## High-level schema

              ____                   ____         ____                    __       _ __          __  _
             / __ \_________  ____  / __/  ____  / __/  _________  ____  / /______(_) /_  __  __/ /_(_)___  ____  _____
            / /_/ / ___/ __ \/ __ \/ /_   / __ \/ /_   / ___/ __ \/ __ \/ __/ ___/ / __ \/ / / / __/ / __ \/ __ \/ ___/
           / ____/ /  / /_/ / /_/ / __/  / /_/ / __/  / /__/ /_/ / / / / /_/ /  / / /_/ / /_/ / /_/ / /_/ / / / (__  )
          /_/   /_/   \____/\____/_/     \____/_/     \___/\____/_/ /_/\__/_/  /_/_.___/\__,_/\__/_/\____/_/ /_/____/

    +---------------------------------------------------------------+               +---------------------------------------+
    |                       gno.land/{p,r} contracts                |               |               gno.land                |
    |                                                               |               |                                       |
    |  +-----------------------------+     +---------------------+  |               |                                       |
    |  |                             |     |   r/system/valset   |  |               |                                       |
    |  |                             |  +->|                     |--+-------+       |  +-------------+     +-------------+  |
    |  |           worxDAO           |  |  |    validator set    |  |       |       |  |             |     |             |  |
    |  |                             |--+  +---------------------+  |       +-------+->|   Gno SDK   |---->|     TM2     |  |
    |  |   the "Contributors DAO"    |  |  |  r/system/chaincfg  |  |       |       |  |             |     |             |  |
    |  |                             |  +->|                     |--+-------+       |  +-------------+     +-------------+  |
    |  |                             |     | chain configuration |  |               |         |                   |         |
    |  +-----------------------------+     +---------------------+  |               |         |                   |         |
    |                 |                                             |               |         |                   |         |
    |                 v                                             |               |         |                   |         |
    |      +--------------------+                                   |               |         |                   |         |
    |      |   Evaluation DAO   |                                   |               |         |  +-------------+  |         |
    |      |                    |                                   |               |         |  |             |  |         |
    |      |Qualification system|                                   |               |         +->|    GnoVM    |<-+         |
    |      |to distribute ^worx |                                   |               |            |             |            |
    |      +--------------------+                                   |               |            +-------------+            |
    |                                                               |               |                   |                   |
    +---------------------------------------------------------------+               +-------------------+-------------------+
                                    ^                                                                   |
                                    |                                                                   |
                                    +-------------------------------------------------------------------+

## Components

### `gno.land`

The main blockchain powered by the `TM2` engine. It offers permissionless smart contracts with the `GnoVM` and can self-configure from contracts using the `GnoSDK`.

### `worxDAO`

The governance entity consisting of contributors, responsible for governing the `r/system` realms, including `valset` and `chaincfg`.

### `r/system/valset`

A realm (smart contract) that enables the `worxDAO` to update the validator set. Similar to a PoA system, the authority is decentralized in a DAO.

Additionally, this contract is queried by `gno.land` to configure `TM2` when changes are made to the validator set.

### `r/system/chaincfg`

A governance-backed smart contract that allows for chain configuration through governance.

It helps prevent unnecessary upgrade campaigns for minor updates.
