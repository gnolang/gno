# `ownable` - Ownership Management

A utility package providing ownership functionality for Gno contracts and objects. Implements access control by restricting certain operations to the designated owner.

## Features

- **Single owner**: One address controls the ownable resource
- **Transfer ownership**: Owner can transfer control to another address
- **Origin caller support**: Initialize with the original transaction sender
- **Embeddable**: Can be embedded in other structs for per-object ownership
- **Events**: Emits ownership transfer events

## Usage

```go
import "gno.land/p/nt/ownable"

// Create ownable (current realm address as owner)
own := ownable.New()

// Create with origin caller as owner (use in init())
own := ownable.NewWithOrigin()

// Check if address is owner
isOwner := own.CallerIsOwner() // true if std.CurrentCaller() is owner

// Get current owner
owner := own.Owner()

// Transfer ownership (only owner can do this)
own.TransferOwnership("g1abc123...") // panics if not called by owner

// Require owner access (use in protected functions)
own.RequireOwner() // panics if caller is not owner
```

## Embedding Example

```go
type MyContract struct {
    *ownable.Ownable
    data string
}

func NewMyContract() *MyContract {
    return &MyContract{
        Ownable: ownable.New(),
        data:    "initial",
    }
}

func (c *MyContract) SetData(newData string) {
    c.RequireOwner() // Only owner can set data
    c.data = newData
}

func (c *MyContract) GetData() string {
    return c.data // Anyone can read
}
```

## API

```go
type Ownable struct {
    // private fields
}

// Constructors
func New() *Ownable
func NewWithOrigin() *Ownable

// Access control
func (o *Ownable) CallerIsOwner() bool
func (o *Ownable) RequireOwner()

// Owner management
func (o *Ownable) Owner() std.Address
func (o *Ownable) TransferOwnership(newOwner std.Address)
```

## Events

The package emits the following events:

- `OwnershipTransfer`: When ownership is transferred from one address to another

## Security Considerations

- Always call `RequireOwner()` at the beginning of protected functions
- Use `NewWithOrigin()` only in `init()` functions where you want the transaction originator as owner
- Be careful when transferring ownership - ensure the new owner address is correct
- Consider implementing two-step ownership transfer for additional security

## Sub-packages

- `ownable/exts/authorizable` - Extended authorization features

This package is fundamental for implementing access control in Gno contracts, ensuring only authorized addresses can perform administrative operations.
