# Protocol Fee Contract

The GnoSwap Protocol Fee Contract manages fees collected from various platform interactions, ensuring proper distribution to $xGNS holders. This contract encompasses fees from swaps, pool creations, liquidity withdrawals, and staking reward claims.

## Overview

The Protocol Fee Contract is integral to GnoSwap's revenue model, collecting fees from key operations and distributing them to $xGNS holders. This mechanism incentivizes staking and supports the platform's sustainability.

## Key Components

### Swap Fee (`protocol_fee_swap.gno`)

- **GetSwapFee**: Retrieves the current swap fee rate.
- **SetSwapFee**: Allows authorized entities to modify the swap fee rate.

The default swap fee is set to **0.15%** of the total swap amount.

### Pool Creation Fee (`protocol_fee_pool_creation.gno`)

- **GetPoolCreationFee**: Returns the current pool creation fee.
- **SetPoolCreationFee**: Enables authorized entities to adjust the pool creation fee.

The standard pool creation fee is **100 GNS**.

### Withdrawal Fee (`protocol_fee_withdrawal.gno`)

- **GetWithdrawalFee**: Fetches the current withdrawal fee rate.
- **SetWithdrawalFee**: Permits authorized entities to change the withdrawal fee rate.
- **HandleWithdrawalFee**: Calculates and processes the fee during liquidity withdrawal.

The default withdrawal fee is **1%** of the liquidity provider's claimed fees.

### Unstaking Fee (`protocol_fee_unstaking.gno`)

- **GetUnstakingFee**: Obtains the current unstaking fee rate.
- **SetUnstakingFee**: Allows authorized entities to set a new unstaking fee rate.

The default unstaking fee is **1%** of the staking rewards claimed.

## Interaction Flow

1. **Swaps**: Users execute token swaps, incurring a **0.15%** fee on the total swap amount.
2. **Pool Creation**: users create new liquidity pools, paying a **100 GNS** fee.
3. **Liquidity Withdrawal**: Liquidity providers withdraw their positions, with **1%** of the claimed fees deducted as a withdrawal fee.
4. **Staking Reward Claims**: Stakers claim their rewards, incurring a **1%** unstaking fee on the claimed amount.

All collected fees are directed to the Protocol Fee Contract and subsequently distributed to $xGNS holders.

## Important Notes for Participants

- **Fee Awareness**: Users should be aware of the applicable fees for swaps, pool creations, liquidity withdrawals, and staking reward claims.
- **Staking Incentives**: Holding $xGNS entitles users to a share of the protocol fees, promoting active participation in governance and staking.

For more detailed information, please refer to the [GnoSwap xGNS Documentation](https://docs.gnoswap.io/gnoswap-token/xgns).
