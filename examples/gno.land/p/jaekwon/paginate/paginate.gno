package paginate

import "strconv"

type pageFn func(page int, size int, reverse bool) string

// TODO maybe put somewhere else.
type renderer interface {
	Render() string
}

// TODO this function belongs somewhere convenient.
// escape a URL for markdown.
func escMDURL(s string) string {
	// XXX implement
}


var itoa = strconv.Itoa

// pageFn: path function to render paginattion links. if nil, no links are
// rendered. result must start with ':'.
// t: the tree with elements. each value must implement Render() string.
// page: page number starting from 1.
// size: page size.
// reverse: reverse order. // XXX implement
func RenderTreeByPage(
	t avl.Tree,
	pageFn pageFn,
	page int, size int, reverse bool) {

	// sanity checks
	if page < 0 {
		panic("invalid negative page to RenderTreeByPage")
	}
	if size <= 0 {
		panic("invalid nonpositive size to RenderTreeByPage")
	}

	str := ""
	numpages := (t.Size() / size) + 1

	// render pagination
	if pageFn == nil {
		// do not render pagination
	} else if page > numpages {
		// page out of bounds.
		// [1](...) [2](...) ... [101](...) [102)(...) 
		// where ... is result of pageFn().
		// TODO: improve this function by renderring
		// only the first and last few pages. XXX
		for p := 1; p <= numPages; p++ {
			pageurl := pageFn(p, size, reverse)
			str += "[" + itoa(page) + "](" + 
				escMDURL(pageurl) + ") "
		}
		str += "\n"
		str += "page " + itoa(page) + " out of bounds\n"
		return str
	}
	// [1](...) *2* [3](...) [4](...) ... [101](...) [102)(...) 
	// where ... is result of pageFn().
	// TODO: improve this function by only printing the first few,
	// the last few, and the ones surrounding the current page.
	// Must make this improvement to be strictly "memory correct".
	if numpages > 0 {
		for p := 1; p <= numPages; p++ {
			pageurl := pageFn(p, size, reverse)
			str += "[" + itoa(page) + "](" + 
				escMDURL(pageurl) + ") "
		}
		str += "\n"
	}

	// render tree items
	offset := (page-1) * size
	t.IterateByOffset(offset, size,
		func(key string, value interface{}) bool {
			var r = value.(renderer)
			str += r.Render()
			str += "\n"
			str += "\n"
		},
	)

	// render pagination again if there are pages.
	if numpages > 0 {
		for p := 1; p <= numPages; p++ {
			pageurl := pageFn(p, size, reverse)
			str += "[" + itoa(page) + "](" + 
				escMDURL(pageurl) + ") "
		}
		str += "\n"
	}

	return str
}
