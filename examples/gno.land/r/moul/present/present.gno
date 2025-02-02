package present

import (
	"net/url"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/collection"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/realmpath"
	"gno.land/p/moul/txlink"
)

type Presentation struct {
	ID        string
	Title     string
	Event     string
	Author    string
	Uploader  std.Address
	Date      time.Time
	Content   string
	EditDate  time.Time
	NumSlides int
}

var (
	presentations *collection.Collection
	Ownable       *ownable.Ownable
)

func init() {
	presentations = collection.New()
	presentations.AddIndex("id", func(v interface{}) string {
		return v.(*Presentation).ID
	}, collection.UniqueIndex)
	presentations.AddIndex("date", func(v interface{}) string {
		return v.(*Presentation).Date.String()
	}, collection.DefaultIndex)
	presentations.AddIndex("uploader", func(v interface{}) string {
		return v.(*Presentation).Uploader.String()
	}, collection.DefaultIndex)
	presentations.AddIndex("event", func(v interface{}) string {
		return v.(*Presentation).Event
	}, collection.DefaultIndex)

	Ownable = ownable.New()
}

// Render handles the realm's rendering logic
func Render(path string) string {
	req := realmpath.Parse(path)
	switch req.PathPart(0) {
	case "":
		return renderList(req)
	case "view":
		return renderView(req)
	case "slides":
		return renderSlides(req)
	default:
		return "404: not found"
	}
}

// Set adds or updates a presentation
func Set(id, title, event, author, date, content string) string {
	Ownable.AssertCallerIsOwner()

	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return "400: invalid date format (expected: YYYY-MM-DD)"
	}

	numSlides := 1 // Count intro slide
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(line, "## ") {
			numSlides++
		}
	}

	p := &Presentation{
		ID:        id,
		Title:     title,
		Event:     event,
		Author:    author,
		Uploader:  std.PrevRealm().Addr(),
		Date:      parsedDate,
		Content:   content,
		EditDate:  time.Now(),
		NumSlides: numSlides,
	}

	presentations.Set(p)
	return "presentation saved successfully"
}

// Delete removes a presentation
func Delete(id string) string {
	Ownable.AssertCallerIsOwner()

	entry := presentations.GetFirst("id", id)
	if entry == nil {
		return "404: presentation not found"
	}

	if entry.Obj.(*Presentation).Uploader != std.PrevRealm().Addr() {
		return "401: unauthorized - only the uploader can delete their presentations"
	}

	// Convert the entry's ID from string to uint64 and delete
	numericID, err := seqid.FromString(entry.ID)
	if err != nil {
		return "500: invalid entry ID format"
	}

	presentations.Delete(uint64(numericID))
	return "presentation deleted successfully"
}

func renderList(req *realmpath.Request) string {
	var out strings.Builder
	out.WriteString(md.H1("Presentations"))

	// Setup pager
	pgr := pager.NewPager(presentations.GetIndex(getSortField(req)), 10, isSortReversed(req))

	// Get current page
	page := pgr.MustGetPageByPath(req.Query.Encode())

	// Create table
	table := mdtable.Table{
		Headers: []string{"Title", "Author", "Event", "Date", "Slides"},
	}

	// Add sort options
	out.WriteString("Sort by: ")
	out.WriteString(renderSortLink(req, "date", "Date"))
	out.WriteString(" | ")
	out.WriteString(renderSortLink(req, "uploader", "Uploader"))
	out.WriteString(" | ")
	out.WriteString(renderSortLink(req, "event", "Event"))
	out.WriteString("\n\n")

	// Add rows from current page
	for _, item := range page.Items {
		// Get the actual presentation using the ID from the index
		entry := presentations.GetFirst("id", item.Key)
		if entry == nil {
			continue
		}
		p := entry.Obj.(*Presentation)

		query := make(url.Values)
		query.Set("view", p.ID)
		table.Append([]string{
			ufmt.Sprintf("[%s](%s)", p.Title, localPath("view", query)),
			p.Author,
			p.Event,
			p.Date.Format("2006-01-02"),
			ufmt.Sprintf("%d", p.NumSlides),
		})
	}

	out.WriteString(table.String())
	out.WriteString("\n\n")
	out.WriteString(page.Picker())
	return out.String()
}

