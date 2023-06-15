package paginate

import (
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/md"
)

const PaginateGap = 3

type pageFn func(page int, size int, reverse bool) string

// TODO maybe put somewhere else.
type renderer interface {
	Render() string
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func navigation(numPages int, current int, reverse bool, pageFn pageFn) string {
	str := ""

	i := 1
	upperLimit := min(PaginateGap+1, current)
	for ; i < upperLimit; i++ {
		pageurl := pageFn(i, numPages, reverse)
		str += md.Link(strconv.Itoa(i), pageurl) + " "
	}
	if i-1 == PaginateGap && i != current {
		if i == current-PaginateGap/2 {
			pageurl := pageFn(i, numPages, reverse)
			str += md.Link(strconv.Itoa(i), pageurl) + " "
			i += 1
		} else {
			str += "... "
		}
	}

	// sanity check
	if current > numPages {
		for i = numPages - PaginateGap + 1; i <= numPages; i++ {
			pageurl := pageFn(i, numPages, reverse)
			str += md.Link(strconv.Itoa(i), pageurl) + " "
		}
		return str
	}

	if i < current-PaginateGap/2 {
		i = current - PaginateGap/2
	}

	upperLimit = min(current+PaginateGap/2+1, numPages)
	for ; i < upperLimit; i++ {
		if i == current {
			str += "*" + strconv.Itoa(i) + "* "
			continue
		}
		pageurl := pageFn(i, numPages, reverse)
		str += md.Link(strconv.Itoa(i), pageurl) + " "
	}
	if i < numPages-PaginateGap+1 {
		if i == numPages-PaginateGap {
			pageurl := pageFn(i, numPages, reverse)
			str += md.Link(strconv.Itoa(i), pageurl) + " "
		} else {
			str += "... "
		}
		i = numPages - PaginateGap + 1
	}
	for ; i <= numPages; i++ {
		pageurl := pageFn(i, numPages, reverse)
		str += md.Link(strconv.Itoa(i), pageurl) + " "
	}
	return str
}

// This syntax actually doesn't work with Gno language
// var itoa = strconv.Itoa

// pageFn: path function to render paginattion links. if nil, no links are
// rendered. result must start with ':'.
// t: the tree with elements. each value must implement Render() string.
// page: page number starting from 1.
// size: page size.
// reverse: reverse order.
func RenderTreeByPage(
	t avl.Tree,
	pageFn pageFn,
	page int, size int, reverse bool,
) string {
	// sanity checks
	if page < 0 {
		panic("invalid negative page to RenderTreeByPage")
	}
	if size <= 0 {
		panic("invalid nonpositive size to RenderTreeByPage")
	}

	str := ""
	numPages := (t.Size() / size)
	if numPages < 1 {
		return ""
	}

	// render pagination
	if pageFn == nil {
		// render tree items
		var offset int
		if reverse {
			offset = (numPages - page) * size
		} else {
			offset = (page - 1) * size
		}
		t.IterateByOffset(offset, size,
			func(key string, value interface{}) bool {
				r := value.(renderer)
				str += "\n"
				str += r.Render()
				str += "\n"
				str += "\n"

				return true
			},
		)
		return str
	} else if page > numPages {
		// page out of bounds.
		// [1](...) [2](...) ... [101](...) [102)(...)
		// where ... is result of pageFn().
		str += navigation(numPages, page, reverse, pageFn)
		str += "\n"
		str += "page " + strconv.Itoa(page) + " out of bounds\n"
		return str
	} else {
		// [1](...) *2* [3](...) [4](...) ... [101](...) [102)(...)
		// where ... is result of pageFn().
		str += navigation(numPages, page, reverse, pageFn)
		str += "\n"

		// render tree items
		var offset int
		if reverse {
			offset = (numPages - page) * size
		} else {
			offset = (page - 1) * size
		}
		t.IterateByOffset(offset, size,
			func(key string, value interface{}) bool {
				r := value.(renderer)
				str += "\n"
				str += r.Render()
				str += "\n"
				str += "\n"

				return true
			},
		)
	}

	// render pagination again if there are pages.
	str += navigation(numPages, page, reverse, pageFn)
	str += "\n"

	return str
}
