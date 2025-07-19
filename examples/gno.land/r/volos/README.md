# Volos

Volos is the first lending protocol built using Gnolang, implementing financial primitives for decentralized lending and borrowing. The protocol features lending markets with configurable parameters, variable interest rate models, collateralized borrowing with health monitoring, and liquidation mechanisms for undercollateralized positions. It employs a shares-based accounting system to track user positions, calculates interest based on utilization metrics, and maintains system solvency through real-time risk assessment.

For price determination, Volos integrates with [Gnoswap](https://github.com/gnoswap-labs/gnoswap)'s liquidity pools, using them as price oracles in the absence of dedicated oracle infrastructure. This approach enables the protocol to obtain reliable price data directly from on-chain sources without requiring external oracle networks, demonstrating how essential financial primitives can be implemented within the current gno.land ecosystem.

The system calculates borrowing capacity based on collateral values derived from Gnoswap pool prices. This integration creates a self-contained lending solution that maintains the security guarantees of the underlying blockchain while providing the necessary infrastructure for expanding DeFi capabilities on gno.land.

> **Warning:** This project is work in progress. The protocol is under active development and contains incomplete features and known issues.

## Prerequisites

- GNU Make 3.81 or higher
- Latest version of [gno.land](https://github.com/gnolang/gno)
- Go 1.21 or higher

## Setup

1. First, follow the setup instructions from [Gnoswap](https://github.com/gnoswap-labs/gnoswap) to set up your development environment.

2. Add Volos realms to your gno repository:

   ```bash
   cp -r volos/* $WORKDIR/gno.land/examples/r/volos/
   ```

3. Configure admin addresses:

   For proper testing and development, you'll need to update the admin addresses in several locations:

   - In Volos test files: Update the admin addresses to match your test accounts
   - In Gnoswap test files: Ensure the admin addresses align with your test environment
   - To avoid manual token minting, you can aloso modify the admin address in `p/gnoswap/consts.gno` to match your preferred test account

4. Run tests:

   Before running Volos tests, you'll need to first initialize a WUGNOT-GNS pool in Gnoswap and mint at least one liquidity position. This provides the price oracle that Volos depends on. After that's done, you can proceed with testing the Volos functions.
