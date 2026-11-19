package gnoweb

import (
	"net/http"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

// Redirects is a map of gnoweb paths that are redirected using RedirectMiddleware.
var Redirects = map[string]string{
	"/blog":            "/r/gnoland/blog",
	"/gor":             "/contribute",
	"/game-of-realms":  "/contribute",
	"/grants":          "/partners",
	"/language":        "/gnolang",
	"/getting-started": "/start",
}

// RedirectMiddleware redirects all incoming requests whose path matches
// any of the Redirects to the corresponding URL, and renders a redirect view.
func RedirectMiddleware(next http.Handler, meta StaticMetadata) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if the request path matches a redirect
		if newPath, ok := Redirects[r.URL.Path]; ok {
			serveRedirect(w, r, newPath, meta)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

// serveRedirect sends the client to to, with a body for one that does not
// follow.
func serveRedirect(w http.ResponseWriter, r *http.Request, to string, meta StaticMetadata) {
	http.Redirect(w, r, to, http.StatusFound)
	components.RedirectView(components.RedirectData{
		To:        to,
		Analytics: meta.RedirectAnalytics(),
	}).Render(w)
}
