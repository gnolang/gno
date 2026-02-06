# `subscriptions` - Recurring payment management

Manage subscription-based services with automated recurring payments using native coins and GRC20 tokens.

## Features

- Support for native coins and GRC20 tokens
- Automatic charging at renewal periods
- Prorated refunds on unsubscribe
- Mix native and GRC20 tokens in same subscription

## Usage

```go
import "gno.land/r/samcrew/subscriptions"

// Create a new service (simple helper - recommended)
subscriptions.NewServiceSimple(
    cross,
    "Premium Plan",
    "Access to premium features",
    30 * 24 * time.Hour, // 30 days
    "ugnot",             // native denomination
    1000000,             // 1 GNOT
    "",                  // no GRC20
    0,                   // no GRC20 amount
)

// Or create with GRC20 tokens
subscriptions.NewServiceSimple(
    cross,
    "Gold Plan",
    "Premium tier with tokens",
    30 * 24 * time.Hour,
    "",                       // no native coins
    0,
    "gno.land/r/demo/foo20",  // GRC20 token
    5000000,                  
)

// Subscribe to a service with native coins
subscriptions.Subscribe(cross, "Premium Plan", "", 0)

// Subscribe with GRC20 tokens
subscriptions.Subscribe(cross, "Premium Plan", "gno.land/r/demo/foo20", 5000000)

// Top up subscription balance
subscriptions.Topup(cross, "Premium Plan", "", 0) // native coins
subscriptions.Topup(cross, "Premium Plan", "gno.land/r/demo/foo20", 1000000) // GRC20

// Unsubscribe (refunds remaining balance)
subscriptions.Unsubscribe(cross, "Premium Plan")

// Service owner withdraws earned fees
subscriptions.ServiceClaimVault(cross, "Premium Plan")
```

## API Reference

```go
// Create a new subscription service (simple helper - recommended)
// - displayName: Service identifier
// - description: Service description
// - renewalPeriod: Duration between payments (e.g., 30 days)
// - nativeDenom: Native coin denomination (e.g., "ugnot"), empty for none
// - nativeAmount: Amount of native coins, 0 for none
// - grc20Denom: GRC20 token fully qualified name, empty for none
// - grc20Amount: Amount of GRC20 tokens, 0 for none
func NewServiceSimple(cur realm, displayName, description string, renewalPeriod time.Duration, 
    nativeDenom string, nativeAmount int64, grc20Denom string, grc20Amount int64)

// Create a new subscription service (advanced)
// - displayName: Service identifier
// - description: Service description
// - renewalPeriod: Duration between payments (e.g., 30 days)
// - native: Native coins price
// - grc20: GRC20 tokens price
func NewService(cur realm, displayName, description string, renewalPeriod time.Duration, 
    native chain.Coins, grc20 chain.Coins)

// Subscribe to a service
// - serviceName: Name of the service
// - fqName: Fully qualified GRC20 token name (empty for native)
// - amount: Initial GRC20 deposit (0 for native)
func Subscribe(cur realm, serviceName string, fqName string, amount int64)

// Unsubscribe from a service (refunds remaining balance)
func Unsubscribe(cur realm, serviceName string)

// Top up subscription balance
// - serviceName: Name of the service
// - fqName: GRC20 token name (optional)
// - amount: Amount to add
func Topup(cur realm, serviceName string, fqName string, amount int64)

// Service owner withdraws earned fees
func WithdrawFees(cur realm, serviceName string)
```
