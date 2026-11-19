package store

// Typed reads, for other realms and for clients (vm/qeval_json). They are
// the way out of this version, which can never be redeployed: a store that
// replaces it pages through Listings and StarsOf to take over what this one
// holds. Every read returns copies, and readPage rows at most.

const readPage = 100

// Listing is a listing as other realms and clients read it.
type Listing struct {
	Path      string
	Slug      string
	Kind      string // app, service or package
	Title     string
	Tagline   string
	Category  string
	Palette   int
	Created   int64 // block time, unix seconds
	Updated   int64
	Stars     int // every star
	Ranked    int // the stars that rank it (see Star)
	Core      bool
	Confirmed bool
	Hidden    bool
	Icon      string // in effect now
	Cover     string
}

// StarRecord is one star of a listing.
type StarRecord struct {
	Address address
	At      int64 // block time, unix seconds
	Ranked  bool
}

func (l *listing) export(now int64) Listing {
	r, _ := l.effective(now)
	return Listing{
		Path: l.Path, Slug: l.Slug, Kind: l.Kind, Title: l.Title, Tagline: l.Tagline,
		Category: l.Category, Palette: l.Palette, Created: l.Created, Updated: l.Updated,
		Stars: l.starCount(), Ranked: l.rankedCount(),
		Core: l.core, Confirmed: l.Confirmed, Hidden: l.hidden,
		Icon: r.Icon, Cover: r.Cover,
	}
}

// rows bounds a read's limit to [0, readPage].
func rows(limit int) int {
	switch {
	case limit < 0:
		return 0
	case limit > readPage:
		return readPage
	}
	return limit
}

// ListingCount is the number of listings, hidden ones included.
func ListingCount() int {
	return byPath.Size()
}

// Listings returns at most limit listings from offset, ordered by path,
// hidden ones included.
func Listings(offset, limit int) []Listing {
	var out []Listing
	now := blockTime()
	byPath.IterateByOffset(offset, rows(limit), func(_ string, v any) bool {
		out = append(out, v.(*listing).export(now))
		return false
	})
	return out
}

// ListingOf returns the listing of path, hidden or not.
func ListingOf(path string) (Listing, bool) {
	l := lookup(&byPath, path)
	if l == nil {
		return Listing{}, false
	}
	return l.export(blockTime()), true
}

// StarsOf returns at most limit stars of path's listing from offset,
// ordered by address.
func StarsOf(path string, offset, limit int) []StarRecord {
	l := lookup(&byPath, path)
	if l == nil || l.book == nil {
		return nil
	}
	var out []StarRecord
	l.book.all.IterateByOffset(offset, rows(limit), func(addr string, _ any) bool {
		s := l.starOf(addr)
		out = append(out, StarRecord{Address: address(addr), At: s.At, Ranked: s.Ranked})
		return false
	})
	return out
}

// Successor is the realm path of the store GovDAO named to replace this
// one, or "".
func Successor() string {
	return successor
}

// Paused reports whether GovDAO paused new listings and listing updates.
func Paused() bool {
	return paused
}
