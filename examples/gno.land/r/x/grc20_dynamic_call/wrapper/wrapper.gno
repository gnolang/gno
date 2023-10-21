package wrapper

import (
	"gno.land/r/demo/users"

	"gno.land/r/x/grc20_dynamic_call/bar"
	"gno.land/r/x/grc20_dynamic_call/baz"
	"gno.land/r/x/grc20_dynamic_call/foo"

	"gno.land/r/x/grc20_dynamic_call/registry"
)

type FooTokenInterface struct{}

func (FooTokenInterface) Transfer() func(to users.AddressOrName, amount uint64) {
	return foo.Transfer
}

func (FooTokenInterface) TransferFrom() func(from, to users.AddressOrName, amount uint64) {
	return foo.TransferFrom
}

func (FooTokenInterface) BalanceOf() func(owner users.AddressOrName) uint64 {
	return foo.BalanceOf
}

var _ registry.GRC20Interface = FooTokenInterface{}

type BarTokenInterface struct{}

func (BarTokenInterface) Transfer() func(to users.AddressOrName, amount uint64) {
	return bar.Transfer
}

func (BarTokenInterface) TransferFrom() func(from, to users.AddressOrName, amount uint64) {
	return bar.TransferFrom
}

func (BarTokenInterface) BalanceOf() func(owner users.AddressOrName) uint64 {
	return bar.BalanceOf
}

var _ registry.GRC20Interface = BarTokenInterface{}

type BazTokenInterface struct{}

func (BazTokenInterface) Transfer() func(to users.AddressOrName, amount uint64) {
	return baz.Transfer
}

func (BazTokenInterface) TransferFrom() func(from, to users.AddressOrName, amount uint64) {
	return baz.TransferFrom
}

func (BazTokenInterface) BalanceOf() func(owner users.AddressOrName) uint64 {
	return baz.BalanceOf
}

var _ registry.GRC20Interface = BazTokenInterface{}

func init() {
	registry.RegisterGRC20Interface("gno.land/r/x/grc20_dynamic_call/foo", FooTokenInterface{})
	registry.RegisterGRC20Interface("gno.land/r/x/grc20_dynamic_call/bar", BarTokenInterface{})
	registry.RegisterGRC20Interface("gno.land/r/x/grc20_dynamic_call/baz", BazTokenInterface{})
}
