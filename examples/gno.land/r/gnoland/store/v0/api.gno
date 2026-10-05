package store

import (
	"chain/runtime"
	"strconv"
	"strings"
	"time"
)

const notFound = `{"version":1,"error":"not_found"}`

// renderAPI serves api/v1/... as compact JSON. path has the "api/" prefix
// removed.
func renderAPI(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] != "v1" {
		return notFound
	}
	h := runtime.ChainHeight()
	switch {
	case path == "v1/home":
		return apiHome(h)
	case path == "v1/build":
		return apiBuild(h)
	case parts[1] == "category" && len(parts) == 4:
		return apiCategory(parts[2], parts[3], h)
	case parts[1] == "list" && len(parts) == 4:
		return apiList(parts[2], parts[3], h)
	}
	return notFound
}

func apiHome(h int64) string {
	var b strings.Builder
	writeHeader(&b, h)
	b.WriteString(`,"pulse":`)
	writePulse(&b, h/epochBlocks)
	writeCategories(&b)
	refs := writeShelves(&b, homeShelves(h))
	if l := picked(h); l != nil {
		writeKey(&b, "pick")
		writeString(&b, l.Slug)
		refs = append(refs, l)
	}
	refs = append(refs, writeActivity(&b)...)
	writeKey(&b, "listings")
	writeListings(&b, refs, h)
	b.WriteByte('}')
	return b.String()
}

func apiBuild(h int64) string {
	var b strings.Builder
	b.WriteString(`{"version":1`)
	refs := writeShelves(&b, buildShelves(h))
	top, fresh := topBuilders()
	b.WriteString(`,"builders":{"top":`)
	writeBuilders(&b, top)
	writeKey(&b, "new")
	writeBuilders(&b, fresh)
	b.WriteByte('}')
	writeKey(&b, "listings")
	writeListings(&b, refs, h)
	b.WriteByte('}')
	return b.String()
}

func apiCategory(key, pageParam string, h int64) string {
	c, ok := findCategory(key)
	if !ok {
		return notFound
	}
	return apiPage(list{Index: categoryMembers(key)}, pageParam, h, func(b *strings.Builder) {
		b.WriteString(`{"version":1,"category":`)
		writeCategory(b, c)
	})
}

func apiList(key, pageParam string, h int64) string {
	l, ok := findList(key, h)
	if !ok {
		return notFound
	}
	return apiPage(l, pageParam, h, func(b *strings.Builder) {
		b.WriteString(`{"version":1,"list":`)
		writeList(b, l)
	})
}

// apiPage answers one page of l, after the head head writes.
func apiPage(l list, pageParam string, h int64, head func(*strings.Builder)) string {
	n, err := strconv.Atoi(pageParam)
	if err != nil || n < 1 {
		return notFound
	}
	members, pages := l.page(n)
	if n > pages {
		return notFound
	}
	var b strings.Builder
	head(&b)
	writeKey(&b, "page")
	writeInt(&b, int64(n))
	writeKey(&b, "pages")
	writeInt(&b, int64(pages))
	writeKey(&b, "listings")
	writeListings(&b, members, h)
	b.WriteByte('}')
	return b.String()
}

func writeList(b *strings.Builder, l list) {
	b.WriteString(`{"key":`)
	writeString(b, l.Key)
	writeKey(b, "title")
	writeString(b, l.Title)
	writeKey(b, "count")
	writeInt(b, int64(l.count()))
	b.WriteByte('}')
}

// writeHeader opens a response with its version, height and block time.
func writeHeader(b *strings.Builder, h int64) {
	b.WriteString(`{"version":1,"height":`)
	writeInt(b, h)
	writeKey(b, "time")
	writeInt(b, time.Now().Unix())
}

func writePulse(b *strings.Builder, epoch int64) {
	b.WriteByte('{')
	for i, w := range []int{7, 30} {
		suffix := "_" + strconv.Itoa(w) + "d"
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(`"listed` + suffix + `":`)
		writeInt(b, int64(pulseListed.sum(epoch, w)))
		writeKey(b, "stars"+suffix)
		writeInt(b, int64(pulseStars.sum(epoch, w)))
	}
	b.WriteByte('}')
}

// writeShelves writes the "shelves" field and returns the listings it
// references.
func writeShelves(b *strings.Builder, shelves []shelf) []*listing {
	var refs []*listing
	b.WriteString(`,"shelves":`)
	writeArray(b, len(shelves), func(i int) {
		s := shelves[i]
		b.WriteString(`{"title":`)
		writeString(b, s.Title)
		if s.Note != "" {
			writeKey(b, "note")
			writeString(b, s.Note)
		}
		if s.Pinned {
			writeKey(b, "pinned")
			b.WriteString("true")
		}
		if s.Empty != "" {
			writeKey(b, "empty")
			writeString(b, s.Empty)
		}
		writeKey(b, "slugs")
		writeSlugs(b, s.Items)
		if len(s.Counts) > 0 {
			writeKey(b, "stars_7d")
			writeArray(b, len(s.Counts), func(j int) { writeInt(b, int64(s.Counts[j])) })
		}
		if s.More.Key != "" {
			writeKey(b, "more")
			writeArray(b, 1, func(int) { writeList(b, s.More) }) // the wire field is an array
		}
		b.WriteByte('}')
		refs = append(refs, shelves[i].Items...)
	})
	return refs
}

