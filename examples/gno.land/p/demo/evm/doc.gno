// Package evm implements a minimalist Ethereum Virtual Machine (EVM) interpreter in Gno.
//
// The EVM is a stack-based virtual machine that executes smart contracts on the Ethereum network.
// It operates using a defined set of opcodes that perform low-level operations such as arithmetic,
// storage access, control flow, and message calls.
//
// Usage:
// To execute EVM bytecode, use the Execute function and provide the bytecode as a byte slice.
//
//	code := []byte{0x60, 0x0a, 0x60, 0x0b, 0x01} // PUSH1 0x0a, PUSH1 0x0b, ADD
//	result, err := evm.Execute(code)
//
// Note:
// This implementation is minimalist and primarily intended for educational or demonstration purposes.
// It may not support the full Ethereum specification or handle all edge cases and gas rules.
//
// Reference:
// For more information on the EVM and its opcodes, refer to the Ethereum Yellow Paper: https://ethervm.io/
package evm // import "gno.land/p/demo/evm"
