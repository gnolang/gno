package boards2

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/mux"
)

type PaginationOpts struct {
	Pager      pager.Pager
	PageNumber int
}

// Iterate loops over an a page.
//
// Loops though all avl tree contents if PaginationOpts is nil.
func (opts *PaginationOpts) Iterate(tree *avl.Tree, cb func(k string, val interface{}) bool) *pager.Page {
	if opts == nil {
		tree.Iterate("", "", cb)
		return nil
	}

	opts.Pager.Tree = tree
	page := opts.Pager.GetPage(opts.PageNumber)
	for _, item := range page.Items {
		if cb(item.Key, item.Value) {
			break
		}
	}

	return page
}

func mustGetPagination(req *mux.Request, pageSize int) *PaginationOpts {
	p := pager.Pager{
		PageQueryParam:  "page",
		DefaultPageSize: pageSize,
	}

	pageNumber, _, err := p.ParseQuery(req.RawPath)
	if err != nil {
		panic(err)
	}

	return &PaginationOpts{
		Pager:      p,
		PageNumber: pageNumber,
	}
}
