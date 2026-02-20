# `normalizedcoins` - Unified coin handling for native and GRC20 tokens

Handle both native coins and GRC20 tokens through a single interface by prefixing denominations.

## Features

- Transform native and GRC20 coins into a unified format
- Arithmetic operations on mixed coin types
- Send both native and GRC20 tokens with a single function call

## Usage

```go
import "gno.land/r/samcrew/normalizedcoins"

// Convert coins to normalized format
native := chain.NewCoins(chain.NewCoin("ugnot", 1000))
grc20 := chain.NewCoins(chain.NewCoin("gno.land/r/demo/foo20", 500))

normalized, err := normalizedcoins.PrefixCoins(native, grc20)
// Result: ["/native/ugnot": 1000, "/grc20/r/demo/foo20": 500]

// Perform arithmetic on mixed coin types
balance := chain.NewCoins(
    chain.NewCoin("/native/ugnot", 5000),
    chain.NewCoin("/grc20/r/demo/foo20", 2000),
)

cost := chain.NewCoins(
    chain.NewCoin("/native/ugnot", 1000),
    chain.NewCoin("/grc20/r/demo/foo20", 500),
)

remaining := normalizedcoins.SubCoins(balance, cost)
// remaining: ["/native/ugnot": 4000, "/grc20/r/demo/foo20": 1500]

// Send mixed coins to an address
normalizedcoins.SendCoins(recipientAddress, remaining)
// Automatically routes native coins through banker and GRC20 through token contracts
```

## Implementation Details

The package works by prefixing coin denominations with their type:
- Native coins: `ugnot` becomes `/native/ugnot`
- GRC20 tokens: `gno.land/r/demo/foo20` becomes `/grc20/r/demo/foo20`

This allows treating different coin types uniformly while maintaining the ability to route them correctly during transfers.
The `sendCoins` function strips prefixes and uses the appropriate transfer mechanism for each type.

## Related

- [Subscriptions](../subscriptions) - Uses normalizedcoins for subscription payment handling
- [GRC20 Registry](../../demo/defi/grc20reg) - Token registration system