// writeActivity writes the "activity" field, the newest activityShown
// store-front events of listings still shown as apps, and returns the
// listings it references.
func writeActivity(b *strings.Builder) []*listing {
	var events []event
	for _, e := range activity.newestFirst() {
		if e.Listing.shownApp() {
			events = append(events, e)
		}
		if len(events) == activityShown {
			break
		}
	}
	refs := make([]*listing, len(events))
	b.WriteString(`,"activity":`)
	writeArray(b, len(events), func(i int) {
		e := events[i]
		b.WriteString(`{"kind":`)
		writeString(b, e.Kind)
		writeKey(b, "slug")
		writeString(b, e.Listing.Slug)
		writeKey(b, "stars")
		writeInt(b, int64(e.Stars))
		b.WriteByte('}')
		refs[i] = e.Listing
	})
	return refs
}

func writeBuilders(b *strings.Builder, rows []*builder) {
	writeArray(b, len(rows), func(i int) {
		b.WriteString(`{"namespace":`)
		writeString(b, rows[i].Namespace)
		writeKey(b, "listings")
		writeInt(b, int64(rows[i].Listings))
		writeKey(b, "stars")
		writeInt(b, int64(rows[i].Stars))
		b.WriteByte('}')
	})
}

// writeCategories writes the "categories" field: the whole taxonomy, in
// order, with the number of visible apps in each.
func writeCategories(b *strings.Builder) {
	b.WriteString(`,"categories":`)
	writeArray(b, len(categories), func(i int) { writeCategory(b, categories[i]) })
}

func writeCategory(b *strings.Builder, c category) {
	b.WriteString(`{"key":`)
	writeString(b, c.Key)
	writeKey(b, "label")
	writeString(b, c.Label)
	writeKey(b, "count")
	writeInt(b, int64(appCount(c.Key)))
	b.WriteByte('}')
}

// writeListings writes each listing once, in first-seen order.
func writeListings(b *strings.Builder, ls []*listing, h int64) {
	ls = unique(ls)
	writeArray(b, len(ls), func(i int) { writeListing(b, ls[i], h) })
}

// unique drops repeated listings, keeping first-seen order. ls holds about
// a hundred references on home, so they are deduplicated by slug, unique
// among the visible listings a response references.
func unique(ls []*listing) []*listing {
	var out []*listing
	seen := map[string]bool{}
	for _, l := range ls {
		if !seen[l.Slug] {
			seen[l.Slug] = true
			out = append(out, l)
		}
	}
	return out
}

func writeSlugs(b *strings.Builder, ls []*listing) {
	writeArray(b, len(ls), func(i int) { writeString(b, ls[i].Slug) })
}

// writeArray writes a JSON array of n items, each written by item(i).
func writeArray(b *strings.Builder, n int, item func(i int)) {
	b.WriteByte('[')
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		item(i)
	}
	b.WriteByte(']')
}

func writeListing(b *strings.Builder, l *listing, h int64) {
	r, _ := l.effective(h)
	b.WriteString(`{"slug":`)
	writeString(b, l.Slug)
	for _, f := range [][2]string{
		{"kind", l.Kind},
		{"path", l.Path},
		{"title", l.Title},
		{"tagline", l.Tagline},
		{"category", l.Category},
	} {
		writeKey(b, f[0])
		writeString(b, f[1])
	}
	writeKey(b, "palette")
	writeInt(b, int64(l.Palette))
	writeKey(b, "stars")
	writeInt(b, int64(l.stars.Size()))
	writeKey(b, "ranked")
	writeInt(b, int64(l.rankedStars.Size()))
	writeKey(b, "earned")
	b.WriteString(strconv.FormatBool(l.earned(h)))
	writeKey(b, "icon")
	writeString(b, r.Icon)
	writeKey(b, "cover")
	writeString(b, r.Cover)
	b.WriteByte('}')
}

// writeKey writes the separator and key of a field that is not the first.
func writeKey(b *strings.Builder, key string) {
	b.WriteString(`,"` + key + `":`)
}

func writeInt(b *strings.Builder, n int64) {
	b.WriteString(strconv.FormatInt(n, 10))
}

// writeString writes s as a JSON string. <, > and & are escaped too, so the
// output is safe to embed in HTML; U+2028 and U+2029 so it is safe in JS.
func writeString(b *strings.Builder, s string) {
	const hex = "0123456789abcdef"
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"' || r == '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		case r < 0x20 || r == '<' || r == '>' || r == '&' || r == 0x2028 || r == 0x2029:
			b.WriteString(`\u`)
			for shift := 12; shift >= 0; shift -= 4 {
				b.WriteByte(hex[(r>>uint(shift))&0xf])
			}
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
}
