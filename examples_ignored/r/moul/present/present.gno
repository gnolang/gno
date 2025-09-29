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

var chainDomain = std.ChainDomain()

type Presentation struct {
	Slug      string
	Title     string
	Event     string
	Author    string
	Uploader  address
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
	// for /view and /slides
	presentations.AddIndex("slug", func(v any) string {
		return v.(*Presentation).Slug
	}, collection.UniqueIndex)

	// for table sorting
	presentations.AddIndex("date", func(v any) string {
		return v.(*Presentation).Date.String()
	}, collection.DefaultIndex)
	presentations.AddIndex("author", func(v any) string {
		return v.(*Presentation).Author
	}, collection.DefaultIndex)
	presentations.AddIndex("title", func(v any) string {
		return v.(*Presentation).Title
	}, collection.DefaultIndex)

	Ownable = ownable.NewWithOrigin()
}

// Render handles the realm's rendering logic
func Render(path string) string {
	req := realmpath.Parse(path)

	// Get slug from path
	slug := req.PathPart(0)

	// List view (home)
	if slug == "" {
		return renderList(req)
	}

	// Slides view
	if req.PathPart(1) == "slides" {
		page := 1
		if pageStr := req.Query.Get("page"); pageStr != "" {
			var err error
			page, err = strconv.Atoi(pageStr)
			if err != nil {
				return "400: invalid page number"
			}
		}
		return renderSlides(slug, page)
	}

	// Regular view
	return renderView(slug)
}

// Set adds or updates a presentation
func Set(slug, title, event, author, date, content string) string {
	Ownable.AssertOwnedByPrevious()

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
	numSlides++ // Count thank you slide

	p := &Presentation{
		Slug:      slug,
		Title:     title,
		Event:     event,
		Author:    author,
		Uploader:  std.PreviousRealm().Address(),
		Date:      parsedDate,
		Content:   content,
		EditDate:  time.Now(),
		NumSlides: numSlides,
	}

	presentations.Set(p)
	return "presentation saved successfully"
}

// Delete removes a presentation
func Delete(slug string) string {
	Ownable.AssertOwnedByPrevious()

	entry := presentations.GetFirst("slug", slug)
	if entry == nil {
		return "404: presentation not found"
	}

	// XXX: consider this:
	// if entry.Obj.(*Presentation).Uploader != std.PreviousRealm().Address() {
	// 	return "401: unauthorized - only the uploader can delete their presentations"
	// }

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
	index := presentations.GetIndex(getSortField(req))
	pgr := pager.NewPager(index, 10, isSortReversed(req))

	// Get current page
	page := pgr.MustGetPageByPath(req.String())

	// Create table
	dateColumn := renderSortLink(req, "date", "Date")
	titleColumn := renderSortLink(req, "title", "Title")
	authorColumn := renderSortLink(req, "author", "Author")
	table := mdtable.Table{
		Headers: []string{dateColumn, titleColumn, "Event", authorColumn, "Slides"},
	}

	// Add rows from current page
	for _, item := range page.Items {
		// Get the actual presentation using the ID from the index
		// XXX: improve p/moul/collection to make this more convenient.
		//   - no need to make per-id lookup.
		//   - transparently support multi-values.
		//   - integrate a sortable pager?
		var ids []string
		if ids_, ok := item.Value.([]string); ok {
			ids = ids_
		} else if id, ok := item.Value.(string); ok {
			ids = []string{id}
		}

		for _, id := range ids {
			entry := presentations.GetFirst(collection.IDIndex, id)
			if entry == nil {
				continue
			}
			p := entry.Obj.(*Presentation)

			table.Append([]string{
				p.Date.Format("2006-01-02"),
				md.Link(p.Title, localPath(p.Slug, nil)),
				p.Event,
				p.Author,
				ufmt.Sprintf("%d", p.NumSlides),
			})
		}
	}

	out.WriteString(table.String())
	out.WriteString(page.Picker(req.String()))
	return out.String()
}

func (p *Presentation) FirstSlide() string {
	var out strings.Builder
	out.WriteString(md.H1(p.Title))
	out.WriteString(md.Paragraph(md.Bold(p.Event) + ", " + p.Date.Format("2 Jan 2006")))
	out.WriteString(md.Paragraph("by " + md.Bold(p.Author))) // XXX: link to u/?
	return out.String()
}

