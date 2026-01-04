# Gas Fees in Gno.land

This document explains how gas works in the Gno.land ecosystem, including gas
pricing, estimation, and optimization.

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

You will be charged the gas fee you specified.

## Typical Gas Values

Here are some recommended gas values for common operations:

| Operation                 | Recommended Gas Wanted | Typical Gas Fee |
| ------------------------- | ---------------------- | --------------- |
| Simple transfer           | 100,000                | 1000000ugnot    |
| Calling a realm function  | 2,000,000              | 1000000ugnot    |
| Deploying a small package | 5,000,000              | 1000000ugnot    |
| Deploying a complex realm | 10,000,000+            | 1000000ugnot    |

These values may vary based on network conditions and the specific
implementation of your code.

## Gas Estimation

Use the `-simulate only` flag to simulate a transaction without executing it
on-chain. This allows you to estimate gas usage and fees without incurring
any cost or incrementing the account sequence number.

```bash
gnokey maketx addpkg \
  -pkgdir "./hello" \
  -pkgpath gno.land/r/hello \
  -gas-wanted 2000000 \
  -gas-fee 1000000ugnot \
  -remote https://rpc.gno.land:443 \
  -broadcast \
  -chainid staging \
  -simulate only \
  YOUR_KEY_NAME
```
You will see output similar to the following:
```
GAS WANTED: 2000000
GAS USED:   268994
INFO:       estimated gas usage: 268994, gas fee: 282ugnot, current gas price: 1ugnot/1000gas
```
You can then use the estimated gas as the -gas-wanted and -gas-fee values in your
actual transaction.

## Example Transaction with Gas Parameters

Here's an example of sending a transaction with gas parameters:

```bash
gnokey maketx call \
  -pkgpath "gno.land/r/demo/boards" \
  -func "CreateBoard" \
  -args "MyBoard" \
  -args "Board description" \
  -gas-fee 1000000ugnot \
  -gas-wanted 2000000 \
  -remote https://rpc.gno.land:443 \
  -chainid staging \
  YOUR_KEY_NAME
```

## Gas Optimization Tips

To minimize gas costs, consider these optimization strategies:

1. **Minimize on-chain storage**: Only store essential data on-chain
2. **Batch operations**: Combine multiple operations into a single transaction when possible
3. **Use efficient data structures**: Well-optimized code consumes less gas
4. **Precompute values off-chain**: Do as much computation as possible before submitting to the blockchain
5. **Test locally first**: Use `gnodev` to test and optimize your code before deploying to a network
