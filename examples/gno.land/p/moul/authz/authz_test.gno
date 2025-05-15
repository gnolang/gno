package authz_test

import (
	"std"
)

// authorityMock is a simple mock implementation of Authority for testing.
type authorityMock struct {
	canFunc func(addr std.Address, action string) bool
}

func (m *authorityMock) Can(addr std.Address, action string) bool {
	return m.canFunc(addr, action)
}

func (m *authorityMock) String() string {
	return "mock authority"
}

// newAuthorityMock creates a new authorityMock with the given can function.
func newAuthorityMock(canFunc func(addr std.Address, action string) bool) *authorityMock {
	return &authorityMock{canFunc: canFunc}
}
