# Gas Fees in gno.land

This document explains how gas works in the gno.land ecosystem, including gas
pricing, estimation, and optimization.

## What is Gas?

Gas is a measure of computational and storage resources required to execute
operations on the blockchain. Every transaction on gno.land consumes gas based
on:

1. The complexity of the operation being performed
2. The amount of data being stored
3. The current network conditions

Gas serves several important purposes:
- Prevents spam and denial-of-service attacks
- Allocates network resources fairly among users
- Compensates validators for the computational resources they provide

## Gas Parameters

When submitting transactions to gno.land, you need to specify two gas-related parameters:

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

## Gas Estimation

Currently, gno.land doesn't provide automatic gas estimation. You need to:

1. Start with conservative (higher) gas values
2. Observe actual gas usage from transaction receipts
3. Adjust your gas parameters for future transactions

Future updates to the platform will likely include improved gas estimation tools.

## Example Transaction with Gas Parameters

Here's an example of sending a transaction with gas parameters:

```bash
gnokey maketx call \
  --pkgpath "gno.land/r/demo/boards" \
  --func "CreateBoard" \
  --args "MyBoard" "Board description" \
  --gas-fee 1000000ugnot \
  --gas-wanted 2000000 \
  --remote https://rpc.gno.land:443 \
  --chainid portal-loop \
  YOUR_KEY_NAME
```

## Gas Optimization Tips

To minimize gas costs, consider these optimization strategies:

1. **Minimize on-chain storage**: Only store essential data on-chain
2. **Batch operations**: Combine multiple operations into a single transaction when possible
3. **Use efficient data structures**: Well-optimized code consumes less gas
4. **Precompute values off-chain**: Do as much computation as possible before submitting to the blockchain
5. **Test locally first**: Use `gnodev` to test and optimize your code before deploying to a network
