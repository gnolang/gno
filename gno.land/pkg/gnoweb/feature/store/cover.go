package store

import (
	"fmt"
	"hash/fnv"
	"html/template"
	"math/rand/v2"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gnolang/gno/gno.land/pkg/gnoweb/components"
)

// The generated visuals are SVG written entirely by this file from a slug
// and a palette index: no listing text reaches the markup except
// initials, which are escaped. Colours are presentation attributes, not
// style attributes, so gnoweb's CSP (style-src 'self') allows them.

// swatch is one palette entry: a deep background and a light accent.
type swatch struct{ bg, fg string }

var palette = [paletteSize]swatch{
	{"#1d5c4a", "#7cc4ad"}, // green (brand family)
	{"#1f4a63", "#6cbcec"}, // blue
	{"#43215f", "#b48bd1"}, // purple
	{"#5c4a06", "#f2cf6b"}, // yellow
	{"#5e1a10", "#f19a86"}, // red
	{"#0f4b50", "#6cc5c2"}, // teal
	{"#2a333e", "#9fb2c6"}, // slate
	{"#5e3512", "#f0ad6b"}, // orange
	{"#5c1f47", "#e79bc8"}, // pink
	{"#3f4d17", "#bcd174"}, // olive
	{"#272c63", "#98a0ec"}, // indigo
	{"#29292d", "#b4b4bc"}, // graphite
}

// motif draws the foreground of a cover into b, using r for variation.
type motif func(b *strings.Builder, r *rand.Rand, fg string)

// motifs are the cover families. The slug picks one, so a category page,
// whose apps share nothing else, still varies.
var motifs = []motif{waves, network, grid, rings, blobs, bars, dots}

const coverW, coverH = 320, 180

