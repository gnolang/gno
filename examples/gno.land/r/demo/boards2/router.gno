package boards2

import (
	"strings"

	"gno.land/p/demo/mux"
)

type handleFunc = func(*mux.ResponseWriter, request)

type request struct {
	*mux.Request

	requestUri string
}

type wrappedRouter struct {
	*mux.Router

	requestUri string
}

func newWrappedRouter() *wrappedRouter {
	return &wrappedRouter{
		Router: mux.NewRouter(),
	}
}

func (r *wrappedRouter) Render(path string) string {
	reqPath := trimRequestPath(path)
	r.requestUri = path
	return r.Router.Render(reqPath)
}

func (r *wrappedRouter) HandleFunc(pattern string, fn handleFunc) {
	r.Router.HandleFunc(pattern, func(rw *mux.ResponseWriter, req *mux.Request) {
		fn(rw, request{
			Request:    req,
			requestUri: r.requestUri,
		})
	})
}

// trimRequestPath returns request path without query params
func trimRequestPath(path string) string {
	i := strings.Index(path, "?")
	if i == -1 {
		return path
	}

	return path[:i]
}
