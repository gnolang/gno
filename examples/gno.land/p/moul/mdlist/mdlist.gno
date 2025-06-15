package mdlist

import (
	"strings"

	"gno.land/p/moul/md"
)

type List []Entry

type Entry struct {
	Title string
	Body  string
	Links []string
}

func (l List) UL() string {
	return l.ULPager(1, len(l))
}

func (l List) Grid() string {
	return l.GridPager(1, len(l))
}

func (l List) ULPager(page, limit int) string {
	// XXX: add pager API
	items := make([]string, len(l))
	for i, entry := range l {
		title := truncateText(entry.Title, 42)
		body := truncateText(entry.Body, 30)
		items[i] = md.Bold(title) + ": " + md.Italic(body) + " " + strings.Join(entry.Links, " ")
	}
	return md.BulletList(items)
}

func (l List) GridPager(page, limit int) string {
	// XXX: add pager API
	columns := make([]string, len(l))
	for i, entry := range l {
		title := md.H2(entry.Title)
		body := md.Paragraph(entry.Body)
		links := md.Paragraph(strings.Join(entry.Links, " "))
		columns[i] = title + body + links
	}
	return md.Columns(columns)
}

func truncateText(content string, maxLength int) string {
	if len(content) > maxLength {
		return content[:maxLength] + "..."
	}
	return content
}
