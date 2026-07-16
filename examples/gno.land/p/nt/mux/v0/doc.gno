// v0 - Unaudited: This is an initial version that has not yet been formally audited.
// A fully audited version will be published as a subsequent release.
// Use in production at your own risk.
//
// Package mux provides a simple routing and rendering library for handling dynamic path-based requests in Gno contracts.
//
// The `mux` package aims to offer similar functionality to `http.ServeMux` in Go, but for Gno's Render() requests.
// It allows you to define routes with dynamic parts and associate them with corresponding handler functions for rendering outputs.
//
// Usage:
// 1. Create a new Router instance using `NewRouter()` to handle routing and rendering logic.
// 2. Register routes and their associated handler functions using the `Handle(route, handler)` method.
// 3. Implement the rendering logic within the handler functions, utilizing the `Request` and `ResponseWriter` types.
// 4. Use the `Render(path)` method to process a given path and execute the corresponding handler function to obtain the rendered output.
//
// Route Patterns:
// Routes can include dynamic parts enclosed in braces, such as "users/{id}" or "hello/{name}". The `Request` object's `GetVar(key)`
// method allows you to extract the value of a specific variable from the path based on routing rules.
//
// Example:
//
//	router := mux.NewRouter()
//
//	// Define a route with a variable and associated handler function
//	router.HandleFunc("hello/{name}", func(res *mux.ResponseWriter, req *mux.Request) {
//		name := req.GetVar("name")
//		if name != "" {
//			res.Write("Hello, " + name + "!")
//		} else {
//			res.Write("Hello, world!")
//		}
//	})
//
//	// Render the output for the "/hello/Alice" path
//	output := router.Render("hello/Alice")
//	// Output: "Hello, Alice!"
//
// Note: The `mux` package provides a basic routing and rendering mechanism for simple use cases. For more advanced routing features,
// consider using more specialized libraries or frameworks.
package mux
