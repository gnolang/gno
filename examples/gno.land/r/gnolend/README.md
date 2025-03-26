# Gnolend

Gnolend is a lending protocol built on Gno.land that leverages Gnoswap's liquidity pools for price feeds and collateral management. It provides a decentralized lending market where users can participate in lending and borrowing activities with any token pair that has sufficient liquidity in Gnoswap.

## Prerequisites

- GNU Make 3.81 or higher
- Latest version of [gno.land](https://github.com/gnolang/gno)
- Go 1.21 or higher

## Setup

1. First, follow the setup instructions from [Gnoswap](https://github.com/gnoswap-labs/gnoswap) to set up your development environment.

2. Add Gnolend realms to your gno repository:

   ```bash
   # Copy Gnolend realms to examples/r/gnolend
   cp -r gnolend/* $WORKDIR/gno.land/examples/r/gnolend/
   ```

3. Configure admin addresses:

   For proper testing and development, you'll need to update the admin addresses in several locations:

   - In Gnolend test files: Update the admin addresses to match your test accounts
   - In Gnoswap test files: Ensure the admin addresses align with your test environment
   - To avoid manual token minting, you can aloso modify the admin address in `p/gnoswap/consts.gno` to match your preferred test account

4. Run tests:

   Before running Gnolend tests, you'll need to first initialize a WUGNOT-GNS pool in Gnoswap and mint at least one liquidity position. This provides the price oracle that Gnolend depends on. After that's done, you can proceed with testing the Gnolend functions.
