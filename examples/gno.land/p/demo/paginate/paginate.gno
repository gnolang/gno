package paginate

import (
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
)

const PAGINATE_GAP = 3

type pageFn func(page int, size int, reverse bool) string

// TODO maybe put somewhere else.
type renderer interface {
	Render() string
}

// TODO this function belongs somewhere convenient.
// escape a URL for markdown.
func escMDURL(s string) string {
	schars := map[string]string{
		"%":  "%25",
		" ":  "%20",
		"!":  "%21",
		"\"": "%22",
		"$":  "%24",
		"&":  "%26",
		"'":  "%27",
		"(":  "%28",
		")":  "%29",
		"*":  "%2A",
		"+":  "%2B",
		",":  "%2C",
		";":  "%3B",
		"<":  "%3C",
		">":  "%3E",
		"?":  "%3F",
		"@":  "%40",
		"[":  "%5B",
		"\\": "%5C",
		"]":  "%5D",
		"^":  "%5E",
		"_":  "%5F",
		"`":  "%60",
		"{":  "%7B",
		"|":  "%7C",
		"}":  "%7D",
		"~":  "%7E",
		"\n": "%0A",
	}
	for k, v := range schars {
		s = strings.Replace(s, k, v, -1)
	}
	return s
}

func navigation(numPages int, current int, reverse bool, pageFn pageFn) string {
	str := ""

	i := 1
	for ; i <= PAGINATE_GAP && i < current; i++ {
		pageurl := pageFn(i, numPages, reverse)
		str += "[" + strconv.Itoa(i) + "](" +
			escMDURL(pageurl) + ") "
	}
	if i-1 == PAGINATE_GAP && i != current {
		if i == current-PAGINATE_GAP/2 {
			pageurl := pageFn(i, numPages, reverse)
			str += "[" + strconv.Itoa(i) + "](" +
				escMDURL(pageurl) + ") "
			i += 1
		} else {
			str += "... "
		}
	}

	// sanity check
	if current > numPages {
		for i = numPages - PAGINATE_GAP + 1; i <= numPages; i++ {
			pageurl := pageFn(i, numPages, reverse)
			str += "[" + strconv.Itoa(i) + "](" +
				escMDURL(pageurl) + ") "
		}
		return str
	}

	if i < current-PAGINATE_GAP/2 {
		i = current - PAGINATE_GAP/2
	}
	for ; i <= current+PAGINATE_GAP/2 && i < numPages; i++ {
		if i == current {
			str += "*" + strconv.Itoa(i) + "* "
			continue
		}
		pageurl := pageFn(i, numPages, reverse)
		str += "[" + strconv.Itoa(i) + "](" +
			escMDURL(pageurl) + ") "
	}
	if i < numPages-PAGINATE_GAP+1 {
		if i == numPages-PAGINATE_GAP {
			pageurl := pageFn(i, numPages, reverse)
			str += "[" + strconv.Itoa(i) + "](" +
				escMDURL(pageurl) + ") "
		} else {
			str += "... "
		}
		i = numPages - PAGINATE_GAP + 1
	}
	for ; i <= numPages; i++ {
		pageurl := pageFn(i, numPages, reverse)
		str += "[" + strconv.Itoa(i) + "](" +
			escMDURL(pageurl) + ") "
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
		panic("invalid tree size to RenderTreeByPage")
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
			},
		)
	}

	// render pagination again if there are pages.
	str += navigation(numPages, page, reverse, pageFn)
	str += "\n"

	return str
}
