// Package upgradeable provides a comprehensive system for
// implementing upgradeable functions and contracts in Gno realms.
//
// It offers low-overhead abstractions that make it easy to develop
// systems with upgradeable components while maintaining proper
// access controls and type safety.
package upgradeable

// Import dependencies to ensure they're included in the package
import (
	_ "std"
	_ "gno.land/p/demo/ownable"
)

// Version information
const (
	PackageName    = "upgradeable"
	PackageVersion = "0.1.0"
)
