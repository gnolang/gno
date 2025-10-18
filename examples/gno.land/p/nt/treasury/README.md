# `treasury` - Multi-Banker Payment System

A treasury management system that coordinates multiple payment methods (bankers) for sending funds. Supports different types of payment backends with unified interface and fallback mechanisms.

## Features

- **Multiple bankers**: Support for different payment backends
- **Fallback system**: Try multiple bankers until payment succeeds
- **Payment validation**: Ensure payments are properly processed
- **Banker management**: Add, remove, and list available bankers
- **Unified interface**: Single API for different payment methods

## Usage

```go
import "gno.land/p/nt/treasury"

// Create bankers (implement the Banker interface)
banker1 := &GnoBanker{}
banker2 := &TokenBanker{}

// Create treasury with multiple bankers
bankers := []treasury.Banker{banker1, banker2}
t, err := treasury.New(bankers)
if err != nil {
    panic(err)
}

// Send payment (tries bankers in order until one succeeds)
err = t.SendPayment("recipient_address", 1000, "ugnot")
if err != nil {
    // All bankers failed
}
```

## Banker Interface

Implement this interface to create custom payment backends:

```go
type Banker interface {
    ID() string                                           // Unique identifier
    SendCoins(to string, amount int64, denom string) error // Send payment
}
```

## Example Banker Implementation

```go
type CustomBanker struct {
    id string
}

func (cb *CustomBanker) ID() string {
    return cb.id
}

func (cb *CustomBanker) SendCoins(to string, amount int64, denom string) error {
    // Implement your payment logic here
    // Return error if payment fails
    return nil
}
```

## API

```go
type Treasury struct {
    // private fields
}

// Constructor
func New(bankers []Banker) (*Treasury, error)

// Payment operations
func (t *Treasury) SendPayment(to string, amount int64, denom string) error

// Banker management  
func (t *Treasury) AddBanker(banker Banker) error
func (t *Treasury) RemoveBanker(bankerID string) error
func (t *Treasury) ListBankers() []string
func (t *Treasury) HasBanker(bankerID string) bool
```

## Advanced Usage

```go
// Multi-backend treasury
type PaymentSystem struct {
    treasury *treasury.Treasury
}

func NewPaymentSystem() *PaymentSystem {
    // Create different payment backends
    gnoBanker := &GnoBanker{id: "gno"}
    usdcBanker := &USDCBanker{id: "usdc"}
    ethBanker := &EthBanker{id: "eth"}
    
    bankers := []treasury.Banker{gnoBanker, usdcBanker, ethBanker}
    t, err := treasury.New(bankers)
    if err != nil {
        panic(err)
    }
    
    return &PaymentSystem{treasury: t}
}

func (ps *PaymentSystem) PayUser(userAddr string, amount int64, token string) error {
    return ps.treasury.SendPayment(userAddr, amount, token)
}

// Add new payment method dynamically
func (ps *PaymentSystem) AddPaymentMethod(banker treasury.Banker) error {
    return ps.treasury.AddBanker(banker)
}
```

## Error Handling

```go
var (
    ErrNoBankerProvided  = errors.New("no banker provided")
    ErrDuplicateBanker   = errors.New("duplicate banker")
    ErrBankerNotFound    = errors.New("banker not found")
    ErrSendPaymentFailed = errors.New("failed to send payment")
)
```

## Fallback Logic

The treasury tries bankers in the order they were registered:

1. Attempt payment with first banker
2. If it fails, try next banker
3. Continue until payment succeeds or all bankers fail
4. Return error if all bankers fail

## Use Cases

- **Multi-token payments**: Support payments in different cryptocurrencies
- **Payment redundancy**: Fallback to different payment methods if primary fails
- **Cross-chain payments**: Route payments through different blockchain networks
- **Payment aggregation**: Combine multiple payment services under single interface
- **A/B testing**: Try different payment methods for optimization

## Integration Example

```go
var treasurySystem *treasury.Treasury

func init() {
    // Initialize with available payment methods
    bankers := []treasury.Banker{
        &StandardBanker{id: "standard"},
        &ExpressBanker{id: "express"},
    }
    
    var err error
    treasurySystem, err = treasury.New(bankers)
    if err != nil {
        panic(err)
    }
}

func SendReward(winner string, prize int64) error {
    return treasurySystem.SendPayment(winner, prize, "ugnot")
}
```

This package provides robust payment infrastructure with multiple backend support and automatic failover capabilities.
