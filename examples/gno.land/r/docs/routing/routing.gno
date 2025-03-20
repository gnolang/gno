package routing

import (
	"gno.land/p/demo/mux"
	"gno.land/r/sys/users"
)

var (
	// Initialize the router object
	router = mux.NewRouter()
)

func init() {
	// Pass specific path patterns and their function handlers
	// The handler functions need to have a specific signature:
	// func(*mux.ResponseWriter, *mux.Request)

	// Below are some examples for specific path patterns and their handlers

	// When no render path is passed, ie the root render
	router.HandleFunc("", homeHandler)

	// When a specific render path is passed
	router.HandleFunc("staticpage", staticpageHanlder)

	// When a render path with a variable is passed, ie `addr`
	router.HandleFunc("user/{name}", profileHandler)

	// When a wildcard path is passed
	router.HandleFunc("wildcard/*/", wildcardHandler)

	// Todo add query param examples
}

// Render displays the current number value, last update timestamp, and a link to call Add with 42
func Render(path string) string {
	return router.Render(path)
}

func homeHandler(res *mux.ResponseWriter, _ *mux.Request) {
	out := "# Routing\n\n"

	out += `This short example showcases how the [p/demo/mux](/p/demo/mux) package.
This pure package aims to offer similar functionality to "http.ServeMux" in Go, but for Gno's Render() requests.

This home page is handled by the homeHandler function. Check out the examples below for more ways
to use this package:
- [A static page](/r/docs/routing:staticpage)
- [A path handling path variables, such as user pages](/r/docs/routing:user/test1)
- [A path handling wildcard](/r/docs/routing:wildcard/wildcardhelp})
`

	// Write to the result at the end
	res.Write(out)
}

func staticpageHanlder(res *mux.ResponseWriter, _ *mux.Request) {
	out := "# Welcome to the static page!\n\n"
	out += "There isn't much on this page, but it's cool because it routes you to [another special page]()!"
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
	// todo
}
