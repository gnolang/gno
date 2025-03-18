# Emission

## Overview

The emission contract manages minting and distributing the `GNS` token.

## Features

### Token Emission and Distribution

- Mints `GNS` tokens on a block generation time basis (initially set [here](https://github.com/gnoswap-labs/gnoswap/blob/main/contract/p/gnoswap/consts/consts.gno#L126), but can be changed by governance)
- Distributes tokens to the set targets
- Tracks undistributed tokens and include it in future distribution cycles

### Distribution Targets

Tokens are distributed to four main targets as follows by default:

- **Liquidity Staker**: 75%
- **DevOps Team**: 20%
- **Community Pool**: 5%
- **Governance Stakers**: 0%

### Distribution Ratio Management

- Governance or admin can adjust token distribution ratio
- Distribution percentages are tracked in basis points (1 bp = 0.01%)
- Total distribution must always equal 100% (10,000 basis points)

### Undistributed Token Handling

- Any tokens not distributed in a cycle are tracked
- Undistributed tokens are included in the next distribution cycle

### Halving

There is a halving mechanism that reduces the issuance amount by half every two years

### Callback Mechanisms

Provides callback function for inter-contract communication. This will notify relevant components when distribution ratios change. Details to be found on our [bridge contract](https://github.com/gnoswap-labs/gnoswap/tree/main/contract/r/gnoswap/bridge#bridge).

For more details about the emission, check out the [GnoSwap docs](https://docs.gnoswap.io/gnoswap-token/emission).
