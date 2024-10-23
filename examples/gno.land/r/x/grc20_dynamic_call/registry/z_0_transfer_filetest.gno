// PKGPATH: gno.land/r/x/grc20_dynamic_call/registries_test
package registries_test

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
	transferTo      std.Address
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

	transferTo = testutils.TestAddress("transferTo")
}

func main() {
	// FOO
	fooBalanceRealm := registry.BalanceOfByInterface("gno.land/r/x/grc20_dynamic_call/foo", actualRealmAddr)
	println(fooBalanceRealm)

	fooBalanceTo := registry.BalanceOfByInterface("gno.land/r/x/grc20_dynamic_call/foo", transferTo)
	println(fooBalanceTo)

	registry.TransferByInterface("gno.land/r/x/grc20_dynamic_call/foo", transferTo, 12345)
	fooBalanceTo = registry.BalanceOfByInterface("gno.land/r/x/grc20_dynamic_call/foo", transferTo)
	println(fooBalanceTo)

	// BAR
	barBalanceRealm := registry.BalanceOfByInterface("gno.land/r/x/grc20_dynamic_call/bar", actualRealmAddr)
	println(barBalanceRealm)

	barBalanceTo := registry.BalanceOfByInterface("gno.land/r/x/grc20_dynamic_call/bar", transferTo)
	println(barBalanceTo)

	registry.TransferByInterface("gno.land/r/x/grc20_dynamic_call/bar", transferTo, 54321)
	barBalanceTo = registry.BalanceOfByInterface("gno.land/r/x/grc20_dynamic_call/bar", transferTo)
	println(barBalanceTo)
}

// Output:
// 10000000
// 0
// 12345
// 10000000
// 0
// 54321
