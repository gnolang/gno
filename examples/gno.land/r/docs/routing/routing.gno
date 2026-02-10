package routing

import (
	"strings"

	"gno.land/p/nt/mux"
	"gno.land/p/nt/ufmt"
	"gno.land/r/sys/users"
)

func Render(path string) string {
	// Initialize the router
	router := mux.NewRouter()

	// Then, pass specific path patterns and their function handlers
	// The handler functions need to have a specific signature:
	// func(*mux.ResponseWriter, *mux.Request)

	// Below are some examples for specific path patterns and their handlers

	// When no render path is passed, ie the root render
	router.HandleFunc("", homeHandler)

	// When a specific render path is passed
	router.HandleFunc("staticpage", staticpageHandler)

	// When a render path with a variable is passed, ie `addr`
	router.HandleFunc("user/{name}", profileHandler)

	// When a wildcard path is passed
	router.HandleFunc("wildcard/*/", wildcardHandler)

	// Finally, Pass the render path to the router
	return router.Render(path)
}

func homeHandler(res *mux.ResponseWriter, _ *mux.Request) {
	out := "# Routing\n\n"

	out += `This short example showcases how the [p/demo/mux](/p/demo/mux) package works.
This pure package aims to offer similar functionality to [**http.ServeMux**](https://pkg.go.dev/net/http#ServeMux) in Go, but for Gno's **Render()** requests.

This home page is handled by the homeHandler function. Check out the examples below for more ways
to use this package:
- [A static page](/r/docs/routing:staticpage)
- [A path handling path variables, such as user pages](/r/docs/routing:user/test1)
- [A path handling wildcard](/r/docs/routing:wildcard/wildcardhelp/whatever)
`

	// Write to the result at the end
	res.Write(out)
}

func staticpageHandler(res *mux.ResponseWriter, req *mux.Request) {
	out := "# Welcome to the static page!\n\n"
	out += "There isn't much on this page, but it's cool because shows you that using query parameters is also possible!\n\n"
	out += "Try adding [`?gno_is=value`](/r/docs/routing:staticpage?gno_is=cool) to the URL.\n\n"

	// We can use the Query API to get the values of query parameters
	val := req.Query.Get("gno_is")
	if val != "" {
		out += ufmt.Sprintf("You inputted a value for the parameter `gno_is`: **%s**\n\n", val)

	}

	out += "### Multiple values for the same parameter\n\n"
	out += "The same parameter can have [more than one value](/r/docs/routing:staticpage?gno_is=cool&gno_is=transparent).\n\n"
	out += "We can access it via the `*mux.Request.Query` map.\n\n"

	res.Write(out)
}

func profileHandler(res *mux.ResponseWriter, req *mux.Request) {
	out := "# User display page\n\n"

	// Integrate with r/sys/users to get user data
	name := req.GetVar("name")
	userData, _ := users.ResolveName(name)
	if userData == nil {
		out += "This name does not exist in the [User Registry](/r/sys/users)!"
		res.Write(out)
		return
	}

	out += "Found an address matching the name:\n\n"
	out += userData.RenderLink("")
	res.Write(out)
}

func wildcardHandler(res *mux.ResponseWriter, req *mux.Request) {
	out := "# Wildcard handler\n\n"

	// Use GetVar("*") to access the wildcard part of the route
	wild := req.GetVar("*")
	out += ufmt.Sprintf("This page matched a wildcard route.\n\n")
	out += ufmt.Sprintf("**Wildcard path:** `%s`\n\n", wild)

	out += "You can use this to match and handle arbitrarily nested paths.\n\n"
	out += "Try visiting:\n"
	out += "- [/r/docs/routing:wildcard/foo/bar](/r/docs/routing:wildcard/foo/bar)\n"
	out += "- [/r/docs/routing:wildcard/a/b/c](/r/docs/routing:wildcard/a/b/c)\n"

	// Optionally split the wildcard into segments
	segments := strings.Split(wild, "/")
	out += "\n### Wildcard segments:\n"
	for i, seg := range segments {
		out += ufmt.Sprintf("- Segment %d: `%s`\n", i+1, seg)
	}

	res.Write(out)
}
