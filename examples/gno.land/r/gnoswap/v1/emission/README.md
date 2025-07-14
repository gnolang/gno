# Emission

## Overview

The emission contract manages minting and distributing the `GNS` token. The GNS tokens are minted and distributed through the `MintAndDistributeGns()` function, which is a public function that can be called by anyone. This mechanism ensures that token distribution is executed dynamically as user interactions occur in the GnoSwap protocol.

## Features

### Token Emission and Distribution

- **Block-Time-Based Distribution**: The GNS distribution is designed to follow the network’s block generation time, initially set [here](https://github.com/gnoswap-labs/gnoswap/blob/main/contract/p/gnoswap/consts/consts.gno#L126). This value can be adjusted through governance if needed, such as in response to network delays.
- **Triggering Mechanism**: The MintAndDistributeGns() function is integrated into key user transaction entry points. This means that whenever users interact with GnoSwap through transactions, the function is triggered, making it appear as though GNS is minted and distributed on a per-block basis, as long as GnoSwap remains active.
- **Distribution to Defined Targets**: Each call to MintAndDistributeGns() mints new tokens and distributes them to predefined targets, ensuring continuous and fair token emission.

### Distribution Targets

Tokens are distributed to four main targets as follows by default:

- **Liquidity Staker**: 75%
- **DevOps Team**: 20%
- **Community Pool**: 5%
- **Governance Stakers**: 0%

### Distribution Ratio Management

- Governance or admin can adjust the distribution ratio.
- Distribution percentages are tracked in basis points (1 bp = 0.01%).
- Total distribution must always equal 100% (10,000 basis points).

### Undistributed Token Handling

- Any tokens not distributed due to skipped or delayed calls, they are tracked.
- Undistributed tokens are included in future distribution cycles, ensuring no loss of allocation.

### Halving

There is a halving mechanism that reduces the issuance amount by half every two years. For details, refer to [halving.gno](../gns/halving.gno).

### Callback Mechanisms

Provides callback function for inter-contract communication. This will notify relevant components when distribution ratios change. Details to be found on our [bridge contract](https://github.com/gnoswap-labs/gnoswap/tree/main/contract/r/gnoswap/bridge#bridge).

For more details about the emission, check out the [GnoSwap docs](https://docs.gnoswap.io/gnoswap-token/emission).
