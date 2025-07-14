# GnoSwap Launchpad

GnoSwap Launchpad is a decentralized platform for launching early-stage GRC20 projects. Users participate by staking $GNS tokens for a fixed period, earning project tokens as rewards while keeping their principal stake intact. The staked tokens generate yield via the [xGNS Governance Contract](../gov/README.md), which powers project launches. This ensures a fair, transparent, and risk-minimized way to engage in the gno.land ecosystem.

## Key Components

### Launchpad Initialization (`launchpad_init.gno`)
- **Project Creation**: Allows project teams to create launchpad pools, setting parameters like reward allocation, duration, and deposit limits.
- **Pool Management**: Handles the configuration of launchpad pools, defining participation rules and reward structures.

### Launchpad Deposit (`launchpad_deposit.gno`)
- **Deposit Function**: Enables users to lock $GNS tokens in a project’s launchpad pool.
- **Staking Mechanism**: The deposited tokens are staked in the governance contract, generating yield that funds project development.

### Launchpad Reward (`launchpad_reward.gno`)
- **Reward Distribution**: Allocates project tokens to participants based on their staked $GNS.
- **Reward Claiming**: Users can claim their project tokens during or after the pool duration.

## Interaction Flow

1. **Project Creation**: Project teams initialize launchpad pools with defined parameters.
2. **User Participation**: Users deposit $GNS into a launchpad pool.
3. **Reward Accumulation**: Participants earn project tokens over time based on their staked amount.
4. **Reward Claiming**: Users claim their earned tokens during or after the event.
5. **Principal Return**: Once the launchpad period ends, users retrieve their original $GNS deposit.

## Important Notes for Participants

- **Deposit Lock-up**: Deposited $GNS is locked until the pool reaches maturity.
- **Yield Usage**: Yield from deposited $GNS is used to support project development.
- **Reward Type**: Participants earn project tokens, distributed based on their share of deposits.
- **Governance Power**: Deposited $GNS does not grant additional governance influence.

For more details, visit the [GnoSwap Launchpad Documentation](https://docs.gnoswap.io/core-concepts/launchpad).
