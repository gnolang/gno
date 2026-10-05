package store

import (
	"chain/runtime"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	maxSlug    = 40
	minSlug    = 3
	maxTitle   = 40 // runes
	maxTagline = 80 // runes
	palettes   = 12
	minCID     = 46
	maxCID     = 64
)

// Reasons a listing is refused for that are not about its own fields.
const (
	reasonSlugTaken     = "slug_taken"
	reasonTitleTaken    = "title_taken"
	reasonNamespaceFull = "namespace_full"
)

// validateListing returns a short reason why l cannot be listed, or "".
// Reasons are emitted in events and errors, so they stay stable identifiers.
func validateListing(l *listing) string {
	_, knownCategory := findCategory(l.Category)
	switch {
	case l.Namespace == "":
		return "bad_path"
	case !isKind(l.Kind) || (l.Kind == kindPackage) != isPackagePath(l.Path):
		return "bad_kind"
	case !validSlug(l.Slug):
		return "bad_slug"
	case !validText(l.Title, maxTitle):
		return "bad_title"
	case !validText(l.Tagline, maxTagline):
		return "bad_tagline"
	case !knownCategory:
		return "bad_category"
	case l.Palette < 0 || l.Palette >= palettes:
		return "bad_palette"
	}
	return ""
}

// validateRich returns a short reason why r cannot be submitted, or "".
func validateRich(r rich) string {
	switch {
	case !validAsset(r.Icon):
		return "bad_icon"
	case !validAsset(r.Cover):
		return "bad_cover"
	}
	return ""
}

// namespaceOf returns the namespace of a realm or package path on this
// chain ("<domain>/{r,p}/<namespace>/<segment>..."), or "" if path is not
// one. Claim takes paths from users, so every segment is checked:
// [a-z][a-z0-9_-]*.
func namespaceOf(path string) string {
	rest, ok := strings.CutPrefix(path, runtime.ChainDomain()+"/r/")
	if !ok {
		rest, ok = strings.CutPrefix(path, runtime.ChainDomain()+"/p/")
	}
	segments := strings.Split(rest, "/")
	if !ok || len(segments) < 2 {
		return ""
	}
	for _, seg := range segments {
		if !validSegment(seg) {
			return ""
		}
	}
	return segments[0]
}

func isPackagePath(path string) bool {
	return strings.HasPrefix(path, runtime.ChainDomain()+"/p/")
}

func validSegment(s string) bool {
	if s == "" || s[0] < 'a' || s[0] > 'z' {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

func validSlug(s string) bool {
	if len(s) < minSlug || len(s) > maxSlug {
		return false
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
			return false
		}
	}
	return true
}

// validText refuses what makes one name look like another (T6): padding or
// doubled spaces, spaces other than ' ', anything not graphic (controls,
// format characters such as bidi controls and zero-width characters,
// private-use and unassigned code points), blank-drawn fillers, mimics
// (isMimic), combining marks but those of a script that needs them, on a
// letter of that script (so no accent added to a Latin letter, no variation
// selector, no invisible or foreign mark), Latin letters outside the
// alphabets in common use (IPA, small capitals, Roman numerals, click
// letters), Latin letters mixed with letters of a script that is not known
// to mix with it ("Gnο" with a Greek omicron, "ꓖno" with a Lisu letter),
// and text with no letter or digit. titleKey folds the look-alikes that
// remain. gnoweb refuses a subset of this, so it shows whatever this
// accepts.
func validText(s string, maxRunes int) bool {
	if s == "" || strings.TrimSpace(s) != s || strings.Contains(s, "  ") ||
		!utf8.ValidString(s) || utf8.RuneCountInString(s) > maxRunes {
		return false
	}
	var latin, foreign, alnum bool
	prev := rune(0)
	for _, r := range s {
		switch {
		case !unicode.IsGraphic(r), unicode.IsSpace(r) && r != ' ', isBlankFiller(r), isMimic(r):
			return false
		case unicode.IsMark(r):
			if !sameMixingScript(prev, r) {
				return false
			}
		case unicode.Is(unicode.Latin, r):
			if !commonLatin(r) {
				return false
			}
			latin = true
		case unicode.IsLetter(r) && !unicode.In(r, mixesWithLatin...):
			foreign = true
		}
		alnum = alnum || unicode.IsLetter(r) || unicode.IsDigit(r)
		prev = r
	}
	return alnum && (!latin || !foreign)
}

// mixesWithLatin are the scripts a title may mix with Latin letters: none
// of their letters looks like a Latin one. Their own combining marks are
// accepted on their letters.
var mixesWithLatin = []*unicode.RangeTable{
	unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul,
	unicode.Arabic, unicode.Hebrew, unicode.Devanagari, unicode.Thai,
}

// sameMixingScript reports whether mark follows a character of its own
// script, one of mixesWithLatin.
func sameMixingScript(prev, mark rune) bool {
	for _, t := range mixesWithLatin {
		if unicode.Is(t, mark) {
			return unicode.Is(t, prev)
		}
	}
	return false
}

// commonLatin reports Latin letters of the alphabets in use: Basic Latin,
// Latin-1, Extended-A and -B (Romanian, Pinyin) but for the click letters,
// and the precomposed Vietnamese letters.
func commonLatin(r rune) bool {
	return r < 0x0250 && (r < 0x01C0 || r > 0x01C3) || r >= 0x1E00 && r <= 0x1EFF
}

// isMimic reports enclosing marks and the letters that mimic ASCII:
// fullwidth forms and mathematical alphanumerics.
func isMimic(r rune) bool {
	return unicode.Is(unicode.Me, r) ||
		r >= 0xFF01 && r <= 0xFF5E ||
		r >= 0x1D400 && r <= 0x1D7FF
}

// isBlankFiller reports letters and symbols that draw as blank space;
// gnoweb treats them as invisible too.
func isBlankFiller(r rune) bool {
	switch r {
	case 0x2800, 0x115F, 0x1160, 0x3164, 0xFFA0: // braille blank, Hangul fillers
		return true
	}
	return false
}

// validAsset accepts "" or ipfs://<cid>: content addressed, so it cannot
// change after review (T3), and small, so list responses stay small.
func validAsset(s string) bool {
	if s == "" {
		return true
	}
	cid, ok := strings.CutPrefix(s, "ipfs://")
	return ok && len(cid) >= minCID && len(cid) <= maxCID && isAlnum(cid)
}

func isAlnum(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}
