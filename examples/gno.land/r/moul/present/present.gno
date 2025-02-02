package present

import (
	"std"
	"strings"
	"time"

	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/ownable"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/collection"
	"gno.land/p/moul/md"
	"gno.land/p/moul/mdtable"
	"gno.land/p/moul/realmpath"
)

type Presentation struct {
	ID        string
	Title     string
	Event     string
	Author    std.Address
	Date      time.Time
	Content   string
	EditDate  time.Time
	NumSlides int
}

var (
	presentations *collection.Collection
	owner         *ownable.Ownable
)

func init() {
	presentations = collection.New()
	presentations.AddIndex("id", func(v interface{}) string {
		return v.(*Presentation).ID
	}, collection.UniqueIndex)
	presentations.AddIndex("date", func(v interface{}) string {
		return v.(*Presentation).Date.String()
	}, collection.DefaultIndex)
	presentations.AddIndex("author", func(v interface{}) string {
		return v.(*Presentation).Author.String()
	}, collection.DefaultIndex)
	presentations.AddIndex("event", func(v interface{}) string {
		return v.(*Presentation).Event
	}, collection.DefaultIndex)

	owner = ownable.New()
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
	case "set":
		return renderSet(req)
	case "delete":
		return renderDelete(req)
	default:
		return "404: not found"
	}
}

func renderList(req *realmpath.Request) string {
	var out strings.Builder
	out.WriteString(md.H1("Presentations"))

	// Setup pager
	p := pager.NewPager(presentations.GetIndex(getSortField(req)), 10, isSortReversed(req))

	// Get current page
	page := p.MustGetPageByPath(req.Query.Encode())

	// Create table
	table := mdtable.Table{
		Headers: []string{"Title", "Author", "Event", "Date", "Slides"},
	}

	// Add sort options
	out.WriteString("Sort by: ")
	out.WriteString(renderSortLink(req, "date", "Date"))
	out.WriteString(" | ")
	out.WriteString(renderSortLink(req, "author", "Author"))
	out.WriteString(" | ")
	out.WriteString(renderSortLink(req, "event", "Event"))
	out.WriteString("\n\n")

	// Add rows from current page
	for _, item := range page.Items {
		p := item.Value.(*Presentation)
		table.Append([]string{
			ufmt.Sprintf("[%s](?view=%s)", p.Title, p.ID),
			p.Author.String(),
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
	out.WriteString(ufmt.Sprintf("[View as slides](?slides=%s)\n\n", p.ID))
	out.WriteString(md.HorizontalRule())

	// Content
	out.WriteString(p.Content)

	// Metadata footer
	out.WriteString("\n\n")
	out.WriteString(md.HorizontalRule())
	out.WriteString(ufmt.Sprintf("Last edited: %s\n\n", p.EditDate.Format("2006-01-02 15:04:05")))
	out.WriteString(ufmt.Sprintf("ID: `%s`\n\n", p.ID))
	out.WriteString(ufmt.Sprintf("Number of slides: %d\n\n", p.NumSlides))

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
		currentPage = std.MustAtoi(page)
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

// Set adds or updates a presentation
func Set(id, title, event, date, content string) string {
	owner.AssertCallerIsOwner()

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
		Author:    std.PrevRealm().Addr(),
		Date:      parsedDate,
		Content:   content,
		EditDate:  time.Now(),
		NumSlides: numSlides,
	}

	presentations.Set(p)
	return "presentation saved successfully"
}

func renderSet(req *realmpath.Request) string {
	if !owner.CallerIsOwner() {
		return "401: unauthorized"
	}

	id := req.Query.Get("set")
	if id == "" {
		return "400: missing presentation ID"
	}

	entry := presentations.GetFirst("id", id)
	if entry == nil {
		return "404: presentation not found"
	}

	p := entry.Obj.(*Presentation)
	return ufmt.Sprintf("Current presentation:\nID: %s\nTitle: %s\nEvent: %s\nDate: %s\nContent:\n%s",
		p.ID, p.Title, p.Event, p.Date.Format("2006-01-02"), p.Content)
}

// Delete removes a presentation
func Delete(id string) string {
	if !owner.CallerIsOwner() {
		return "401: unauthorized"
	}

	entry := presentations.GetFirst("id", id)
	if entry == nil {
		return "404: presentation not found"
	}

	presentations.Delete(id)
	return "presentation deleted successfully"
}

func renderDelete(req *realmpath.Request) string {
	if !owner.CallerIsOwner() {
		return "401: unauthorized"
	}

	id := req.Query.Get("delete")
	if id == "" {
		return "400: missing presentation ID"
	}

	return Delete(id)
}

// Helper functions for sorting and pagination
func getSortField(req *realmpath.Request) string {
	field := req.Query.Get("sort")
	switch field {
	case "date", "author", "event":
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
