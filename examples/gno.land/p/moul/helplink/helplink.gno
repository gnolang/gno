// Package helplink provides utilities for crafting help page links that work
// with Gnoweb, Gnobro, and other clients supporting the flavored markdown of
// Gno contracts.
//
// This package is designed to facilitate the creation of dynamic help links,
// making it easier for users to access relevant documentation within Gno's
// ecosystem.
package helplink

// XXX: support advanced help options like field type
// XXX: follow https://github.com/gnolang/gno/issues/2602
// XXX: follow https://github.com/gnolang/gno/pull/2876

// Link creates a markdown link for a help page.
func Link(title string, fn string, args ...string) string {
	return "[" + title + "](" + URL(fn, args...) + ")"
}

// URL constructs a URL for the specified function with optional key-value
// arguments.
func URL(fn string, args ...string) string {
	// Start with the base query
	url := "?help&__func=" + fn

	// Check if args length is even
	if len(args)%2 != 0 {
		// If not even, we can choose to handle the error here
		// For example, we can just return the URL without appending more args
		return url
	}

	// Append key-value pairs to the URL
	for i := 0; i < len(args); i += 2 {
		key := args[i]
		value := args[i+1]
		url += "&" + key + "=" + value
	}

	return url
}

// Home returns the base help URL.
func Home() string {
	return "?help"
}
