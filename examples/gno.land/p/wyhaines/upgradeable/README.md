# Upgradeable

A Gno.land package for implementing upgradeable functions and contracts.

## Overview

This package provides a system for implementing upgradeable functions in Gno.land realms, with proper access controls and minimal boilerplate. It solves common issues with ad-hoc upgradeability patterns:

1. **Access Control**: Only authorized users can upgrade functions
2. **Reduced Type Casting**: Specialized holders for common function types
3. **Low Boilerplate**: Centralized system reduces repetitive code
4. **Flexibility**: Works with any function type

## Usage

### Basic Usage

```go
package mywebsite

import (
    "std"
    "gno.land/p/demo/upgradeable"
)

// Global registry and function holder
var (
    registry = upgradeable.New()
    renderFunc = upgradeable.NewStringFuncHolder(
        registry,
        "render",
        func(path string) string {
            return "Initial version"
        },
    )
)

// Public function that uses the upgradeable implementation
func Render(path string) string {
    fn := renderFunc.Get()
    return fn(path)
}

// Admin function to upgrade the implementation
func UpgradeRender(newRender func(string) string) error {
    // Only owner can call successfully
    return renderFunc.Update(newRender)
}
```

## Components

1. **Registry**: Central manager for upgradeable functions
2. **FunctionHolder**: Base wrapper for function types
3. **StringFuncHolder**, **AddressBoolFuncHolder**, etc.: Type-specific function holders
4. **ContractProxy**: Implementation of the proxy pattern

## Security

- Uses `ownable` for access control
- Only owners can register/upgrade functions
- Emits events for all upgrades for transparency

## License

Apache 2.0
