package upgradeable

import (
	"errors"
	"std"
)

// ContractProxy is a proxy that delegates calls to an implementation contract
// It stores data but delegates logic to the implementation
type ContractProxy struct {
	owners         map[std.Address]bool // Owner addresses
	implementation string               // Path to the implementation realm
	data           Any                  // Contract state
}

// NewContractProxy creates a new proxy with the caller as owner
func NewContractProxy() *ContractProxy {
	return &ContractProxy{
		owners: map[std.Address]bool{
			std.OriginCaller(): true,
		},
	}
}

// IsOwner checks if an address is an owner
func (p *ContractProxy) IsOwner(addr std.Address) bool {
	return p.owners[addr]
}

// CallerIsOwner checks if the caller is an owner
func (p *ContractProxy) CallerIsOwner() bool {
	return p.IsOwner(std.OriginCaller())
}

// AddOwner adds a new owner (only callable by an existing owner)
func (p *ContractProxy) AddOwner(addr std.Address) error {
	if !p.CallerIsOwner() {
		return ErrCallerNotAdmin
	}

	p.owners[addr] = true
	return nil
}

// RemoveOwner removes an owner (only callable by an existing owner)
func (p *ContractProxy) RemoveOwner(addr std.Address) error {
	if !p.CallerIsOwner() {
		return ErrCallerNotAdmin
	}

	// Prevent removing the last owner
	if len(p.owners) <= 1 {
		return ErrCannotRemoveLastOwner
	}

	delete(p.owners, addr)
	return nil
}

// SetImplementation updates the implementation realm path
// Only an owner can update the implementation
func (p *ContractProxy) SetImplementation(implementationPath string) error {
	if !p.CallerIsOwner() {
		return ErrCallerNotAdmin
	}

	// Check if the implementation exists
	// Note: In a real implementation, we would validate that the realm exists
	// and implements the required interface

	oldImpl := p.implementation
	p.implementation = implementationPath

	std.Emit(
		EventImplementationUpgraded,
		"from", oldImpl,
		"to", implementationPath,
	)

	return nil
}

// Implementation returns the current implementation path
func (p *ContractProxy) Implementation() string {
	return p.implementation
}

// GetState returns the contract state
func (p *ContractProxy) GetState() Any {
	return p.data
}

// SetState updates the contract state
// Only the implementation can update the state
func (p *ContractProxy) SetState(newState Any) error {
	// In a real implementation, we would check that the caller
	// is the implementation contract

	p.data = newState
	return nil
}

// DelegateCall helper
func (p *ContractProxy) DelegateCall(method string, args ...Any) (Any, error) {
	if p.implementation == "" {
		return nil, ErrImplementationNotSet
	}

	// In a real implementation, this would use Gno's inter-realm call mechanism
	// to invoke the method on the implementation contract
	// and pass along the proxy's state

	return nil, errors.New("delegate call not implemented")
}
