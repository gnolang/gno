package main

import (
	"std"

	"gno.land/p/demo/testutils"

	"gno.land/r/x/grc20_dynamic_call/registry"

	"gno.land/r/x/grc20_dynamic_call/bar"
	"gno.land/r/x/grc20_dynamic_call/foo"

	"gno.land/r/demo/users"
)

var (
	actualRealmAddr std.Address

	transferFromFrom std.Address = testutils.TestAddress("transferFromFrom")
	transferFromTo   std.Address = testutils.TestAddress("transferFromTo")
)

// FOO
type FooStruct struct{}

func (FooStruct) Transfer() func(to users.AddressOrName, amount uint64) {
	return foo.Transfer
}
func (FooStruct) TransferFrom() func(from, to users.AddressOrName, amount uint64) {
	return foo.TransferFrom
}
func (FooStruct) BalanceOf() func(owner users.AddressOrName) uint64 {
	return foo.BalanceOf
}

// BAR
type BarStruct struct{}

func (BarStruct) Transfer() func(to users.AddressOrName, amount uint64) {
	return bar.Transfer
}
func (BarStruct) TransferFrom() func(from, to users.AddressOrName, amount uint64) {
	return bar.TransferFrom
}
func (BarStruct) BalanceOf() func(owner users.AddressOrName) uint64 {
	return bar.BalanceOf
}

func init() {
	registry.RegisterGRC20Interface("gno.land/r/x/grc20_dynamic_call/foo", FooStruct{})
	registry.RegisterGRC20Interface("gno.land/r/x/grc20_dynamic_call/bar", BarStruct{})

	actualRealmAddr = std.DerivePkgAddr("gno.land/r/x/grc20_dynamic_call/registry")
}

func main() {
	foo.Faucet()
	foo.Approve(users.AddressOrName(std.DerivePkgAddr("gno.land/r/x/grc20_dynamic_call/registry")), 1000*10000)

	fooBalanceOrigCaller := registry.BalanceOfByInterface("gno.land/r/x/grc20_dynamic_call/foo", std.GetOrigCaller())
	println(fooBalanceOrigCaller)

	fooBalanceTo := registry.BalanceOfByInterface("gno.land/r/x/grc20_dynamic_call/foo", transferFromTo)
	println(fooBalanceTo)

	registry.TransferFromByInterface("gno.land/r/x/grc20_dynamic_call/foo", std.GetOrigCaller(), transferFromTo, 54321)

	fooBalanceOrigCaller = registry.BalanceOfByInterface("gno.land/r/x/grc20_dynamic_call/foo", std.GetOrigCaller())
	println(fooBalanceOrigCaller)

	fooBalanceTo = registry.BalanceOfByInterface("gno.land/r/x/grc20_dynamic_call/foo", transferFromTo)
	println(fooBalanceTo)
}

// Output:
// 10000000
// 0
// 9945679
// 54321
