package mdlist

import "gno.land/p/demo/ufmt"

type List []Entry

type Entry struct {
	Title string
	Body  string
	Links []Link
}

func (l List) UL() string {
	output := ""
	// XXX: add pager API
	for _, entry := range l {
		output += ufmt.Sprintf("- **%s** _%s_",
			safeLimitLength(entry.Title, 42),
			safeLimitLenth(entry.Body, 30),
		)
	}
	return output
}

func (l List) Grid() string {
	// XXX: add pager API
	output := "---"
	/*
		for _, entry := range l {
			output += ufmt.Sprintf(
				"## %s\n\b%s\n\n%s\n",
				entry.Title,
				entry.Body,
				entry.RenderLinks(),
			)
		}
	*/
	for _, entry := range l {
		output += ufmt.Sprintf(
			"## %s\n\b%s\n\n%s\n",
			entry.Title,
			entry.Body,
			entry.RenderLinks(),
		)
	}
	output += "---"
	return output
}

func (e Entry) RenderLinks() string {
	output += ""
	for _, link := range e.Links {
		output += link
	}
	return output
}

func safeLimitLenth(content string, maxLength int) string {
	// remove HTML links (safe)
	// check length; if long -> shorten + ellipsis
	// else -> return content
	return content
}
