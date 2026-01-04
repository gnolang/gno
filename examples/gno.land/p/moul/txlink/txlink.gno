// Package txlink provides utilities for creating transaction-related links
// compatible with Gnoweb, Gnobro, and other clients within the Gno ecosystem.
//
// This package is optimized for generating lightweight transaction links with
// flexible arguments, allowing users to build dynamic links that integrate
// seamlessly with various Gno clients.
//
// The package offers a way to generate clickable transaction MD links
// for the current "relative realm":
//
//  Using a builder pattern for more structured URLs:
//     txlink.NewLink("MyFunc").
//         AddArgs("k1", "v1", "k2", "v2"). // or multiple at once
//         SetSend("1000000ugnot").
//         URL()
//
// The builder pattern (TxBuilder) provides a fluent interface for constructing
// transaction URLs in the current "relative realm". Like Call, it supports both
// local realm paths and fully qualified paths through the underlying Call
// implementation.
//
// The Call function remains the core implementation, used both directly and
// internally by the builder pattern to generate the final URLs.
//
// This package is a streamlined alternative to helplink, providing similar
// functionality for transaction links without the full feature set of helplink.

package txlink

import (
	"chain/runtime"
	"net/url"
	"strings"
)

var chainDomain = runtime.ChainDomain()

// Realm represents a specific realm for generating tx links.
type Realm string

// TxBuilder provides a fluent interface for building transaction URLs
type TxBuilder struct {
	fn        string   // function name
	args      []string // key-value pairs
	send      string   // optional send amount
	realm_XXX Realm    // realm for the URL
}

// NewLink creates a transaction link builder for the specified function in the current realm.
func NewLink(fn string) *TxBuilder {
	return Realm("").NewLink(fn)
}

// NewLink creates a transaction link builder for the specified function in this realm.
func (r Realm) NewLink(fn string) *TxBuilder {
	if fn == "" {
		return nil
	}
	return &TxBuilder{fn: fn, realm_XXX: r}
}

// addArg adds a key-value argument pair. Returns the builder for chaining.
func (b *TxBuilder) addArg(key, value string) *TxBuilder {
	if b == nil {
		return nil
	}
	if key == "" {
		return b
	}

	// Special case: "." prefix is for reserved keywords.
	if strings.HasPrefix(key, ".") {
		panic("invalid key")
	}

	b.args = append(b.args, key, value)
	return b
}

// AddArgs adds multiple key-value pairs at once. Arguments should be provided
// as pairs: AddArgs("key1", "value1", "key2", "value2").
func (b *TxBuilder) AddArgs(args ...string) *TxBuilder {
	if b == nil {
		return nil
	}
	if len(args)%2 != 0 {
		panic("odd number of arguments")
	}
	// Add key-value pairs
	for i := 0; i < len(args); i += 2 {
		key := args[i]
		value := args[i+1]
		b.addArg(key, value)
	}
	return b
}

// SetSend adds a send amount. (Only one send amount can be specified.)
func (b *TxBuilder) SetSend(amount string) *TxBuilder {
	if b == nil {
		return nil
	}
	if amount == "" {
		return b
	}
	b.send = amount
	return b
}

// URL generates the final URL using the standard $help&func=name format.
func (b *TxBuilder) URL() string {
	if b == nil || b.fn == "" {
		return ""
	}
	args := b.args
	if b.send != "" {
		args = append(args, ".send", b.send)
	}
	return b.realm_XXX.Call(b.fn, args...)
}

// Call returns a URL for the specified function with optional key-value
// arguments, for the current realm.
func Call(fn string, args ...string) string {
	return Realm("").Call(fn, args...)
}

// prefix returns the URL prefix for the realm.
func (r Realm) prefix() string {
	// relative
	if r == "" {
		curPath := runtime.CurrentRealm().PkgPath()
		return strings.TrimPrefix(curPath, chainDomain)
	}

	// local realm -> /realm
	rlm := string(r)
	if strings.HasPrefix(rlm, chainDomain) {
		return strings.TrimPrefix(rlm, chainDomain)
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
		panic("odd number of arguments")
	}
	// Add key-value pairs to values
	for i := 0; i < len(args); i += 2 {
		key := args[i]
		value := args[i+1]
		values.Add(key, value)
	}

	// Build the base URL and append encoded query parameters
	return r.prefix() + "$help&func=" + fn + "&" + values.Encode()
}