// cover returns a deterministic 16:9 artwork for a listing.
func cover(l *listing) template.HTML {
	sw := palette[l.Palette]
	r := newRNG(l.Slug)
	draw := motifs[r.IntN(len(motifs))]

	// No gradient: an <svg> id would collide when a card appears twice on
	// a page. A soft light disc, placed by the slug, gives the same depth
	// without one.
	var b strings.Builder
	fmt.Fprintf(&b, `<svg class="cover" viewBox="0 0 %d %d" preserveAspectRatio="xMidYMid slice" aria-hidden="true" focusable="false">`, coverW, coverH)
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="%s"/><circle cx="%d" cy="%d" r="%d" fill="%s" fill-opacity=".1"/>`,
		coverW, coverH, sw.bg, r.IntN(coverW+1), r.IntN(coverH/2+1), coverH*3/4+r.IntN(coverH/2), sw.fg)
	draw(&b, r, sw.fg)
	b.WriteString(`</svg>`)
	return template.HTML(b.String()) //nolint:gosec // built from constants and numbers only
}

// icon returns the fallback app icon: initials on the palette swatch.
func icon(l *listing) template.HTML {
	sw := palette[l.Palette]
	return template.HTML(fmt.Sprintf( //nolint:gosec // initials are escaped
		`<svg class="icon" viewBox="0 0 40 40" aria-hidden="true" focusable="false"><rect width="40" height="40" rx="8" fill="%s"/><text x="20" y="25.5" text-anchor="middle" font-size="15" font-weight="600" fill="%s">%s</text></svg>`,
		sw.bg, sw.fg, template.HTMLEscapeString(initials(l.Title))))
}

// initials returns a monogram of two letters or digits, never empty: the
// first of each of the first two words ("DAO Forge" → "DF"); for a single
// word, of its first two camel-case or punctuated parts ("GnoSwap" → "GS",
// "GovDAO" → "GD", "Gno.land" → "GL"), or its first two letters when it has
// no such parts ("Gnotify" → "GN"). A title with no letter or digit shows
// its first visible rune ("★★★" → "★"), and noGlyph if it has none.
func initials(name string) string {
	parts := slices.DeleteFunc(strings.Fields(name), func(w string) bool { return strings.IndexFunc(w, isAlnum) < 0 })
	if len(parts) == 1 {
		parts = wordParts(parts[0])
	}
	var out []rune
	switch len(parts) {
	case 0:
		if i := strings.IndexFunc(name, components.IsVisibleRune); i >= 0 {
			r, _ := utf8.DecodeRuneInString(name[i:])
			return string(unicode.ToUpper(r))
		}
		return noGlyph
	case 1:
		for _, r := range parts[0] {
			if out = append(out, unicode.ToUpper(r)); len(out) == 2 {
				break
			}
		}
	default:
		for _, p := range parts[:2] {
			r, _ := utf8.DecodeRuneInString(p)
			out = append(out, unicode.ToUpper(r))
		}
	}
	return string(out)
}

// noGlyph stands in for initials when a title has nothing to draw them from.
const noGlyph = "•"

// wordParts splits a word where a new part starts: at an upper-case letter
// after a lower-case one or before the last capital of a run ("GovDAO",
// "DAOForge"), and at anything that is not a letter or a digit.
func wordParts(w string) []string {
	rs := []rune(w)
	var parts []string
	start := -1
	for i, r := range rs {
		if !isAlnum(r) {
			if start >= 0 {
				parts = append(parts, string(rs[start:i]))
			}
			start = -1
			continue
		}
		boundary := i > 0 && start >= 0 && unicode.IsUpper(r) &&
			(unicode.IsLower(rs[i-1]) || unicode.IsUpper(rs[i-1]) && i+1 < len(rs) && unicode.IsLower(rs[i+1]))
		if boundary {
			parts = append(parts, string(rs[start:i]))
			start = i
		} else if start < 0 {
			start = i
		}
	}
	if start >= 0 {
		parts = append(parts, string(rs[start:]))
	}
	return parts
}

func isAlnum(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }

func waves(b *strings.Builder, r *rand.Rand, fg string) {
	for i := range 5 {
		y := 40 + i*26 + r.IntN(10)
		amp := 8 + r.IntN(18)
		fmt.Fprintf(b, `<path d="M0 %d C80 %d 160 %d 240 %d S320 %d 320 %d" fill="none" stroke="%s" stroke-width="2" stroke-opacity="%.2f"/>`,
			y, y-amp, y+amp, y, y-amp, y, fg, 0.25+0.12*float64(i))
	}
}

func network(b *strings.Builder, r *rand.Rand, fg string) {
	const n = 9
	var xs, ys [n]int
	for i := range n {
		xs[i], ys[i] = 30+r.IntN(coverW-60), 25+r.IntN(coverH-50)
	}
	for i := range n {
		j := (i + 1 + r.IntN(n-1)) % n
		fmt.Fprintf(b, `<line x1="%d" y1="%d" x2="%d" y2="%d" stroke="%s" stroke-opacity=".4" stroke-width="1.5"/>`, xs[i], ys[i], xs[j], ys[j], fg)
	}
	for i := range n {
		fmt.Fprintf(b, `<circle cx="%d" cy="%d" r="%d" fill="%s" fill-opacity=".85"/>`, xs[i], ys[i], 3+r.IntN(6), fg)
	}
}

func grid(b *strings.Builder, r *rand.Rand, fg string) {
	const cell = 30
	for x := 0; x < coverW; x += cell {
		for y := 0; y < coverH; y += cell {
			if r.IntN(4) == 0 {
				fmt.Fprintf(b, `<rect x="%d" y="%d" width="%d" height="%d" rx="3" fill="%s" fill-opacity="%.2f"/>`,
					x+2, y+2, cell-4, cell-4, fg, 0.2+0.15*float64(r.IntN(5)))
			}
		}
	}
}

func rings(b *strings.Builder, r *rand.Rand, fg string) {
	cx, cy := 200+r.IntN(80), 60+r.IntN(60)
	for i := range 7 {
		fmt.Fprintf(b, `<circle cx="%d" cy="%d" r="%d" fill="none" stroke="%s" stroke-width="2" stroke-opacity="%.2f"/>`,
			cx, cy, 18+i*22, fg, 0.7-0.08*float64(i))
	}
}

func blobs(b *strings.Builder, r *rand.Rand, fg string) {
	for range 6 {
		fmt.Fprintf(b, `<circle cx="%d" cy="%d" r="%d" fill="%s" fill-opacity=".28"/>`,
			r.IntN(coverW), r.IntN(coverH), 24+r.IntN(46), fg)
	}
}

func bars(b *strings.Builder, r *rand.Rand, fg string) {
	for i := range 9 {
		w := 60 + r.IntN(180)
		fmt.Fprintf(b, `<rect x="%d" y="%d" width="%d" height="8" rx="4" fill="%s" fill-opacity="%.2f"/>`,
			28+r.IntN(3)*16, 22+i*16, w, fg, 0.3+0.07*float64(r.IntN(6)))
	}
}

func dots(b *strings.Builder, r *rand.Rand, fg string) {
	for x := 18; x < coverW; x += 36 {
		for y := 18; y < coverH; y += 36 {
			fmt.Fprintf(b, `<circle cx="%d" cy="%d" r="%d" fill="%s" fill-opacity=".45"/>`, x, y, 1+r.IntN(4), fg)
		}
	}
}

// newRNG draws from PCG seeded by the slug's FNV hash: the same slug draws
// the same cover on every gnoweb instance.
func newRNG(seed string) *rand.Rand {
	h := fnv.New64a()
	h.Write([]byte(seed))
	return rand.New(rand.NewPCG(h.Sum64(), 0)) //nolint:gosec // art, not security: determinism is the point
}
