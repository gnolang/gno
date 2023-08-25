package grc20_registry

import (
	"std"
	"testing"

	"gno.land/p/demo/grc/grc20"
)

func TestRegistry(t *testing.T) {
	coin := &dummyImpl{}
	realmAddr := std.CurrentRealm().PkgPath()
	Register(coin)
	regCoin, ok := Get(realmAddr)
	if !ok {
		t.Fatal("expected to find coin")
	}
	if coin.GetSymbol() != "TST" {
		t.Fatal("expected coin to have symbol TST")
	}
	expected := `# GRC20 Registry

## Registered Tokens

* [TestToken ()]()
`
	got := Render("")
	if got != expected {
		t.Fatalf("expected `%s`, got `%s`", expected, got)
	}

	// we test this here because there is more chance to find a bug after a token has been registered
	if _, ok := Get("0xdeadbeef"); ok {
		t.Fatal("expected not to find coin")
	}
}

type dummyImpl struct{}

// FIXME: this should fail.
var _ grc20.IGRC20 = (*dummyImpl)(nil)

func (impl *dummyImpl) GetName() string                               { return "TestToken" }
func (impl *dummyImpl) GetSymbol() string                             { return "TST" }
func (impl *dummyImpl) GetDecimals() uint                             { panic("not implemented") }
func (impl *dummyImpl) TotalSupply() uint64                           { panic("not implemented") }
func (impl *dummyImpl) BalanceOf(account std.Address) (uint64, error) { panic("not implemented") }
func (impl *dummyImpl) Transfer(to std.Address, amount uint64) error  { panic("not implemented") }
func (impl *dummyImpl) Allowance(owner, spender std.Address) (uint64, error) {
	panic("not implemented")
}
func (impl *dummyImpl) Approve(spender std.Address, amount uint64) error { panic("not implemented") }
func (impl *dummyImpl) TransferFrom(from, to std.Address, amount uint64) error {
	panic("not implemented")
}
