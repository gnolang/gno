# Halt Package

## Overview

The **Halt** package provides a flexible and extensible framework for managing protocol operation permissions based on different halt levels. It allows protocols to control which operations are permitted during various states, enabling graceful degradation during emergencies or maintenance periods.

To use this package, add the following import path:

```plain
"gno.land/p/gnoswap/halt"
```

## Key Concepts

### Operations

Operations represent actions that can be performed within a protocol. Each operation has:

- **Type**: A unique identifier (e.g., `OpTypeSwap`, `OpTypeLiquidity`)
- **Name**: Human-readable name
- **Description**: Detailed explanation

### Halt Levels

Halt levels define which operations are allowed at specific protocol states. Each level has:

- **ID**: A unique identifier (e.g., `LvNoHalt`, `LvEmergencyHalt`)
- **Name**: Human-readable name (e.g., `"NoHalt"`, `"EmergencyHalt"`)
- **Description**: Explanation of the halt level's purpose
- **Allowed Operations**: Map of operation types to boolean permissions

### Manager

The Manager orchestrates the halt system by:

- Maintaining a registry of operations and halt levels
- Tracking the current halt level
- Determining if operations are allowed based on the current level

## Default Halt Levels

| Level ID | Name | Description | Swap | Liquidity | Withdraw |
|----------|------|-------------|------|-----------|----------|
| 1 | `NoHalt` | Normal operation | ✅ | ✅ | ✅ |
| 2 | `SwapHalt` | Swaps disabled | ❌ | ✅ | ✅ |
| 3 | `LiquidityHalt` | No swaps, no liquidity ops | ❌ | ❌ | ✅ |
| 4 | `EmergencyHalt` | Only withdrawals allowed | ❌ | ❌ | ✅ |
| 5 | `CompleteHalt` | All ops disabled | ❌ | ❌ | ❌ |

## Usage Scenarios

### Scenario 1: Normal Operations

During normal functioning, all operations are permitted. The protocol operates at `LvNoHalt` level.

### Scenario 2: High Market Volatility

During periods of extreme volatility, swaps might be temporarily disabled to prevent exploits or unfair trades. The protocol switches to `LvSwapHalt` level, allowing only liquidity and withdrawal operations.

### Scenario 3: Critical Vulnerability Found

If a critical vulnerability is discovered in the liquidity provision logic, the protocol can switch to `LvEmergencyHalt`, allowing only withdrawals until the issue is resolved.

### Scenario 4: Complete System Upgrade

During a major system upgrade, all operations may need to be halted temporarily. The protocol switches to `LvCompleteHalt` until the upgrade is complete.

## Basic Usage

```go
import (
    "gno.land/p/demo/ufmt"

    "gno.land/p/gnoswap/halt"
)

// Create a default manager with standard operations and halt levels
manager := halt.DefaultManager()

// Check the current halt level
currentLevel := manager.Level()
println("Current level:", currentLevel.Name()) // "NoHalt"

// Change the halt level during an emergency
if err := manager.SetCurrentLevel(halt.LvEmergencyHalt); err != nil {
    // ...
}

// Check if a specific operation is allowed
swapOp := halt.NewOperation(halt.OpTypeSwap, "Token Swap", "Swap tokens")
if manager.Level().IsOperationAllowed(swapOp) {
    // ...
} else {
    return ufmt.Errorf("swap operations are currently disabled: %s", manager.Status(halt.OpTypeSwap))
}
```

## Custom Operations and Levels

You can extend the system with custom operations and halt levels:

> **Note**: When adding custom operation types or level IDs, it's recommended to define them as _**named constants**_ rather than using string or numeric literals directly.

```go
import (
    "gno.land/p/gnoswap/halt"
)

// Define custom operation types
const (
    OpTypeStake   halt.OpType = "stake"
    OpTypeUnstake halt.OpType = "unstake"
)

// Define custom level IDs
const (
    LvStakeOnly   halt.LevelID = 10
    LvMaintenance halt.LevelID = 11
)

// Define a custom operation
stakeOp := halt.NewOperation(OpTypeStake, "Stake Tokens", "Stake tokens for rewards")

// Define a custom halt level that allows only staking
customLevel := halt.NewHaltLevel(
    LvStakeOnly, 
    "StakeOnly", 
    "Only staking operations allowed",
    map[halt.OpType]bool{
        halt.OpTypeSwap:      false,
        halt.OpTypeLiquidity: false,
        halt.OpTypeWithdraw:  false,
        OpTypeStake:          true,
    },
)

// Create a manager with custom configuration
manager := halt.NewManager(
    halt.WithOperations([]halt.Operation{stakeOp}),
    halt.WithLevels([]halt.HaltLevel{customLevel}),
    halt.WithInitialLevel(LvStakeOnly),
)
```

## Composite Halt Levels

For complex scenarios, you can create composite halt levels that combine multiple levels with logical operators:

```go
// Create a composite level that requires BOTH level1 AND level2 conditions to be met
compositeLevel := &halt.CompositeHaltLevel{
    Levels:   halt.HaltLevels{level1, level2},
    Operator: halt.CompositeOpAnd,
}

// Check if an operation is allowed based on composite rules
if compositeLevel.IsOperationAllowed(swapOp) {
    // ...
}
```

## Best Practices

1. **Centralize Management**: Maintain a single halt manager instance at the protocol level
2. **Clear Communication**: Provide clear error messages or event messages when operations are halted
3. **Access Control**: Restrict the ability to change halt levels to authorized entities
4. **Graceful Degradation**: Design halt levels to prioritize user funds safety

## Halt Level Transition Guidelines

- **Admin-Only Access**: It is strongly recommended to restrict halt level adjustments to addresses registered as admins in your system
- **Emergency Response**: Have clear procedures for who can trigger emergency halt levels
- **Gradual Recovery**: When recovering from an incident, move gradually through halt levels
- **Temporary Restrictions**: Use intermediate halt levels for maintenance or minor issues
