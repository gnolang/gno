# GnoSwap Halt Package

`gno.land/r/gnoswap/v1/halt`

## Overview

This package uses the core halt framework (`gno.land/p/gnoswap/halt`) and adapts it for specific needs.

## For Mainnet

The protocol is initially deployed in `MainnetSafeMode`, which disables most operations for safety during the right after mainnet launch. Operations will be gradually enabled as stability is confirmed.

## Features

- **Granular Operation Control**: Individual operations can be enabled/disabled
- **Multiple Halt Levels**: Supports various protocol states from fully operational to completely halted
- **Admin & Governance Management**: Controlled by authorized addresses only

## Halt Levels

| Level | Description | Initial State |
|-------|-------------|--------------|
| **`MainnetSafeMode`** | Special mode for beta mainnet with governance-only operations | ✓ Current |

## Usage

### Checking if Operations are Allowed

```go
import "gno.land/r/gnoswap/v1/halt"

func YourFunction() {
    // Check if pool contract are allowed
    if err := halt.IsHalted(halt.OpTypePool); err != nil {
        panic(err)
    }

    // Or, you cancheck multiple operations at once
    if err := halt.IsHalted(halt.OpTypePool, halt.OpTypeRouter); err != nil {
        panic(err)
    }
}
```

### Admin Operations

These functions can only be called by admin addresses. Mostly these functions are named with `ByAdmin` suffix:

#### Setting Halt Level

| Function | Description |
|----------|-------------|
| `halt.SetHaltLevel(haltLevel halt.LevelID)` | Sets the halt level |

#### Legacy Support Functions

| Function | Description |
|----------|-------------|
| `halt.SetHalt(halt bool)` | Active/Deactivate halt state |

#### Specific Operation Settings

| Function | Description |
|----------|-------------|
| `halt.SetOperationStatus(halt.OpTypeSwap, false)` | Sets the activation status for a specific operation (e.g., Swap) |

```go
// Usage examples
halt.SetHaltLevel(halt.LvEmergencyHalt)
halt.SetHalt(true)  // Complete halt
halt.SetHalt(false) // No halt
halt.SetOperationStatus(halt.OpTypeSwap, false)
```

### Governance Operations

These functions can only be called by the governance contract:

| Function | Description | Access |
|----------|-------------|---------|
| `halt.SetHaltLevel(haltLevel halt.LevelID)` | Sets the halt level | Governance Only |
| `halt.SetHalt(halt bool)` | Deactivates halt state (Legacy) | Governance Only |
| `halt.SetOperationStatus(opType halt.OpType, allowed bool)` | Configures specific operation status (e.g., Swap) | Governance Only |

```go
// Usage examples
halt.SetHaltLevel(halt.LvEmergencyHalt)
halt.SetHalt(true)  // Complete halt
halt.SetHalt(false) // No halt
halt.SetOperationStatus(halt.OpTypeSwap, false)
```

## Beta Mainnet Deployment

For the beta mainnet deployment, operations will be enabled in this sequence:

1. Deploy with `MainnetSafeMode` (current state)
2. Enable withdrawals after stability confirmation
3. Enable swaps after further testing
4. Enable liquidity operations
5. Eventually transition to `NoHalt` for full operation

## References

- Core halt protocol: `gno.land/p/gnoswap/halt`
- Access control: `gno.land/r/gnoswap/v1/access`
- Issue tracking MainnetSafeMode: [GitHub #517](https://github.com/gnoswap-labs/gnoswap/issues/517)
