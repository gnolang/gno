# GnoSwap Governance

The GnoSwap governance contract allows $GNS holders to participate in protocol decision-making through staking, delegation, proposal creation, and voting. By staking GNS, users receive xGNS, which represents voting power and enables them to propose and vote on governance changes. Additionally, xGNS holders earn a share of [protocol fees](../protocol_fee/README.md), providing an incentive for active participation. The governance system is designed to be transparent, decentralized, and fully on-chain.

## Overview

The governance system comprises several key realms:

- **Staker Realm (`staker.gno`)**: Manages staking, delegation, and reward collection.
- **Proposal Realm (`proposal.gno`)**: Facilitates the creation and management of governance proposals.
- **Vote Realm (`vote.gno`)**: Oversees the voting process on active proposals.

## Key Components

### Staker Realm (`staker.gno`)

- **Delegation Functions:**
  - `Delegate(to std.Address, amount uint64)`: Delegate voting power to a specified address.
  - `Redelegate(from std.Address, to std.Address, amount uint64)`: Reassign delegated voting power to a different delegate.
  - `Undelegate(from std.Address, amount uint64)`: Retract delegated voting power.

- **Reward Functions:**
  - `CollectReward()`: Collect accumulated rewards based on delegated tokens.
  - `CollectUndelegatedGns()`: Collect undelegated GNS tokens after a lock period (7 days).

### Proposal Realm (`proposal.gno`)

- **Proposal Creation**: Users can submit proposals suggesting protocol changes or new features. Each proposal includes parameters such as the proposal's content, the proposer’s address, and the submission timestamp.

### Vote Realm (`vote.gno`)

- **Voting Process**: Users cast their votes on active proposals during the voting period. The system records and tallies these votes to determine the outcome of each proposal.

## Interaction Flow

1. **Staking & Delegation**: Users stake their $GNS tokens to receive xGNS, representing their voting power. They can delegate this voting power to themselves or other delegates.

2. **Proposal Creation**: With sufficient GNS, users can create proposals suggesting protocol changes or new features.

3. **Voting**: During the voting period, xGNS holders cast their votes on active proposals. The outcome is determined based on the majority of votes and predefined quorum requirements.

4. **Execution**: Approved proposals are executed, implementing the proposed changes within the protocol.

For more detailed information about the rationale behind the governance module, please refer to [GnopSwap Docs](https://docs.gnoswap.io/core-concepts/governance).
