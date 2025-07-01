# Gas Fees in Gno.land

This document explains how gas works in the Gno.land ecosystem, including 
automatic gas estimation, manual control, and optimization.

## What is Gas?

Gas is a measure of computational and storage resources required to execute
operations on the blockchain. Every transaction on Gno.land consumes gas based
on:

1. The complexity of the operation being performed
2. The amount of data being stored
3. The current network conditions

Gas serves several important purposes:
- Prevents spam and denial-of-service attacks
- Allocates network resources fairly among users
- Compensates validators for the computational resources they provide

## Gas Parameters

When submitting transactions to Gno.land, you need to specify two gas-related parameters:

### Gas Wanted

`--gas-wanted` specifies the maximum amount of gas your transaction is allowed
to consume. If your transaction requires more gas than this limit, it will fail
with an "out of gas" error, but will still consume the gas up to that point.

### Gas Fee

`--gas-fee` specifies how much you're willing to pay per unit of gas. This is
typically expressed in `ugnot` (micro-GNOT). For example, `1000000ugnot` means
you're willing to pay 1 GNOT per unit of gas.

The total maximum fee you might pay is calculated as:
```
Maximum Fee = Gas Wanted × Gas Fee
```

However, you'll only be charged for the gas actually used.

## Typical Gas Values

Here are some recommended gas values for common operations:

| Operation | Recommended Gas Wanted | Typical Gas Fee |
|-----------|------------------------|----------------|
| Simple transfer | 100,000 | 1000000ugnot |
| Calling a realm function | 2,000,000 | 1000000ugnot |
| Deploying a small package | 5,000,000 | 1000000ugnot |
| Deploying a complex realm | 10,000,000+ | 1000000ugnot |

These values may vary based on network conditions and the specific
implementation of your code.

## Automatic Gas Estimation

Gno.land provides **automatic gas estimation** by default through `gnokey`. 
This feature:

1. **Simulates your transaction** to determine exact gas requirements
2. **Automatically adds a safety buffer** to prevent "out of gas" errors
3. **Queries current gas prices** for optimal fee calculation
4. **Eliminates guesswork** - no need to manually set gas parameters

### Using Auto Gas

```bash
# No gas parameters needed - automatic estimation
gnokey maketx call \
  --pkgpath "gno.land/r/demo/boards" \
  --func "CreateBoard" \
  --args "MyBoard" "Board description" \
  --remote https://rpc.gno.land:443 \
  --chainid staging \
  YOUR_KEY_NAME
```

### Manual Gas Control (Advanced)

For users who need precise control over gas parameters:

```bash
# Full manual control
gnokey maketx call \
  --pkgpath "gno.land/r/demo/boards" \
  --func "CreateBoard" \
  --args "MyBoard" "Board description" \
  --gas-fee 1000000ugnot \
  --gas-wanted 2000000 \
  --remote https://rpc.gno.land:443 \
  --chainid staging \
  YOUR_KEY_NAME

# Partial auto - estimate gas amount, set fee manually
gnokey maketx call \
  --gas-wanted auto \
  --gas-fee 1000000ugnot \
  # ... other flags

# Partial auto - set gas amount, estimate fee
gnokey maketx call \
  --gas-wanted 2000000 \
  --gas-fee auto \
  # ... other flags
```

## Example Transaction with Gas Parameters

Here are examples of sending transactions:

```bash
# Automatic gas estimation (default)
gnokey maketx call \
  --pkgpath "gno.land/r/demo/boards" \
  --func "CreateBoard" \
  --args "MyBoard" "Board description" \
  --remote https://rpc.gno.land:443 \
  --chainid staging \
  YOUR_KEY_NAME

# Advanced: Manual gas control
gnokey maketx call \
  --pkgpath "gno.land/r/demo/boards" \
  --func "CreateBoard" \
  --args "MyBoard" "Board description" \
  --gas-fee 1000000ugnot \
  --gas-wanted 2000000 \
  --remote https://rpc.gno.land:443 \
  --chainid staging \
  YOUR_KEY_NAME
```

## When to Use Manual Gas Control

While automatic gas estimation is the default, manual control is necessary or 
beneficial in these scenarios:

1. **Airgapped transactions**: When creating unsigned transactions on an 
   offline machine for later signing and broadcasting
2. **Multisig transactions**: When multiple parties need to sign and 
   simulation isn't possible with the required conditions
3. **No simulation node available**: When the simulation endpoint is 
   unavailable or unreachable
4. **Network congestion planning**: To ensure transactions will pass under 
   the heaviest network conditions with explicit higher limits
5. **High-frequency applications**: When you need predictable gas costs for 
   automated systems
6. **Gas price optimization**: During periods of high network activity, you 
   may want to set lower fees and accept slower confirmation times
7. **Testing and development**: When you need precise control for testing 
   gas consumption

## Gas Optimization Tips

To minimize gas costs in your smart contracts, consider these optimization 
strategies:

1. **Minimize on-chain storage**: Only store essential data on-chain
2. **Batch operations**: Combine multiple operations into a single 
   transaction when possible
3. **Use efficient data structures**: Well-optimized code consumes less gas
4. **Precompute values off-chain**: Do as much computation as possible 
   before submitting to the blockchain
5. **Test locally first**: Use `gnodev` to test and optimize your code 
   before deploying to a network
