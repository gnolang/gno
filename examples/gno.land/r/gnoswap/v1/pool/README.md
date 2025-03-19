# Pool

## Overview

**Pool** is a core component of GnoSwap, designed as a smart contract that facilitates liquidity provision and trading between two GRC20 tokens. Unlike traditional models where each liquidity pool has a separate contract, **Pool** adopts a **single skeleton design**, meaning all pools exist within a single contract. This approach enhances efficiency, reduces deployment overhead, and aligns with Uniswap V4's singleton architecture.

## Key Features

- **Single Skeleton Architecture**: All liquidity pools are managed within a single contract, optimizing gas efficiency and simplifying interactions.
- **Concentrated Liquidity**: Liquidity providers (LPs) can specify custom price ranges for their liquidity, maximizing capital efficiency.
- **Multiple Fee Tiers**: Supports various fee tiers to accommodate different trading strategies and risk appetites.
- **Dynamic Liquidity Adjustments**: Liquidity adapts to price fluctuations automatically, ensuring seamless trading and efficient market-making.
- **GRC20 Token Support**: Designed exclusively for GRC20 token pairs, enabling decentralized trading within the GnoSwap ecosystem.

## How It Works

1. **Liquidity Provision**: Users deposit two GRC20 tokens into the contract within a specified price range, defining their liquidity position.
2. **Trading**: Traders swap tokens within the pool, utilizing the available liquidity at the current market price.
3. **Fee Collection**: Each trade incurs a fee based on the selected fee tier, distributed proportionally to liquidity providers.
4. **Liquidity Adjustment**: As market prices change, LPs can adjust or remove their liquidity positions as needed.

## Advantages of Single Skeleton Design

- **Gas Efficiency**: Eliminates the need to deploy multiple contracts for each pool, reducing transaction costs.
- **Unified Management**: A single contract governs all pools, simplifying governance, upgrades, and security audits.
- **Optimized User Experience**: Traders and liquidity providers interact with a single contract, streamlining operations and enhancing composability within GnoSwap.

For more details, refer to [the smart contract documentation and API reference](https://docs.gnoswap.io/contracts/pool/pool.gno).