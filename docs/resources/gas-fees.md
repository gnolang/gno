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

You will be charged the gas fee you specified.

### Calculating Your Gas Fee

Your `--gas-fee` must meet or exceed the [network gas price](#gas-price) for your transaction
to be accepted.

The easiest way is to use [`-simulate only`](#gas-estimation), which automatically queries the
current gas price and calculates the recommended fee (with a 5% buffer).

## Gas Price

The network dynamically adjusts the minimum required gas price after each block
based on demand. This ensures the network responds to congestion by increasing
prices when usage is high and decreasing them when usage is low.

### How Gas Price Works

The gas price is returned as a `GasPrice` object with two fields:
- `gas` - the gas units (e.g., 1000)
- `price` - the price for those gas units (e.g., "100ugnot")

Together, these represent a **rate**. For example, `{gas: 1000, price: "100ugnot"}`
means the minimum rate is 100 ugnot per 1000 gas units, which simplifies to
0.1 ugnot per gas unit.

To calculate the minimum fee manually:

1. [Query](#querying-gas-price) the current gas price
2. Calculate the rate: `price / gas`
3. Multiply by your `--gas-wanted`

**Example:**
```bash
# Query returns: {gas: 1000, price: "100ugnot"}
# Rate = 100 ÷ 1000 = 0.1 ugnot/gas

# If you want --gas-wanted 2000000:
# Minimum fee = 2,000,000 × 0.1 = 200,000 ugnot
# So set: --gas-fee 200000ugnot (or higher)
```

### Querying Gas Price

You can query the current network gas price using:
```bash
gnokey query auth/gasprice -remote https://rpc.gno.land:443
```

This returns the gas price calculated from the most recently completed block,
which is the minimum rate currently required for new transactions.

For more details, see [`auth/gasprice`](../users/interact-with-gnokey.md#authgasprice).

### How the Network Adjusts Gas Price

The network automatically adjusts the gas price after each block based on demand:

- **Low demand**: Price decreases (but never below 1 ugnot/1000 gas)
- **High demand**: Price increases

The network targets 70% utilization of the maximum block gas limit (3B gas) by default.
When blocks exceed this target, prices rise. When blocks fall below it, prices drop.
Changes are gradual to avoid sudden price spikes.

**Note**: Individual validators can also set their own minimum gas price through the
`min_gas_prices` configuration parameter in their `config.toml` file. Different validators may have different minimums.

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

Use the `-simulate only` flag to estimate gas usage and the recommended fee
without executing on-chain or incrementing the account sequence number:

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

Output:
```
GAS WANTED: 2000000
GAS USED:   268994
INFO:       estimated gas usage: 268994, gas fee: 282ugnot, current gas price: 1ugnot/1000gas
```

Use the `estimated gas usage` and `gas fee` values as your `-gas-wanted` and `-gas-fee`
for the actual transaction.

## Gas Optimization Tips

To minimize gas costs, consider these optimization strategies:

1. **Minimize on-chain storage**: Storage writes are the most expensive operations
   (2,000 gas flat + 30 gas/byte for writes vs. 1,000 gas flat + 3 gas/byte for
   reads). Only store essential data on-chain.
2. **Batch operations**: Combine multiple operations into a single transaction
   when possible, reducing the overhead of per-transaction costs (signature
   verification, etc.).
3. **Use efficient data structures**: Well-optimized code consumes less gas.
   Avoid unnecessary iterations — each iterator step costs gas.
4. **Precompute values off-chain**: Do as much computation as possible before
   submitting to the blockchain.
5. **Test locally first**: Use `gnodev` to test and optimize your code before
   deploying to a network.

## Common Errors

**Insufficient fees:** `insufficient fees; got: 50000ugnot required: 200000ugnot`
- Your `--gas-fee` is too low. Increase it to meet the minimum required.

**Out of gas:** `out of gas in location: ... wanted: 100000, used: 150000`
- Your `--gas-wanted` is too low. Use `-simulate only` to estimate needed gas, then increase.
- ⚠️ **You're still charged for failed transactions!** Fees are deducted before
  execution, so even if your transaction runs out of gas, you pay for the gas
  consumed up to the limit.

> **Note:** By default, `gnokey maketx -broadcast` uses `-simulate test`, which
> simulates the transaction before submitting it. If the simulation fails (e.g.,
> out of gas), the transaction won't be submitted and you won't be charged.
> However, a transaction that passes simulation can still fail on-chain, in
> which case you will be charged.

