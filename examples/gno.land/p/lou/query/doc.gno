/*
Package query provides a simple way to handle queries in your realm.

It includes helper functions to extract individual query parameters, convert the full query string into a map,
and update one or more parameters in a URL.

Example:

	import "gno.land/p/lou/query"

	func Render(rawPath string) string {
		queryTime := query.ParseQueryParams("time", rawPath)
		if queryTime == "" {
			queryTime = "default"
		}
		// Handle queryTime value
		return queryTime
	}

*/

package query // import "gno.land/p/lou/query"
