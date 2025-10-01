// Package helplink provides utilities for creating help page links compatible
// with Gnoweb, Gnobro, and other clients that support the Gno contracts'
// flavored Markdown format.
//
// This package simplifies the generation of dynamic, context-sensitive help
// links, enabling users to navigate relevant documentation seamlessly within
// the Gno ecosystem.
//
// For a more lightweight alternative, consider using p/moul/txlink.
//
// The primary functions — Func, FuncURL, and Home — are intended for use with
// the "relative realm". When specifying a custom Realm, you can create links
// that utilize either the current realm path or a fully qualified path to
// another realm.
package helplink

import (
	"chain/runtime"
	"strings"

	"gno.land/p/moul/txlink"
)

var chainDomain = runtime.ChainDomain()

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
	rlmstr := string(r)
	if strings.HasPrefix(rlmstr, chainDomain) {
		return strings.TrimPrefix(rlmstr, chainDomain)
	}

	// remote realm -> https://remote.land/realm
	return "https://" + rlmstr
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
	tlr := txlink.Realm(r)
	return tlr.Call(fn, args...)
}

// Home returns the base help URL for the specified realm.
func (r Realm) Home() string {
	return r.prefix() + "$help"
}