func (p *Presentation) LastSlide() string {
	var out strings.Builder
	out.WriteString(md.H1(p.Title))
	out.WriteString(md.H2("Thank You!"))
	out.WriteString(md.Paragraph(p.Author))
	fullPath := "https://" + chainDomain + localPath(p.Slug, nil)
	out.WriteString(md.Paragraph("🔗 " + md.Link(fullPath, fullPath)))
	// XXX: QRCode
	return out.String()
}

func renderView(slug string) string {
	if slug == "" {
		return "400: missing presentation slug"
	}

	entry := presentations.GetFirst("slug", slug)
	if entry == nil {
		return "404: presentation not found"
	}

	p := entry.Obj.(*Presentation)
	var out strings.Builder

	// Header using FirstSlide helper
	out.WriteString(p.FirstSlide())

	// Slide mode link
	out.WriteString(md.Link("View as slides", localPath(p.Slug+"/slides", nil)) + "\n\n")
	out.WriteString(md.HorizontalRule())
	out.WriteString(md.Paragraph(p.Content))

	// Metadata footer
	out.WriteString(md.HorizontalRule())
	out.WriteString(ufmt.Sprintf("Last edited: %s\n\n", p.EditDate.Format("2006-01-02 15:04:05")))
	out.WriteString(ufmt.Sprintf("Uploader: `%s`\n\n", p.Uploader))
	out.WriteString(ufmt.Sprintf("Number of slides: %d\n\n", p.NumSlides))

	// Admin actions
	// XXX: consider a dynamic toggle for admin actions
	editLink := txlink.Call("Set",
		"slug", p.Slug,
		"title", p.Title,
		"author", p.Author,
		"event", p.Event,
		"date", p.Date.Format("2006-01-02"),
	)
	deleteLink := txlink.Call("Delete", "slug", p.Slug)
	out.WriteString(md.Paragraph(md.Link("Edit", editLink) + " | " + md.Link("Delete", deleteLink)))

	return out.String()
}

// renderSlidesNavigation returns the navigation bar for slides
func renderSlidesNavigation(slug string, currentPage, totalSlides int) string {
	var out strings.Builder
	if currentPage > 1 {
		prevLink := localPath(slug+"/slides", url.Values{"page": {ufmt.Sprintf("%d", currentPage-1)}})
		out.WriteString(md.Link("← Prev", prevLink) + " ")
	}
	out.WriteString(ufmt.Sprintf("| %d/%d |", currentPage, totalSlides))
	if currentPage < totalSlides {
		nextLink := localPath(slug+"/slides", url.Values{"page": {ufmt.Sprintf("%d", currentPage+1)}})
		out.WriteString(" " + md.Link("Next →", nextLink))
	}
	return md.Paragraph(out.String())
}

func renderSlides(slug string, currentPage int) string {
	if slug == "" {
		return "400: missing presentation ID"
	}

	entry := presentations.GetFirst("slug", slug)
	if entry == nil {
		return "404: presentation not found"
	}

	p := entry.Obj.(*Presentation)
	slides := strings.Split("\n"+p.Content, "\n## ")
	if currentPage < 1 || currentPage > p.NumSlides {
		return "404: invalid slide number"
	}

	var out strings.Builder

	// Display current slide
	if currentPage == 1 {
		out.WriteString(p.FirstSlide())
	} else if currentPage == p.NumSlides {
		out.WriteString(p.LastSlide())
	} else {
		out.WriteString(md.H1(p.Title))
		out.WriteString("## " + slides[currentPage-1] + "\n\n")
	}

	out.WriteString(renderSlidesNavigation(slug, currentPage, p.NumSlides))
	return out.String()
}

// Helper functions for sorting and pagination
func getSortField(req *realmpath.Request) string {
	field := req.Query.Get("sort")
	switch field {
	case "date", "slug", "author", "title":
		return field
	}
	return "date"
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

	return md.Link(label, "?"+query.Encode())
}

// helper to create local realm links
func localPath(path string, query url.Values) string {
	req := &realmpath.Request{
		Path:  path,
		Query: query,
	}
	return req.String()
}
