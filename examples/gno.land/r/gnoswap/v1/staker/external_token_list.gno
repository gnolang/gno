package staker

import (
	"std"

	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"
)

// defaultAllowed is the list of default allowed tokens to create external incentive
var defaultAllowed = []string{consts.GNS_PATH, consts.GNOT}

// allowedTokens is a slice of all allowed token paths, including the default and added tokens.
var allowedTokens = []string{}

func init() {
	allowedTokens = defaultAllowed
}

// AddToken adds a new token path to the list of allowed tokens
// Only the admin can add a new token.
//
// Parameters:
//   - tokenPath (string): The path of the token to add
//
// Panics:
//   - If the caller is not the admin
func AddToken(tokenPath string) {
	caller := getPrevAddr()
	if err := common.AdminOnly(caller); err != nil {
		panic(err.Error())
	}

	// if exist just return
	for _, t := range allowedTokens {
		if t == tokenPath {
			return
		}
	}

	allowedTokens = append(allowedTokens, tokenPath)
}

// RemoveToken removes a token path from the list of allowed tokens.
// Only the admin  can remove a token.
//
// Default tokens can not be removed.
//
// Parameters:
//   - tokenPath (string): The path of the token to remove
//
// Panics:
//   - If the caller is not the admin
func RemoveToken(tokenPath string) {
	caller := std.PrevRealm().Addr()
	if err := common.AdminOnly(caller); err != nil {
		panic(err.Error())
	}

	// if default token, can not remove
	isDefault := contains(defaultAllowed, tokenPath)
	if isDefault {
		panic(addDetailToError(
			errDefaultExternalToken,
			ufmt.Sprintf("external_token_list.gno__RemoveToken() || can not remove default token(%s)", tokenPath),
		))
	}

	for i, t := range allowedTokens {
		if t == tokenPath {
			allowedTokens = append(allowedTokens[:i], allowedTokens[i+1:]...)
			return
		}
	}
}
