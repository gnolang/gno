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

type Entries []Entry

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

// BulletList returns entries as a markdown bullet list without truncation.
func (l Entries) BulletList() string {
	return l.BulletListTruncated(-1, -1)
}

// BulletListTruncated returns entries as a markdown bullet list with truncation.
func (l Entries) BulletListTruncated(titleMaxLength, bodyMaxLength int) string {
	items := make([]string, len(l))
	for i, entry := range l {
		title := truncateText(entry.Title, titleMaxLength, false)
		if entry.Link != "" {
			title = md.Link(title, entry.Link)
		}
		body := truncateText(entry.Body, bodyMaxLength, true)
		links := strings.Join(entry.Links, " ")
		item := md.Bold(title) + ": " + md.Italic(body) + " " + links
		items[i] = strings.TrimSpace(item)
	}
	return md.BulletList(items)
}

// NumberedList returns entries as a markdown numbered list without truncation.
func (l Entries) NumberedList() string {
	return l.NumberedListTruncated(-1, -1)
}

// NumberedListTruncated returns entries as a markdown numbered list with truncation.
func (l Entries) NumberedListTruncated(titleMaxLength, bodyMaxLength int) string {
	items := make([]string, len(l))
	for i, entry := range l {
		title := truncateText(entry.Title, titleMaxLength, false)
		if entry.Link != "" {
			title = md.Link(title, entry.Link)
		}
		body := truncateText(entry.Body, bodyMaxLength, true)
		links := strings.Join(entry.Links, " ")
		item := md.Bold(title) + ": " + md.Italic(body) + " " + links
		items[i] = strings.TrimSpace(item)
	}
	return md.OrderedList(items)
}

// Paragraphs returns entries as simple paragraphs with bold titles.
func (l Entries) Paragraphs() string {
	paragraphs := make([]string, len(l))
	for i, entry := range l {
		title := entry.Title
		if entry.Link != "" {
			title = md.Link(title, entry.Link)
		}
		paragraph := md.Bold(title) + ". " + entry.Body
		if len(entry.Links) > 0 {
			paragraph += " " + strings.Join(entry.Links, " ")
		}
		paragraphs[i] = paragraph
	}
	return strings.Join(paragraphs, "\n\n")
}

func (l Entries) Columns() string {
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
	return md.Columns(columns, false)
}

func truncateText(content string, maxLength int, withNewLines bool) string {
	// If maxLength is -1, don't truncate
	if maxLength < 0 {
		if !withNewLines {
			content = strings.ReplaceAll(content, "\n", " ")
		}
		return content
	}
	if !withNewLines {
		content = strings.ReplaceAll(content, "\n", " ")
	}
	if len(content) > maxLength {
		return content[:maxLength-3] + "..."
	}
	return content
}
