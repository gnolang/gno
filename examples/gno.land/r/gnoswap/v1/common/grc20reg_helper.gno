package common

import (
	"regexp"
	"std"
	"strings"

	"gno.land/p/demo/grc/grc20"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/grc20reg"
)

var re = regexp.MustCompile(`\[gno\.land/r/[^\]]+\]`)

// GetToken returns a grc20.Token instance
// if token is not registered, it will panic
// token instance supports following methods:
// - GetName
// - GetSymbol
// - GetDecimals
// - TotalSupply
// - KnownAccounts
// - BalanceOf
// - Allowance
// - RenderHome
func GetToken(path string) *grc20.Token {
	tokenGetter := grc20reg.MustGet(path) // if token is not registered, it will panic

	return tokenGetter()
}

// GetTokenTeller returns a grc20.Teller instance
// if token is not registered, it will panic
// teller instance supports following methods:
// - Transfer
// - Approve
// - TransferFrom
func GetTokenTeller(path string) grc20.Teller {
	tokenGetter := grc20reg.MustGet(path) // if token is not registered, it will panic
	token := tokenGetter()
	return token.CallerTeller()
}

// IsRegistered returns nil if token is registered to grc20reg
// otherwise, it returns an error
func IsRegistered(path string) error {
	getter := grc20reg.Get(path)
	if getter == nil {
		return ufmt.Errorf("token(%s) is not registered to grc20reg", path)
	}
	return nil
}

// MustRegistered is a helper function to check if token is registered to grc20reg
// if token is not registered, it will panic
func MustRegistered(path string) {
	if err := IsRegistered(path); err != nil {
		panic(newErrorWithDetail(
			errNotRegistered,
			ufmt.Sprintf("token(%s)", path),
		))
	}
}

// ListRegisteredTokens returns the list of registered tokens
// NOTE:
// - Unfortunately, grc20reg doesn't support this.
// - We need to parse the rendered grc20reg page to get the list of registered tokens.
func ListRegisteredTokens() []string {
	render := grc20reg.Render("")
	return extractTokenPathsFromRender(render)
}

// extractTokenPathsFromRender extracts the token paths from the rendered grc20reg realm.
// It returns the list of token paths.
func extractTokenPathsFromRender(render string) []string {
	matches := re.FindAllString(render, -1)

	tokenPaths := make([]string, 0, len(matches))
	for _, match := range matches {
		tokenPath := strings.Trim(match, "[]") // Remove the brackets
		tokenPaths = append(tokenPaths, tokenPath)
	}
	return tokenPaths
}

// TotalSupply returns the total supply of the token
func TotalSupply(path string) uint64 {
	return GetToken(path).TotalSupply()
}

// BalanceOf returns the balance of the token for the given address
func BalanceOf(path string, addr std.Address) uint64 {
	return GetToken(path).BalanceOf(addr)
}

// Allowance returns the allowance of the token for the given owner and spender
func Allowance(path string, owner, spender std.Address) uint64 {
	return GetToken(path).Allowance(owner, spender)
}
