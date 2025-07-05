package net

import (
	"filetests/extern/context"
)

// XXX dummy
type Conn interface {
	Read(b []byte) (n int, err error)
}

// XXX dummy
type Resolver struct {
	PreferGo     bool
	StrictErrors bool                                                             // Go 1.9
	Dial         func(ctx context.Context, network, address string) (Conn, error) // Go 1.9
}

// XXX dummy
func (r *Resolver) LookupAddr(ctx context.Context, addr string) (names []string, err error) {
	return []string{"127.0.0.1"}, nil // XXX wrong
}

// XXX dummy
func (r *Resolver) LookupHost(ctx context.Context, host string) (addrs []string, err error) {
	return []string{"127.0.0.1"}, nil // XXX wrong
}

var DefaultResolver = &Resolver{}

type IPNet struct {
	IP   IP     // network number
	Mask IPMask // network mask
}

type IP []byte

type IPMask []byte

// XXX dummy
func ParseCIDR(s string) (IP, *IPNet, error) {
	return nil, nil, nil
}
