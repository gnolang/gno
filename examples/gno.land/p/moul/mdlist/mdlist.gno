// Package mdlist is a proof-of-concept for composable UIs patterns.
//
// It allows specifying a data source and then dynamically rendering it as
// a list or a grid.
package mdlist

import (
	"strings"

	"gno.land/p/moul/md"
)

// XXX: configurable length UL
// XXX: moul/pager.Pager-powered auto mdlist
// XXX: func (l List) Render(path string) string {
// XXX: add .Table()

var defaultEncoder = NewEncoder()

type List []Entry

type Entry struct {
	// Title is the title of the entry. Required.
	Title string
	// Body is the body of the entry. Required.
	Body string
	// Link is the URL of the entry. Optional.
	Link string
	// Links is a list of additional URLs of the entry. Optional.
	Links []string
}

func (l List) UL() string {
	items := make([]string, len(l))
	for i, entry := range l {
		title := truncateText(entry.Title, 40, false)
		if entry.Link != "" {
			title = md.Link(title, entry.Link)
		}
		body := truncateText(entry.Body, 80, true)
		links := strings.Join(entry.Links, " ")
		item := md.Bold(title) + ": " + md.Italic(body) + " " + links
		items[i] = strings.TrimSpace(item)
	}
	return md.BulletList(items)
}

func (l List) Grid() string {
	columns := make([]string, len(l))
	for i, entry := range l {
		title := entry.Title
		if entry.Link != "" {
			title = md.Link(title, entry.Link)
		}
		body := entry.Body
		links := strings.Join(entry.Links, " ")
		columns[i] = md.H2(title) + md.Paragraph(body) + md.Paragraph(links)
	}
	return md.Columns(columns)
}

func truncateText(content string, maxLength int, withNewLines bool) string {
	if !withNewLines {
		content = strings.ReplaceAll(content, "\n", " ")
	}
	if len(content) > maxLength {
		return content[:maxLength-3] + "..."
	}
	return content
}

type Encoder struct {
	MaxWidth int
	Mode     string
	Page     int
	Limit    int
}

func NewEncoder() *Encoder {
	return &Encoder{
		MaxWidth: 80,
	}
}