func renderView(req *realmpath.Request) string {
	id := req.Query.Get("view")
	if id == "" {
		return "400: missing presentation ID"
	}

	entry := presentations.GetFirst("id", id)
	if entry == nil {
		return "404: presentation not found"
	}

	p := entry.Obj.(*Presentation)
	var out strings.Builder

	// Header
	out.WriteString(md.H1(p.Title))
	out.WriteString(ufmt.Sprintf("Event: %s\n\n", p.Event))
	out.WriteString(ufmt.Sprintf("Author: %s\n\n", p.Author))
	out.WriteString(ufmt.Sprintf("Date: %s\n\n", p.Date.Format("2006-01-02")))

	// Slide mode link
	query := make(url.Values)
	query.Set("slides", p.ID)
	out.WriteString(ufmt.Sprintf("[View as slides](%s)\n\n", localPath("slides", query)))
	out.WriteString(md.HorizontalRule())

	// Content
	out.WriteString(p.Content)

	// Metadata footer
	out.WriteString("\n\n")
	out.WriteString(md.HorizontalRule())
	out.WriteString(ufmt.Sprintf("Last edited: %s\n\n", p.EditDate.Format("2006-01-02 15:04:05")))
	out.WriteString(ufmt.Sprintf("ID: `%s`\n\n", p.ID))
	out.WriteString(ufmt.Sprintf("Uploader: `%s`\n\n", p.Uploader))
	out.WriteString(ufmt.Sprintf("Number of slides: %d\n\n", p.NumSlides))

	// Admin actions
	if p.Uploader == std.PrevRealm().Addr() {
		out.WriteString("\n\n")
		out.WriteString("Admin actions:\n\n")
		// Edit link
		editLink := txlink.Call("Set",
			"id", p.ID,
			"title", p.Title,
			"event", p.Event,
			"date", p.Date.Format("2006-01-02"),
		)
		out.WriteString(ufmt.Sprintf("- [Edit presentation](%s)\n", editLink))

		// Delete link
		deleteLink := txlink.Call("Delete", "id", p.ID)
		out.WriteString(ufmt.Sprintf("- [Delete presentation](%s)\n", deleteLink))
	}

	return out.String()
}

func renderSlides(req *realmpath.Request) string {
	id := req.Query.Get("slides")
	if id == "" {
		return "400: missing presentation ID"
	}

	entry := presentations.GetFirst("id", id)
	if entry == nil {
		return "404: presentation not found"
	}

	p := entry.Obj.(*Presentation)
	currentPage := 1
	if page := req.Query.Get("page"); page != "" {
		var err error
		currentPage, err = strconv.Atoi(page)
		if err != nil {
			return "400: invalid page number"
		}
	}

	slides := strings.Split(p.Content, "\n## ")
	if currentPage < 1 || currentPage > len(slides) {
		return "404: invalid slide number"
	}

	var out strings.Builder

	// Display current slide
	if currentPage == 1 {
		// Auto-generated intro slide
		out.WriteString(md.H1(p.Title))
		out.WriteString(ufmt.Sprintf("Event: %s\n\n", p.Event))
		out.WriteString(ufmt.Sprintf("Author: %s\n\n", p.Author))
		out.WriteString(ufmt.Sprintf("Date: %s\n\n", p.Date.Format("2006-01-02")))
	} else {
		out.WriteString("## " + slides[currentPage-1])
	}

	// Navigation
	out.WriteString("\n\n---\n\n")
	if currentPage > 1 {
		out.WriteString(ufmt.Sprintf("[← Prev](?slides=%s&page=%d) ", id, currentPage-1))
	}
	out.WriteString(ufmt.Sprintf("| %d/%d |", currentPage, p.NumSlides))
	if currentPage < p.NumSlides {
		out.WriteString(ufmt.Sprintf(" [Next →](?slides=%s&page=%d)", id, currentPage+1))
	}

	return out.String()
}

// Helper functions for sorting and pagination
func getSortField(req *realmpath.Request) string {
	field := req.Query.Get("sort")
	switch field {
	case "date", "uploader", "event":
		return field
	default:
		return "date"
	}
}

func isSortReversed(req *realmpath.Request) bool {
	return req.Query.Get("order") != "asc"
}

func renderSortLink(req *realmpath.Request, field, label string) string {
	currentField := getSortField(req)
	currentOrder := req.Query.Get("order")

	newOrder := "desc"
	if field == currentField && currentOrder != "asc" {
		newOrder = "asc"
	}

	query := req.Query
	query.Set("sort", field)
	query.Set("order", newOrder)

	if field == currentField {
		if newOrder == "asc" {
			label += " ↑"
		} else {
			label += " ↓"
		}
	}

	return ufmt.Sprintf("[%s](?%s)", label, query.Encode())
}

// helper to create local realm links
func localPath(path string, query url.Values) string {
	req := &realmpath.Request{
		Path:  path,
		Query: query,
	}
	return req.String()
}
