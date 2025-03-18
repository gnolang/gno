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

- **Type**: A unique identifier (e.g., `OpTypePool`, `OpTypePosition`)
- **Name**: Human-readable name
- **Description**: Detailed explanation

### Halt Levels

Halt levels define which operations are allowed at specific protocol states. Each level has:

- **ID**: A unique identifier (e.g., `LvNoHalt`, `LvContractHalt`, `LvEmergencyHalt`)
- **Name**: Human-readable name (e.g., `"NoHalt"`, `"ContractHalt"`)
- **Description**: Explanation of the halt level's purpose
- **Allowed Operations**: Map of operation types to boolean permissions

### Manager

The Manager orchestrates the halt system by:

- Maintaining a registry of operations and halt levels
- Tracking the current halt level
- Determining if operations are allowed based on the current level
- Enabling/disabling specific operations at the current halt level

## Default Halt Levels

| Level ID | Name | Description | Allowed Operations |
|----------|------|-------------|-------------------|
| 1 | `NoHalt` | Normal operation | All operations enabled by default |
| 2 | `ContractHalt` | Specific contract operations disabled | All operations enabled by default, can be selectively disabled |
| 3 | `EmergencyHalt` | Only withdrawals allowed | Only `OpTypeWithdraw` and `OpTypeGovernance` |
| 4 | `CompleteHalt` | All ops disabled | Only `OpTypeWithdraw` |

## Operation Types

The package defines several default operation types:

- `OpTypePool`: Pool management operations
- `OpTypePosition`: Position management operations
- `OpTypeProtocolFee`: Fee-related operations
- `OpTypeRouter`: Swap-related operations
- `OpTypeStaker`: Liquidity-related operations
- `OpTypeLaunchpad`: Launchpad operation
- `OpTypeGovernance`: `gov/governance` operation
- `OpTypeGovStaker`: `gov/staker` operation
- `OpTypeXGns`: `gov/xgns` contract operation
- `OpTypeCommunityPool`: Community pool operations
- `OpTypeEmission`: Emission operations
- `OpTypeWithdraw`: Withdrawal operations

## Usage Scenarios

### Scenario 1: Normal Operations

During normal functioning, all operations are permitted. The protocol operates at `LvNoHalt` level.

### Scenario 2: Contract-Specific Halt

The `LvContractHalt` level allows for fine-grained control over which specific contract operations are allowed. This enables targeted maintenance or emergency responses for individual components without affecting the entire protocol.

For example, if issues are detected in the pool contract, you can set the system to `LvContractHalt` and then disable only pool operations:

```go
// Set to contract halt level
manager.SetCurrentLevel(halt.LvContractHalt)

// Disable only pool operations
manager.SetOperationStatus(halt.OpTypePool, false)

// Other operations like positions, withdrawals, etc. remain enabled
```

### Scenario 3: Critical Vulnerability Found

If a critical vulnerability is discovered, the protocol can switch to `LvEmergencyHalt`, allowing only withdrawals and governance operations until the issue is resolved.

### Scenario 4: Complete System Upgrade

During a major system upgrade, all operations except withdrawals may need to be halted temporarily. The protocol switches to `LvCompleteHalt` until the upgrade is complete.

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

// Change the halt level for contract-specific halts
if err := manager.SetCurrentLevel(halt.LvContractHalt); err != nil {
    // Handle error
}

// Disable specific operations while in ContractHalt level
if err := manager.SetOperationStatus(halt.OpTypePool, false); err != nil {
    // Handle error
}

// Check if a specific operation is allowed
poolOp := halt.NewOperation(halt.OpTypePool, "Pool Operations", "Manage liquidity pools")
if manager.Level().IsOperationAllowed(poolOp) {
    // Perform pool operation
} else {
    return ufmt.Errorf("pool operations are currently disabled: %s", manager.Status(halt.OpTypePool))
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
        halt.OpTypePool:       false,
        halt.OpTypePosition:   false,
        halt.OpTypeWithdraw:   false,
        OpTypeStake:           true,
    },
)

// Create a manager with custom configuration
manager := halt.NewManager(
    halt.WithOperations([]halt.Operation{stakeOp}),
    halt.WithLevels([]halt.HaltLevel{customLevel}),
    halt.WithInitialLevel(LvStakeOnly),
)
```

## Managing Operation Status

The `SetOperationStatus` method allows for granular control over operations at the current halt level:

```go
// Enable a specific operation
if err := manager.SetOperationStatus(halt.OpTypePool, true); err != nil {
    // Handle error
}

// Disable a specific operation
if err := manager.SetOperationStatus(halt.OpTypePool, false); err != nil {
    // Handle error
}
```

This is particularly useful with the `LvContractHalt` level, which is designed to allow selective enabling/disabling of operations for specific contracts.

## Composite Halt Levels

For complex scenarios, you can create composite halt levels that combine multiple levels with logical operators:

```go
// Create a composite level that requires BOTH level1 AND level2 conditions to be met
compositeLevel := &halt.CompositeHaltLevel{
    baseInfo: halt.baseInfo{name: "Composite", desc: "Composite level"},
    levels:   halt.HaltLevels{level1, level2},
    operator: halt.CompositeOpAnd,
}

// Check if an operation is allowed based on composite rules
if compositeLevel.IsOperationAllowed(poolOp) {
    // Perform pool operation
}
```

## Best Practices

1. **Centralize Management**: Maintain a single halt manager instance at the protocol level
2. **Use Contract-Specific Halts**: When possible, use `LvContractHalt` with `SetOperationStatus` to minimize disruption
3. **Clear Communication**: Provide clear error messages or event messages when operations are halted
4. **Access Control**: Restrict the ability to change halt levels to authorized entities
5. **Graceful Degradation**: Design halt levels to prioritize user funds safety

## Halt Level Transition Guidelines

- **Admin-Only Access**: It is strongly recommended to restrict halt level adjustments to addresses registered as admins in your system
- **Operation Status Management**: Use `SetOperationStatus` to fine-tune which operations are allowed at the current halt level
- **Emergency Response**: Have clear procedures for who can trigger emergency halt levels
- **Gradual Recovery**: When recovering from an incident, move gradually through halt levels
- **Temporary Restrictions**: Use `LvContractHalt` with selective operation disabling for maintenance or minor issues
