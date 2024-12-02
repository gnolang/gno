package boards2

import (
	"strings"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/mux"
)

type wrappedRouter struct {
	*mux.Router

	requestURI string
}

// newWrappedRouter returns a wrapped mux router that supports query strings.
//
// TODO: remove this when query strings PR will be merged: https://github.com/gnolang/gno/pull/2594
func newWrappedRouter(r *mux.Router) *wrappedRouter {
	return &wrappedRouter{
		Router: r,
	}
}

func (r *wrappedRouter) Render(path string) string {
	reqPath := trimRequestPath(path)
	r.requestURI = path
	return r.Router.Render(reqPath)
}

func (r *wrappedRouter) HandleFunc(pattern string, fn mux.HandlerFunc) {
	r.Router.HandleFunc(pattern, func(rw *mux.ResponseWriter, req *mux.Request) {
		fn(rw, &mux.Request{
			Path:        r.requestURI,
			HandlerPath: req.HandlerPath,
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

func getPagination(r *mux.Request, defaultPageSize int) Pagination {
	// reuse query parameter extract logic from pager.
	pg := pager.NewPager(nil, defaultPageSize)
	pg.SizeQueryParam = ""
	pageNum, limit, err := pg.ParseQuery(r.Path)
	if err != nil {
		panic(err)
	}

	return Pagination{
		PageNumber: pageNum,
		PageSize:   limit,
	}
}
