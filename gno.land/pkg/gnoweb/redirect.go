package gnoweb

import (
	"net/http"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

// Redirects is a map of gnoweb paths that are redirected using RedirectMiddleware.
var Redirects = map[string]string{
	"/r/demo/boards:gnolang/6": "/r/demo/boards:gnolang/3", // XXX: temporary
	"/blog":                    "/r/gnoland/blog",
	"/gor":                     "/contribute",
	"/game-of-realms":          "/contribute",
	"/grants":                  "/partners",
	"/language":                "/gnolang",
	"/getting-started":         "/start",
	"/newsletter":              "/r/gnoland/pages:p/newsletter",
}

// RedirectMiddleware redirects all incoming requests whose path matches
// any of the Redirects to the corresponding URL, and renders a redirect view.
func RedirectMiddleware(next http.Handler, analytics bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request path matches a redirect
		if newPath, ok := Redirects[r.URL.Path]; ok {
			http.Redirect(w, r, newPath, http.StatusFound)
			components.RedirectView(components.RedirectData{
				To:            newPath,
				WithAnalytics: analytics,
			}).Render(w)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}
