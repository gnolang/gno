/*
Package query provides a simple way to handle queries in your realm.

It includes helper functions to extract individual query parameters, convert the full query string into a map,
and update one or more parameters in a URL.

Example:

	import (
		"gno.land/p/lou/query"
		"gno.land/p/moul/md"
	)

	func Render(rawPath string) string {
		queryMode := query.ParseQueryParams("mode", rawPath)
		if queryMode == "" {
			queryMode = "grid" // default
		}
		// Handle queryMode value
		out := md.Link("list mode", query.UpdateQuery(rawPath, "mode", "list")) + "\n\n"
		out += md.Link("grid mode", query.UpdateQuery(rawPath, "mode", "grid")) +"\n\n"
		return out
	}

*/

package query // import "gno.land/p/lou/query"
