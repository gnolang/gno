package staker

import (
	"gno.land/p/demo/ufmt"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/common"
)

// defaultAllowed is the list of default allowed tokens to create external incentive
var defaultAllowed = []string{consts.GNS_PATH, consts.GNOT}

// allowedTokens is a slice of all allowed token paths, including the default and added tokens.
var allowedTokens = make([]string, 0, len(defaultAllowed))

func init() {
	allowedTokens = defaultAllowed
}

// TokenValidator is a function type that validates a token
type TokenValidator func(tokenPath string) error

// TokenExecutor is a token manipulation function type
type TokenExecutor func(tokenPath string, tokens []string) []string

// modifyTokenList handles common token modification logics, such as admin check, validation, and execution.
func modifyTokenList(cur realm, tokenPath string, validator TokenValidator, executor TokenExecutor) error {
	caller := getPrevAddr()
	if err := common.AdminOnly(caller); err != nil {
		return err
	}

	// validate token operation if validator is provided
	if validator != nil {
		if err := validator(tokenPath); err != nil {
			return err
		}
	}

	allowedTokens = executor(tokenPath, allowedTokens)
	return nil
}

// addTokenExecutor executes token append operation
func addTokenExecutor(tokenPath string, tokens []string) []string {
	if contains(tokens, tokenPath) {
		return tokens
	}
	return append(tokens, tokenPath)
}

// removeTokenExecutor executes token removal operation
func removeTokenExecutor(tokenPath string, tokens []string) []string {
	// find and remove token
	for i, t := range tokens {
		if t == tokenPath {
			return append(tokens[:i], tokens[i+1:]...)
		}
	}

	// if token not found, return the original list
	return tokens
}

// removeTokenValidator validates token removal operation
func removeTokenValidator(tokenPath string) error {
	if contains(defaultAllowed, tokenPath) {
		return ufmt.Errorf("%v: can not remove default token(%s)", errDefaultExternalToken, tokenPath)
	}
	return nil
}

// AddToken adds a new token path to the list of allowed tokens
// Only the admin can add a new token.
//
// Parameters:
//   - tokenPath (string): The path of the token to add
//
// Panics:
//   - If the caller is not the admin
func AddToken(cur realm, tokenPath string) {
	if err := modifyTokenList(cur, tokenPath, nil, addTokenExecutor); err != nil {
		panic(err.Error())
	}
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
func RemoveToken(cur realm, tokenPath string) {
	if err := modifyTokenList(cur, tokenPath, removeTokenValidator, removeTokenExecutor); err != nil {
		panic(err.Error())
	}
}
