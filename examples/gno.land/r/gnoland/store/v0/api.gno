package store

import (
	"chain/runtime"
	"strconv"
	"strings"
)

const notFound = `{"version":1,"error":"not_found"}`

// itemRoom is more than the longest array item but a nested array: a
// listing, with a 256-byte path, 40 and 80 runes of title and tagline, each
// at most 6 bytes once escaped, two content ids and the fixed fields.
const itemRoom = 2048

// renderAPI serves api/v1/... as compact JSON. path has the "api/" prefix
// removed.
func renderAPI(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) < 2 || parts[0] != "v1" {
		return notFound
	}
	now := blockTime()
	switch {
	case path == "v1/home":
		return apiHome(now)
	case path == "v1/build":
		return apiBuild(now)
	case parts[1] == "category" && len(parts) == 4:
		return apiCategory(parts[2], parts[3], now)
	case parts[1] == "list" && len(parts) == 4:
		return apiList(parts[2], parts[3], now)
	}
	return notFound
}

func apiHome(now int64) string {
	var b strings.Builder
	writeHeader(&b, now)
	b.WriteString(`,"pulse":`)
	writePulse(&b, now/epochLength)
	writeCategories(&b)
	refs := writeShelves(&b, homeShelves(now))
	if l := picked(now); l != nil {
		writeKey(&b, "pick")
		writeString(&b, l.Slug)
		refs = append(refs, l)
	}
	writeKey(&b, "listings")
	writeListings(&b, refs, now)
	b.WriteByte('}')
	return b.String()
}

func apiBuild(now int64) string {
	var b strings.Builder
	b.WriteString(`{"version":1`)
	writeMoved(&b)
	refs := writeShelves(&b, buildShelves(now))
	top, fresh := topBuilders()
	b.WriteString(`,"builders":{"top":`)
	writeBuilders(&b, top)
	writeKey(&b, "new")
	writeBuilders(&b, fresh)
	b.WriteByte('}')
	writeKey(&b, "listings")
	writeListings(&b, refs, now)
	b.WriteByte('}')
	return b.String()
}

func apiCategory(key, pageParam string, now int64) string {
	c, ok := findCategory(key)
	if !ok {
		return notFound
	}
	return apiPage(list{Index: categoryMembers(key)}, pageParam, now, func(b *strings.Builder) {
		b.WriteString(`{"version":1,"category":`)
		writeCategory(b, c)
	})
}

func apiList(key, pageParam string, now int64) string {
	l, ok := findList(key, now)
	if !ok {
		return notFound
	}
	return apiPage(l, pageParam, now, func(b *strings.Builder) {
		b.WriteString(`{"version":1,"list":`)
		writeList(b, l)
	})
}

// apiPage answers one page of l, after the head head writes.
func apiPage(l list, pageParam string, now int64, head func(*strings.Builder)) string {
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
	writeListings(&b, members, now)
	b.WriteByte('}')
	return b.String()
}

func writeList(b *strings.Builder, l list) {
	b.WriteString(`{"key":`)
	writeToken(b, l.Key)
	writeKey(b, "count")
	writeInt(b, int64(l.count()))
	b.WriteByte('}')
}

// writeHeader opens a response with its version, the block height and the
// block time now.
func writeHeader(b *strings.Builder, now int64) {
	b.WriteString(`{"version":1,"height":`)
	writeInt(b, runtime.ChainHeight())
	writeKey(b, "time")
	writeInt(b, now)
	writeMoved(b)
}

// writeMoved names the store's successor, if any, on the pages that open a
// lens: home and build.
func writeMoved(b *strings.Builder) {
	if successor != "" {
		writeKey(b, "moved_to")
		writeString(b, successor)
	}
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
		b.WriteString(`{"key":`)
		writeToken(b, s.Key)
		writeKey(b, "slugs")
		writeSlugs(b, s.Items)
		if s.More.Key != "" {
			writeKey(b, "more")
			writeArray(b, 1, func(int) { writeList(b, s.More) }) // the wire field is an array
		}
		b.WriteByte('}')
		refs = append(refs, shelves[i].Items...)
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
	writeKey(b, "confirmed")
	writeInt(b, int64(confirmedCount(c.Key)))
	b.WriteByte('}')
}

