# Router

## Overview

The **Router** in GnoSwap is responsible for executing token swaps and managing swap routes. It ensures efficient trade execution across various liquidity pools while supporting both single and multi-hop swaps.

## Key Features

- **Swap Execution (`SwapRoute`)**: Executes token swaps based on specified routes and swap types (`ExactInSwapRoute` or `ExactOutSwapRoute`). Supports both single and multi-hop swaps, handling up to 3-7 routes.
- **Swap Simulation (`DrySwapRoute`)**: Simulates swap routes without executing the actual swap, allowing users to estimate swap outcomes.
- **Router Fee Management**: Implements a router fee for swaps, which can be adjusted by governance.
- **Native and Wrapped Token Support**: Facilitates swaps involving both native `GNOT` and wrapped `WUGNOT` tokens.

## Functionality

1. **Swapping Tokens**
   - Users can execute swaps through predefined routes.
   - Supports both exact input (`ExactInSwapRoute`) and exact output (`ExactOutSwapRoute`) swap types.
   - Routes swaps efficiently to optimize slippage and execution price.

2. **Multi-Hop Routing**
   - Enables swapping across multiple pools to achieve the best price.
   - Supports up to 7-hop routes for complex swaps.

3. **Swap Simulation**
   - Allows users to preview potential outcomes before executing swaps.
   - Helps traders assess slippage, fees, and expected output.

4. **Router Fee Management**
   - A percentage of the swap fee is collected as a router fee.
   - Governance can adjust the fee parameters.

For more details, visit [GnoSwap Docs](https://docs.gnoswap.io/contracts/router).