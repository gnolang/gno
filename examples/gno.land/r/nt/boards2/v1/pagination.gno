package boards2

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
)

type PaginationOpts struct {
	pager      pager.Pager
	pageNumber int
}

// Iterate loops over an a page.
//
// Loops though all avl tree contents if PaginationOpts is nil.
func (opts *PaginationOpts) Iterate(tree *avl.Tree, cb func(k string, val interface{}) bool) *pager.Page {
	if opts == nil {
		tree.Iterate("", "", cb)
		return nil
	}

	opts.pager.Tree = tree
	page := opts.pager.GetPage(opts.pageNumber)
	for _, item := range page.Items {
		if cb(item.Key, item.Value) {
			break
		}
	}

	if page.TotalPages > 1 {
		return page
	}
	return nil
}

func mustGetPagination(rawPath string, pageSize int) *PaginationOpts {
	p := pager.Pager{
		PageQueryParam:  "page",
		DefaultPageSize: pageSize,
	}

	pageNumber, _, err := p.ParseQuery(rawPath)
	if err != nil {
		panic(err)
	}

	return &PaginationOpts{
		pager:      p,
		pageNumber: pageNumber,
	}
}
