package gnoweb

import "net/http"

// realm aliases
var Aliases = map[string]string{
	"/":           "/r/gnoland/home",
	"/about":      "/r/gnoland/pages:p/about",
	"/gnolang":    "/r/gnoland/pages:p/gnolang",
	"/ecosystem":  "/r/gnoland/pages:p/ecosystem",
	"/partners":   "/r/gnoland/pages:p/partners",
	"/testnets":   "/r/gnoland/pages:p/testnets",
	"/start":      "/r/gnoland/pages:p/start",
	"/license":    "/r/gnoland/pages:p/license",
	"/contribute": "/r/gnoland/pages:p/contribute",
	"/events":     "/r/gnoland/events",
}

// http redirects
var Redirects = map[string]string{
	"/r/demo/boards:gnolang/6": "/r/demo/boards:gnolang/3", // XXX: temporary
	"/blog":                    "/r/gnoland/blog",
	"/gor":                     "/contribute",
	"/game-of-realms":          "/contribute",
	"/grants":                  "/partners",
	"/language":                "/gnolang",
	"/getting-started":         "/start",
	"/gophercon24":             "https://docs.gno.land",
}

func AliasAndRedirectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request path matches an alias
		if newPath, ok := Aliases[r.URL.Path]; ok {
			r.URL.Path = newPath
		}

		// Check if the request path matches a redirect
		if newPath, ok := Redirects[r.URL.Path]; ok {
			http.Redirect(w, r, newPath, http.StatusFound)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
