package grc20_registry

import (
	"std"
	"testing"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/urequire"
)

func TestRegistry(t *testing.T) {
	std.TestSetRealm(std.NewCodeRealm("gno.land/r/demo/foo"))
	realmAddr := std.CurrentRealm().PkgPath()
	tokenImpl := &dummyImpl{}

	// register
	Register(tokenImpl, "")
	regToken := Get(realmAddr)
	urequire.True(t, regToken != nil, "expected to find a token") // fixme: use urequire.NotNil
	urequire.Equal(t, regToken.GetSymbol(), "TST")

	expected := `- **TestToken** - [gno.land/r/demo/foo](/r/demo/foo)
`
	got := Render("")
	urequire.Equal(t, expected, got)

	// 404
	invalidToken := Get("0xdeadbeef")
	urequire.True(t, invalidToken == nil)

	// register with a slug
	Register(tokenImpl, "mySlug")
	regToken = Get(realmAddr + ".mySlug")
	urequire.True(t, regToken != nil, "expected to find a token") // fixme: use urequire.NotNil
	urequire.Equal(t, regToken.GetSymbol(), "TST")

	// override
	Register(tokenImpl, "")
	regToken = Get(realmAddr + "")
	urequire.True(t, regToken != nil, "expected to find a token") // fixme: use urequire.NotNil
	urequire.Equal(t, regToken.GetSymbol(), "TST")

	expected = `- **TestToken** - [gno.land/r/demo/foo](/r/demo/foo)
- **TestToken** - [gno.land/r/demo/foo](/r/demo/foo).mySlug
`
	got = Render("")
	urequire.Equal(t, expected, got)
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
