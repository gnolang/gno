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
	"std"
	"strings"
)

const chainDomain = "gno.land" // XXX: std.ChainDomain (#2911)

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
	// Start with the base query
	url := r.prefix() + "$help&func=" + fn

	// Check if args length is even
	if len(args)%2 != 0 {
		// If not even, we can choose to handle the error here.
		// For example, we can just return the URL without appending
		// more args.
		return url
	}

	// Append key-value pairs to the URL
	for i := 0; i < len(args); i += 2 {
		key := args[i]
		value := args[i+1]
		// XXX: escape keys and args
		url += "&" + key + "=" + value
	}

	return url
}
