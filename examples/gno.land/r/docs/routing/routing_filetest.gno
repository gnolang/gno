package main

import (
	"gno.land/r/docs/routing"
)

func main() {
	// Home route
	println("=== HOME ===")
	println(routing.Render(""))

	// Static page (no query param)
	println("=== STATIC ===")
	println(routing.Render("staticpage"))

	// Unknown user profile
	println("=== PROFILE (missing) ===")
	println(routing.Render("user/nobody_xyz_abc"))

	// Wildcard route
	println("=== WILDCARD ===")
	println(routing.Render("wildcard/foo/bar"))
}

// Output:
// === HOME ===
// # Routing
//
// This short example showcases how the [p/nt/mux/v0](/p/nt/mux/v0) package works.
// This pure package aims to offer similar functionality to [**http.ServeMux**](https://pkg.go.dev/net/http#ServeMux) in Go, but for Gno's **Render()** requests.
//
// This home page is handled by the homeHandler function. Check out the examples below for more ways
// to use this package:
// - [A static page](/r/docs/routing:staticpage)
// - [A path handling path variables, such as user pages](/r/docs/routing:user/test1)
// - [A path handling wildcard](/r/docs/routing:wildcard/wildcardhelp/whatever)
//
// === STATIC ===
// # Welcome to the static page!
//
// There isn't much on this page, but it's cool because shows you that using query parameters is also possible!
//
// Try adding [`?gno_is=value`](/r/docs/routing:staticpage?gno_is=cool) to the URL.
//
// ### Multiple values for the same parameter
//
// The same parameter can have [more than one value](/r/docs/routing:staticpage?gno_is=cool&gno_is=transparent).
//
// We can access it via the `*mux.Request.Query` map.
//
//
// === PROFILE (missing) ===
// # User display page
//
// This name does not exist in the [User Registry](/r/sys/users)!
// === WILDCARD ===
// # Wildcard handler
//
// This page matched a wildcard route.
//
// **Wildcard path:** `foo/bar`
//
// You can use this to match and handle arbitrarily nested paths.
//
// Try visiting:
// - [/r/docs/routing:wildcard/foo/bar](/r/docs/routing:wildcard/foo/bar)
// - [/r/docs/routing:wildcard/a/b/c](/r/docs/routing:wildcard/a/b/c)
//
// ### Wildcard segments:
// - Segment 1: `foo`
// - Segment 2: `bar`
//
