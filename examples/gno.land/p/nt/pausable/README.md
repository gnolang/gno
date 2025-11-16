# `pausable` - Pausable Functionality

A utility package that provides pause/unpause functionality for contracts and operations. Built on top of the ownable pattern, allowing owners to temporarily halt operations when needed.

## Features

- **Owner-controlled**: Only the owner can pause/unpause
- **State checking**: Query if the system is currently paused
- **Protection mechanism**: Easily restrict operations during paused state
- **Ownable integration**: Works seamlessly with existing ownable contracts

## Usage

```go
import (
    "gno.land/p/nt/pausable"
    "gno.land/p/nt/ownable"
)

// Create pausable with new ownable
own := ownable.New()
p := pausable.NewFromOwnable(own)

// Check pause state
if p.IsPaused() {
    // Handle paused state
}

// Pause operations (only owner)
p.Pause() // panics if not called by owner

// Unpause operations (only owner)  
p.Unpause() // panics if not called by owner

// Require not paused (use in protected functions)
p.RequireNotPaused() // panics if currently paused
```

## Contract Integration Example

```go
type MyContract struct {
    *ownable.Ownable
    *pausable.Pausable
    data []string
}

func NewMyContract() *MyContract {
    own := ownable.New()
    return &MyContract{
        Ownable:  own,
        Pausable: pausable.NewFromOwnable(own),
        data:     make([]string, 0),
    }
}

func (c *MyContract) AddData(item string) {
    c.RequireNotPaused() // Prevent adding data when paused
    c.data = append(c.data, item)
}

func (c *MyContract) GetData() []string {
    // Reading is allowed even when paused
    return c.data
}

func (c *MyContract) EmergencyPause() {
    c.Pause() // Only owner can call this
}

func (c *MyContract) Resume() {
    c.Unpause() // Only owner can call this
}
```

## API

```go
type Pausable struct {
    // private fields
}

// Constructor
func NewFromOwnable(ownable *ownable.Ownable) *Pausable

// State management (owner only)
func (p *Pausable) Pause()
func (p *Pausable) Unpause()

// State checking
func (p *Pausable) IsPaused() bool
func (p *Pausable) RequireNotPaused()
```

## Error Handling

```go
var ErrPaused = errors.New("pausable: realm is currently paused")
```

The `RequireNotPaused()` function will panic with `ErrPaused` if called when the contract is paused.

## Use Cases

- **Emergency stops**: Quickly halt all operations during security incidents
- **Maintenance mode**: Pause user operations during upgrades or fixes
- **Rate limiting**: Temporarily pause high-frequency operations
- **Circuit breaker**: Automatically pause when certain conditions are met

## Security Considerations

- Only use pause functionality for legitimate operational needs
- Document clearly which operations are affected by pause state
- Consider having read operations unaffected by pause state
- Ensure proper access control - only owners should pause/unpause
- Test pause/unpause functionality thoroughly

This package is crucial for implementing emergency controls and operational safety measures in Gno contracts.
