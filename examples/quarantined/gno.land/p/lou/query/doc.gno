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
		mode, err := query.GetQueryValueFromURL("mode", rawPath)
		if err != nil || mode == "" {
			mode = "grid" // default value
		}

		// Handle queryMode value
		listURL, _ := query.UpdateQueryValue(rawPath, "mode", "list")
		gridURL, _ := query.UpdateQueryValue(rawPath, "mode", "grid")

		out := md.Link("list mode", listURL) + "\n\n"
		out += md.Link("grid mode", gridURL) + "\n\n"
		return out
	}

*/

package query // import "gno.land/p/lou/query"
