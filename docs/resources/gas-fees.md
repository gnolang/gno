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
to consume. If your transaction requires more gas than this limit, it fails
with an "out of gas" error and its effects are rolled back.

### Gas Fee

`--gas-fee` is the whole fee for the transaction, one flat amount. It is
expressed in `ugnot` (micro-GNOT), so `1000000ugnot` is one GNOT.

Gas wanted turns that fee into a rate, and the rate is what the network checks:

```
Gas Price = Gas Fee ÷ Gas Wanted
```

`gnokey` simulates first, and a transaction that fails there costs nothing. It
can still fail on chain afterwards, and once it is in a block you pay the whole
fee, whether it used less gas than it asked for or failed outright.

### Calculating Your Gas Fee

Your `--gas-fee` divided by your `--gas-wanted` must meet or exceed the
[network gas price](#gas-price) for your transaction to be accepted.

The easiest way is to use [`-simulate only`](#gas-estimation), which automatically queries the
current gas price and calculates the recommended fee (with a 5% buffer).

## Gas Price

The network dynamically adjusts the minimum required gas price after each block
based on demand. This ensures the network responds to congestion by increasing
prices when usage is high and decreasing them when usage is low.

### How Gas Price Works

The gas price is returned as a `GasPrice` object with two fields:
- `gas` - the gas units (e.g., 1000)
- `price` - the price for those gas units (e.g., "1ugnot")

Together, these represent a **rate**. `{gas: 1000, price: "1ugnot"}` means 1
ugnot per 1000 gas units, which simplifies to 0.001 ugnot per gas unit. That is
the floor the price never drops below, and where the networks in this guide sit
today. Under load the `price` field grows.

To calculate the minimum fee manually:

1. [Query](#querying-gas-price) the current gas price
2. Calculate the rate: `price / gas`
3. Multiply by your `--gas-wanted`

**Example:**
```bash
# Query returns: {gas: 1000, price: "1ugnot"}
# Rate = 1 ÷ 1000 = 0.001 ugnot/gas

# If you want --gas-wanted 2000000:
# Minimum fee = 2,000,000 × 0.001 = 2,000 ugnot
# So set: --gas-fee 2000ugnot (or higher)
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
`min_gas_prices` configuration parameter in their `config.toml` file. A fee that
meets the network price but not a given validator's own minimum is simply left
out of that validator's blocks, so a transaction priced at the bare minimum can
wait longer. The 5% buffer from `-simulate only` covers the usual case.

## Typical Gas Values

Here are some recommended gas values for common operations:

| Operation                 | Recommended Gas Wanted | Gas Fee at 1ugnot/1000gas |
| ------------------------- | ---------------------- | ------------------------- |
| Simple transfer           | 100,000                | 100ugnot                  |
| Calling a realm function  | 2,000,000              | 2000ugnot                 |
| Deploying a small package | 5,000,000              | 5000ugnot                 |
| Deploying a complex realm | 10,000,000+            | 10000ugnot                |

The fee column is the minimum the network accepts at the initial gas price, one
`ugnot` per 1000 gas. It rises with the gas price, so query the current one or
run `-simulate only` rather than copying these numbers into a script.

## Gas Estimation

Use the `-simulate only` flag to estimate gas usage and the recommended fee
without executing on-chain or incrementing the account sequence number:

```bash
gnokey maketx addpkg \
  -pkgdir "./hello_world" \
  -pkgpath gno.land/p/examplenamespace/hello_world \
  -gas-wanted 4000000 \
  -gas-fee 4000ugnot \
  -remote https://rpc.staging.gno.land:443 \
  -chainid staging \
  -simulate only \
  YOUR_KEY_NAME
```

Simulation output, so nothing reached a block:
```
OK!
GAS WANTED: 4000000
GAS USED:   2590066
HEIGHT:     0
STORAGE DELTA:  1748 bytes
STORAGE FEE:    174800ugnot
TOTAL TX COST:  178800ugnot
EVENTS:     [{"bytes_delta":1748,"fee_delta":{"denom":"ugnot","amount":174800},"pkg_path":"gno.land/p/examplenamespace/hello_world"}]
INFO:       estimated gas usage: 2590066 (suggested, with 5% margin: 2719570), gas fee: 2720ugnot, current gas price: 1ugnot/1000gas

TX HASH:
PKGPATH:    gno.land/p/examplenamespace/hello_world
```

That is why `HEIGHT` is 0 and `TX HASH` is empty. `STORAGE FEE` is the
[storage deposit](storage-deposit.md), locked rather than spent, and
`TOTAL TX COST` adds it to the gas fee.

Take the suggested figure, 2719570 here, not the raw estimate: gas shifts
between the simulation and the broadcast, and the 5% margin absorbs it. Do not
take the printed `gas fee` with it, which prices the raw estimate and can leave
the pair under what the chain requires. Price the limit yourself and round up:
at 1ugnot per 1000 gas, 2719570 gas needs 2720ugnot.

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

**Insufficient fees:** `insufficient fees; got: {Gas-Wanted: 2000000, Gas-Fee
1000ugnot}, fee required: 1ugnot/1000gas as block gas price`
- Your `--gas-fee` is too low. Increase it to meet the minimum required.

**Out of gas:** `gas used (2597634) exceeds tx's gas wanted (1000000) during
operation: simulation`
- Your `--gas-wanted` is too low. Use `-simulate only` to estimate needed gas,
  then increase.
- Nothing was charged. `operation: simulation` means it never reached a block,
  because `gnokey maketx -broadcast` simulates before sending.
- ⚠️ **A transaction that reaches a block and then fails pays the whole
  `--gas-fee`.** Its error names the real operation instead, like
  `operation: CPUCycles`. You get there with `-simulate skip`, or when the chain
  moves between the simulation and the real run.

