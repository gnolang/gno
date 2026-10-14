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
	case !validTitle(l.Title):
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

// validTitle is validText for a title, the field titleKey keeps unique, so
// held to what titleKey can fold: letters of the scripts below only (Lisu
// and Cherokee, for one, are drawn as Latin capitals and fold to nothing),
// and of the modern Greek and Cyrillic alphabets only (archaic letters such
// as DIGAMMA draw as Latin ones too) and never both in one title, no
// compatibility form (isCompatibilityForm), ASCII digits only (a Myanmar or
// Devanagari zero draws as an o), and no punctuation or symbol beyond ASCII,
// the separators titleKey drops and typographic quotes (U+2223 DIVIDES draws
// as an l).
func validTitle(s string) bool {
	if !validText(s, maxTitle) {
		return false
	}
	var greek, cyrillic bool
	for _, r := range s {
		greek = greek || unicode.Is(unicode.Greek, r)
		cyrillic = cyrillic || unicode.Is(unicode.Cyrillic, r)
		switch {
		case isCompatibilityForm(r):
			return false
		case unicode.IsLetter(r):
			if !unicode.Is(unicode.Latin, r) && !unicode.In(r, titleScripts...) || !modernGreekCyrillic(r) {
				return false
			}
		case unicode.IsDigit(r):
			if r > '9' {
				return false
			}
		case r > 0x7E && !unicode.IsMark(r) && foldRune(r) != 0 && (r < 0x2018 || r > 0x201D):
			return false
		}
	}
	return !greek || !cyrillic // their capitals draw alike: Π and П, Φ and Ф
}

// isCompatibilityForm reports the second encodings of letters a title may
// use: Hebrew and Arabic presentation forms, halfwidth Hangul and the
// conjoining jamo. Each draws as a base letter titleKey folds, but would key
// apart from it; the base letters and composed syllables stay allowed.
func isCompatibilityForm(r rune) bool {
	return r >= 0xFB1D && r <= 0xFDFF || r >= 0xFE70 && r <= 0xFEFF ||
		r >= 0xFFA0 && r <= 0xFFDC || r >= 0x1100 && r <= 0x11FF
}

// modernGreekCyrillic reports whether r, if Greek or Cyrillic, is a letter
// of the modern alphabets: Greek with its tonos letters, and the Cyrillic
// letters of Russian, Ukrainian, Belarusian, Serbian, Macedonian and
// Bulgarian.
func modernGreekCyrillic(r rune) bool {
	switch {
	case unicode.Is(unicode.Greek, r):
		return r >= 0x0386 && r <= 0x03CE
	case unicode.Is(unicode.Cyrillic, r):
		return r >= 0x0400 && r <= 0x045F || r == 'Ґ' || r == 'ґ'
	}
	return true
}

// titleScripts are the scripts a title may use besides Latin: those whose
// letters drawn like Latin ones titleKey folds, and those with none.
var titleScripts = []*unicode.RangeTable{
	unicode.Greek, unicode.Cyrillic,
	unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul,
	unicode.Arabic, unicode.Hebrew, unicode.Devanagari, unicode.Thai,
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
// and the letters of Latin-1, Extended-A and -B and Latin Extended
// Additional that latinBase folds (French, Romanian, Pinyin, Vietnamese...).
// The others, such as IPA, small capitals and LATIN CAPITAL LETTER IOTA,
// have no base letter to fold to, so each would key apart from the letter it
// draws like.
func commonLatin(r rune) bool {
	return r < 0x80 || latinBase(r) != 0
}

// latinBase returns the ASCII letter a Latin letter above U+00BF is drawn
// from, lower-case, or 0 if it has none. One byte per code point, '.' for
// none, from Unicode's canonical decompositions (the letter before the
// marks), plus the letters drawn as one without a decomposition (Æ, Ø, Ł,
// ß...).
func latinBase(r rune) rune {
	var c byte
	switch {
	case r >= 0x00C0 && r < 0x0250:
		c = latinBaseC0[r-0x00C0]
	case r >= 0x1E00 && r < 0x1F00:
		c = latinBase1E[r-0x1E00]
	}
	if c == 0 || c == '.' {
		return 0
	}
	return rune(c)
}

const (
	latinBaseC0 = "" +
		"aaaaaaaceeeeiiiidnooooo.ouuuuypbaaaaaaaceeeeiiiidnooooo.ouuuuypy" +
		"aaaaaaccccccccddddeeeeeeeeeegggggggghhhhiiiiiiiiillljjkkklllllll" +
		"lllnnnnnnnnnoooooooorrrrrrssssssssttttttuuuuuuuuuuuuwwyyyzzzzzzf" +
		"................................oo.............uu..............." +
		".............aaiioouuuuuuuuuu.aaaa....ggkkoooo..j...gg..nnaa...." +
		"aaaaeeeeiiiioooorrrruuuusstt..hh......aaeeooooooooyy............" +
		"................"
	latinBase1E = "" +
		"aabbbbbbccddddddddddeeeeeeeeeeffgghhhhhhhhhhiiiikkkkkkllllllllmm" +
		"mmmmnnnnnnnnoooooooopppprrrrrrrrssssssssssttttttttuuuuuuuuuuvvvv" +
		"wwwwwwwwwwxxxxyyzzzzzzhtwy....b.aaaaaaaaaaaaaaaaaaaaaaaaeeeeeeee" +
		"eeeeeeeeiiiioooooooooooooooooooooooouuuuuuuuuuuuuuyyyyyyyy......"
)

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
