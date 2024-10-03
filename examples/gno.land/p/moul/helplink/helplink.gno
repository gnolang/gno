/*
Package helplink provides utilities for crafting help page links that work
with Gnoweb, Gnobro, and other clients supporting the flavored markdown of
Gno contracts.

This package is designed to facilitate the creation of dynamic help links,
making it easier for users to access relevant documentation within Gno's
ecosystem.

The top-level Func, FuncURL, and Home functions are for the "relative realm";
specifying a custom Realm allows using a full realm path for the current one
or another one.
*/
package helplink

import "strings"

const chainDomain = "gno.land" // XXX: std.ChainDomain

// Func returns a markdown link for the specific function with optional
// key-value arguments, for the current realm.
func Func(title string, fn string, args ...string) string {
	return Realm("").Func(title, fn, args...)
}

// FuncURL returns a URL for the specified function with optional key-value
// arguments, for the current realm.
func FuncURL(fn string, args ...string) string {
	return Realm("").FuncURL(fn, args...)
}

// Home returns the URL for the help homepage of the current realm.
func Home() string {
	return Realm("").Home()
}

// Realm represents a specific realm for generating help links.
type Realm string

// prefix returns the URL prefix for the realm.
func (r Realm) prefix() string {
	// relative
	if r == "" {
		return ""
	}

	// local realm -> /realm
	realm := string(r)
	if strings.Contains(realm, chainDomain) {
		return strings.TrimPrefix(realm, chainDomain)
	}

	// remote realm -> https://remote.land/realm
	return "https://" + string(r)
}

// Func returns a markdown link for the specified function with optional
// key-value arguments.
func (r Realm) Func(title string, fn string, args ...string) string {
	// XXX: escape title
	return "[" + title + "](" + r.FuncURL(fn, args...) + ")"
}

// FuncURL returns a URL for the specified function with optional key-value
// arguments.
func (r Realm) FuncURL(fn string, args ...string) string {
	// Start with the base query
	url := r.prefix() + "?help&__func=" + fn

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

// Home returns the base help URL for the specified realm.
func (r Realm) Home() string {
	return r.prefix() + "?help"
}
