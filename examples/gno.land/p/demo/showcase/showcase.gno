package showcase

import (
	"strconv"
	"strings"
)

// Version is the published version of the showcase API.
const Version = "1.0.0"

// Lifecycle states an [Item] can occupy.
const (
	// StatusDraft is the initial, unpublished state.
	StatusDraft = "draft"
	// StatusActive marks a live, published item.
	StatusActive = "active"
	// StatusArchived marks a retired item.
	StatusArchived = "archived"
)

// DefaultTags are attached to a freshly created [Item].
var DefaultTags = Tags{"new"}

// MaxTitleLen bounds an item title in bytes; [Clean] truncates longer titles.
var MaxTitleLen = 64

// Catalog maps item names to records. As a defined map type it renders with
// the map glyph in the package index.
type Catalog map[string]*Item

// Tags is an ordered set of labels. As a defined slice type it renders with
// the slice glyph.
type Tags []string

// Hook runs after a catalog mutation. As a defined function type it renders
// with the func glyph.
type Hook func(name string, it *Item)

// ItemRef is a pointer alias to [Item], rendering with the pointer glyph. It is
// contrived and exists only to exercise that path.
type ItemRef *Item

// Name is a defined string type. Backed by a primitive, it renders with the
// generic type glyph.
type Name string

// Item is a single catalog record.
//
// As a struct it renders with the struct glyph and groups its methods —
// [Item.String], [Item.Publish] — beneath it in the index.
type Item struct {
	Title  string // human-readable label
	Price  int64  // amount, in the smallest currency unit
	Status string // one of the Status* constants
	Labels Tags   // optional, free-form labels
}

// Storer abstracts a backing store for items. As an interface it renders with
// the interface glyph.
type Storer interface {
	// Get returns the item stored under name, or nil when absent.
	Get(name string) *Item
	// Set stores it under name, replacing any previous value.
	Set(name string, it *Item)
}

// New returns a draft [Item] with [DefaultTags] applied and its title cleaned.
//
// Example:
//
//	it := showcase.New("  Widget  ", 1500)
//	it.Publish()
func New(title string, price int64) *Item {
	return &Item{
		Title:  Clean(title),
		Price:  price,
		Status: StatusDraft,
		Labels: DefaultTags,
	}
}

// Clean trims surrounding spaces from title and caps it at [MaxTitleLen] bytes.
func Clean(title string) string {
	title = strings.TrimSpace(title)
	if len(title) > MaxTitleLen {
		title = title[:MaxTitleLen]
	}
	return title
}

// Summarize renders one Markdown bullet per catalog entry. The output is meant
// to be embedded in a realm's Render result.
func Summarize(c Catalog) string {
	lines := make([]string, 0, len(c))
	for name, it := range c {
		lines = append(lines, "- "+name+": "+it.String())
	}
	return strings.Join(lines, "\n")
}

// Total sums the price of every item in the catalog.
func (c Catalog) Total() int64 {
	var sum int64
	for _, it := range c {
		sum += it.Price
	}
	return sum
}

// String renders the item as "Title (Status)", satisfying the Stringer contract.
func (it *Item) String() string {
	return it.Title + " (" + it.Status + ")"
}

// PriceString returns the price formatted in base 10.
func (it *Item) PriceString() string {
	return strconv.FormatInt(it.Price, 10)
}

// Publish flips the item to [StatusActive] and returns it, enabling chaining.
func (it *Item) Publish() *Item {
	it.Status = StatusActive
	return it
}

// Has reports whether the tag set contains label.
func (t Tags) Has(label string) bool {
	for _, l := range t {
		if l == label {
			return true
		}
	}
	return false
}
