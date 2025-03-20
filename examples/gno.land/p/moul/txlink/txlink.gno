// Package txlink provides utilities for creating transaction-related links
// compatible with Gnoweb, Gnobro, and other clients within the Gno ecosystem.
//
// This package is optimized for generating lightweight transaction links with
// flexible arguments, allowing users to build dynamic links that integrate
// seamlessly with various Gno clients.
//
// The primary function, Call, is designed to produce markdown links for
// transaction functions in the current "relative realm". By specifying a custom
// Realm, you can generate links that either use the current realm path or a
// fully qualified path for another realm.
//
// This package is a streamlined alternative to helplink, providing similar
// functionality for transaction links without the full feature set of helplink.
package txlink

import (
	"net/url"
	"std"
	"strings"
)

const chainDomain = std.ChainDomain()

// Call returns a URL for the specified function with optional key-value
// arguments, for the current realm.
func Call(fn string, args ...string) string {
	return Realm("").Call(fn, args...)
}

// Realm represents a specific realm for generating tx links.
type Realm string

// prefix returns the URL prefix for the realm.
func (r Realm) prefix() string {
	// relative
	if r == "" {
		curPath := std.CurrentRealm().PkgPath()
		return strings.TrimPrefix(curPath, chainDomain)
	}

	// local realm -> /realm
	realm := string(r)
	if strings.Contains(realm, chainDomain) {
		return strings.TrimPrefix(realm, chainDomain)
	}

	// remote realm -> https://remote.land/realm
	return "https://" + string(r)
}

// Call returns a URL for the specified function with optional key-value
// arguments.
func (r Realm) Call(fn string, args ...string) string {
	if len(args) == 0 {
		return r.prefix() + "$help&func=" + fn
	}

	// Create url.Values to properly encode parameters.
	// But manage &func=fn as a special case to keep it as the first argument.
	values := url.Values{}

	// Check if args length is even
	if len(args)%2 != 0 {
		values.Add("error", "odd number of arguments")
	} else {
		// Add key-value pairs to values
		for i := 0; i < len(args); i += 2 {
			key := args[i]
			value := args[i+1]
			values.Add(key, value)
		}
	}

	// Build the base URL and append encoded query parameters
	return r.prefix() + "$help&func=" + fn + "&" + values.Encode()
}