// writeListings writes each listing once, in first-seen order.
func writeListings(b *strings.Builder, ls []*listing, now int64) {
	ls = unique(ls)
	writeArray(b, len(ls), func(i int) { writeListing(b, ls[i], now) })
}

// unique drops repeated listings, keeping first-seen order. ls holds about
// a hundred references on home, so they are deduplicated by slug, unique
// among visible listings.
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
	writeArray(b, len(ls), func(i int) { writeToken(b, ls[i].Slug) })
}

// writeArray writes a JSON array of n items, each written by item(i). It
// takes room for each by doubling: a builder that grows as it is written to
// is copied whole at every write (gno's append does not over-allocate),
// which costs gas in the square of the response.
func writeArray(b *strings.Builder, n int, item func(i int)) {
	b.WriteByte('[')
	for i := 0; i < n; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		b.Grow(itemRoom)
		item(i)
	}
	b.WriteByte(']')
}

func writeListing(b *strings.Builder, l *listing, now int64) {
	r, pending := l.effective(now)
	b.WriteString(`{"slug":`)
	writeToken(b, l.Slug)
	writeKey(b, "kind")
	writeToken(b, l.Kind)
	writeKey(b, "path")
	writeToken(b, l.Path)
	writeKey(b, "title")
	writeText(b, l.Title)
	writeKey(b, "tagline")
	writeText(b, l.Tagline)
	writeKey(b, "category")
	writeToken(b, l.Category)
	writeKey(b, "palette")
	writeInt(b, int64(l.Palette))
	writeKey(b, "stars")
	writeInt(b, int64(l.starCount()))
	writeKey(b, "ranked")
	writeInt(b, int64(l.rankedCount()))
	writeKey(b, "stars_7d")
	writeInt(b, int64(l.stars7d(now)))
	writeKey(b, "fifth_ranked")
	writeInt(b, l.fifthRanked())
	writeKey(b, "created")
	writeInt(b, l.Created)
	writeKey(b, "updated")
	writeInt(b, l.Updated)
	writeKey(b, "pending")
	b.WriteString(strconv.FormatBool(pending))
	writeKey(b, "core")
	b.WriteString(strconv.FormatBool(l.core))
	writeKey(b, "confirmed")
	b.WriteString(strconv.FormatBool(l.Confirmed))
	writeKey(b, "icon")
	writeToken(b, r.Icon)
	writeKey(b, "cover")
	writeToken(b, r.Cover)
	b.WriteByte('}')
}

// writeKey writes the separator and key of a field that is not the first.
func writeKey(b *strings.Builder, key string) {
	b.WriteString(`,"` + key + `":`)
}

func writeInt(b *strings.Builder, n int64) {
	b.WriteString(strconv.FormatInt(n, 10))
}

// writeToken writes, as a JSON string, a value whose grammar holds nothing
// to escape: a slug, a kind, a path, a category key or a content id, all
// checked when stored (validate.gno). One write, where writeString costs gas
// for every rune.
func writeToken(b *strings.Builder, s string) {
	b.WriteByte('"')
	b.WriteString(s)
	b.WriteByte('"')
}

// writeText writes a title or a tagline as a JSON string. validText left
// them no control character and no line separator, so most need no
// escaping and go in one write.
func writeText(b *strings.Builder, s string) {
	for _, r := range s { // cheaper in gas than strings.ContainsAny
		switch r {
		case '"', '\\', '<', '>', '&', '[', ']', '$', '@':
			writeString(b, s)
			return
		}
	}
	writeToken(b, s)
}

// writeString writes s as a JSON string. <, > and & are escaped too, so the
// output is safe to embed in HTML; U+2028 and U+2029 so it is safe in JS;
// [, ], $, @ and \ so gnoweb, which renders api/ as markdown, makes no
// link, mention or formula of it: a backslash written as two would open a
// formula before a ( or a [.
func writeString(b *strings.Builder, s string) {
	const hex = "0123456789abcdef"
	b.WriteByte('"')
	for _, r := range s {
		switch {
		case r == '"':
			b.WriteString(`\"`)
		case r < 0x20 || r == '\\' || r == '<' || r == '>' || r == '&' || r == '[' || r == ']' || r == '$' || r == '@' || r == 0x2028 || r == 0x2029:
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
