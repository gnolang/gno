// Ported from upstream Go's strings_test.go at go1.25.9:
// https://github.com/golang/go/blob/go1.25.9/src/strings/strings_test.go
// Skipped: cases needing unsafe, internal/testenv, very large allocations,
// or APIs not present in Gno (Clone, Runes, Lines, ContainsFunc, *Seq,
// testing.AllocsPerRun, rand).

package strings_test

import (
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

// --- TestCut ---

var cutTests = []struct {
	s, sep        string
	before, after string
	found         bool
}{
	{"abc", "b", "a", "c", true},
	{"abc", "a", "", "bc", true},
	{"abc", "c", "ab", "", true},
	{"abc", "abc", "", "", true},
	{"abc", "", "", "abc", true},
	{"abc", "d", "abc", "", false},
	{"", "d", "", "", false},
	{"", "", "", "", true},
}

func TestCut(t *testing.T) {
	for _, tt := range cutTests {
		if before, after, found := strings.Cut(tt.s, tt.sep); before != tt.before || after != tt.after || found != tt.found {
			t.Errorf("Cut(%q, %q) = %q, %q, %v, want %q, %q, %v", tt.s, tt.sep, before, after, found, tt.before, tt.after, tt.found)
		}
	}
}

// --- TestCutPrefix ---

var cutPrefixTests = []struct {
	s, sep string
	after  string
	found  bool
}{
	{"abc", "a", "bc", true},
	{"abc", "abc", "", true},
	{"abc", "", "abc", true},
	{"abc", "d", "abc", false},
	{"", "d", "", false},
	{"", "", "", true},
}

func TestCutPrefix(t *testing.T) {
	for _, tt := range cutPrefixTests {
		if after, found := strings.CutPrefix(tt.s, tt.sep); after != tt.after || found != tt.found {
			t.Errorf("CutPrefix(%q, %q) = %q, %v, want %q, %v", tt.s, tt.sep, after, found, tt.after, tt.found)
		}
	}
}

// --- TestCutSuffix ---

var cutSuffixTests = []struct {
	s, sep string
	before string
	found  bool
}{
	{"abc", "bc", "a", true},
	{"abc", "abc", "", true},
	{"abc", "", "abc", true},
	{"abc", "d", "abc", false},
	{"", "d", "", false},
	{"", "", "", true},
}

func TestCutSuffix(t *testing.T) {
	for _, tt := range cutSuffixTests {
		if before, found := strings.CutSuffix(tt.s, tt.sep); before != tt.before || found != tt.found {
			t.Errorf("CutSuffix(%q, %q) = %q, %v, want %q, %v", tt.s, tt.sep, before, found, tt.before, tt.found)
		}
	}
}

// --- TestEqualFold ---

var equalFoldTests = []struct {
	s, t string
	out  bool
}{
	{"abc", "abc", true},
	{"ABcd", "ABcd", true},
	{"123abc", "123ABC", true},
	{"αβδ", "ΑΒΔ", true}, // αβδ vs ΑΒΔ
	{"abc", "xyz", false},
	{"abc", "XYZ", false},
	{"abcdefghijk", "abcdefghijX", false},
	{"abcdefghijk", "abcdefghijK", true}, // Kelvin sign folds to k
	{"abcdefghijK", "abcdefghijK", true},
	{"abcdefghijkz", "abcdefghijKy", false},
	{"abcdefghijKz", "abcdefghijKy", false},
	{"1", "2", false},
	{"utf-8", "US-ASCII", false},
}

func TestEqualFold(t *testing.T) {
	for _, tt := range equalFoldTests {
		if out := strings.EqualFold(tt.s, tt.t); out != tt.out {
			t.Errorf("EqualFold(%#q, %#q) = %v, want %v", tt.s, tt.t, out, tt.out)
		}
		if out := strings.EqualFold(tt.t, tt.s); out != tt.out {
			t.Errorf("EqualFold(%#q, %#q) = %v, want %v", tt.t, tt.s, out, tt.out)
		}
	}
}

// --- TestRepeat ---

var repeatTests = []struct {
	in, out string
	count   int
}{
	{"", "", 0},
	{"", "", 1},
	{"", "", 2},
	{"-", "", 0},
	{"-", "-", 1},
	{"-", "----------", 10},
	{"abc ", "abc abc abc ", 3},
	{" ", " ", 1},
	{"--", "----", 2},
	{"===", "======", 2},
	{"000", "000000000", 3},
	{"\t\t\t\t", "\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t\t", 4},
}

func TestRepeat(t *testing.T) {
	for _, tt := range repeatTests {
		a := strings.Repeat(tt.in, tt.count)
		if a != tt.out {
			t.Errorf("Repeat(%q, %d) = %q; want %q", tt.in, tt.count, a, tt.out)
		}
	}
}

// --- TestRepeatCatchesOverflow ---
// Issue golang.org/issue/16237.

func repeatPanic(s string, count int) (msg string) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				msg = e.Error()
				return
			}
			if s, ok := r.(string); ok {
				msg = s
				return
			}
			msg = "panic"
		}
	}()
	strings.Repeat(s, count)
	return ""
}

func TestRepeatCatchesOverflow(t *testing.T) {
	type testCase struct {
		s      string
		count  int
		errStr string
	}
	const maxInt = int(^uint(0) >> 1)
	tests := []testCase{
		{"--", -2147483647, "negative"},
		{"", maxInt, ""},
		{"-", 10, ""},
		{"gopher", 0, ""},
		{"-", -1, "negative"},
		{"--", -102, "negative"},
		{string(make([]byte, 255)), int((^uint(0))/255 + 1), "overflow"},
	}
	for i, tt := range tests {
		msg := repeatPanic(tt.s, tt.count)
		if tt.errStr == "" {
			if msg != "" {
				t.Errorf("#%d panicked %q", i, msg)
			}
			continue
		}
		if msg == "" || !strings.Contains(msg, tt.errStr) {
			t.Errorf("#%d got %q want %q", i, msg, tt.errStr)
		}
	}

	// TODO: re-enable the upstream 64-bit case {"-", maxInt, "out of range"}
	// once https://github.com/gnolang/gno/pull/5723 merges. Until then the
	// allocator-overflow path triggers a host-side panic that crashes the test
	// binary (not just this test), so it cannot be exercised here.
}

// --- TestToValidUTF8 ---

var toValidUTF8Tests = []struct {
	in   string
	repl string
	out  string
}{
	{"", "�", ""},
	{"abc", "�", "abc"},
	{"﷝", "�", "﷝"},
	{"a\xffb", "�", "a�b"},
	{"a\xffb�", "X", "aXb�"},
	{"a☺\xffb☺\xC0\xAFc☺\xff", "", "a☺b☺c☺"},
	{"a☺\xffb☺\xC0\xAFc☺\xff", "日本語", "a☺日本語b☺日本語c☺日本語"},
	{"\xC0\xAF", "�", "�"},
	{"\xE0\x80\xAF", "�", "�"},
	{"\xed\xa0\x80", "abc", "abc"},
	{"\xed\xbf\xbf", "�", "�"},
	{"\xF0\x80\x80\xaf", "☺", "☺"},
	{"\xF8\x80\x80\x80\xAF", "�", "�"},
	{"\xFC\x80\x80\x80\x80\xAF", "�", "�"},
}

func TestToValidUTF8(t *testing.T) {
	for _, tc := range toValidUTF8Tests {
		got := strings.ToValidUTF8(tc.in, tc.repl)
		if got != tc.out {
			t.Errorf("ToValidUTF8(%q, %q) = %q; want %q", tc.in, tc.repl, got, tc.out)
		}
	}
}

// --- TestToUpper ---

var upperTests = []struct{ in, out string }{
	{"", ""},
	{"ONLYUPPER", "ONLYUPPER"},
	{"abc", "ABC"},
	{"AbC123", "ABC123"},
	{"azAZ09_", "AZAZ09_"},
	{"longStrinGwitHmixofsmaLLandcAps", "LONGSTRINGWITHMIXOFSMALLANDCAPS"},
	{"RENAN BASTOS 93 AOSDAJDJAIDJAIDAJIaidsjjaidijadsjiadjiOOKKO", "RENAN BASTOS 93 AOSDAJDJAIDJAIDAJIAIDSJJAIDIJADSJIADJIOOKKO"},
	{"longɐstringɐwithɐnonasciiⱯchars", "LONGⱯSTRINGⱯWITHⱯNONASCIIⱯCHARS"},
	{"ɐɐɐɐɐ", "ⱯⱯⱯⱯⱯ"},             // grows one byte per char
	{"a\U0010FFFF", "A\U0010FFFF"}, // utf8.RuneSelf and utf8.MaxRune
}

func TestToUpper(t *testing.T) {
	for _, tc := range upperTests {
		got := strings.ToUpper(tc.in)
		if got != tc.out {
			t.Errorf("ToUpper(%q) = %q; want %q", tc.in, got, tc.out)
		}
	}
}

// --- TestToLower ---

var lowerTests = []struct{ in, out string }{
	{"", ""},
	{"abc", "abc"},
	{"AbC123", "abc123"},
	{"azAZ09_", "azaz09_"},
	{"longStrinGwitHmixofsmaLLandcAps", "longstringwithmixofsmallandcaps"},
	{"renan bastos 93 AOSDAJDJAIDJAIDAJIaidsjjaidijadsjiadjiOOKKO", "renan bastos 93 aosdajdjaidjaidajiaidsjjaidijadsjiadjiookko"},
	{"LONGⱯSTRINGⱯWITHⱯNONASCIIⱯCHARS", "longɐstringɐwithɐnonasciiɐchars"},
	{"ⱭⱭⱭⱭⱭ", "ɑɑɑɑɑ"},             // shrinks one byte per char
	{"A\U0010FFFF", "a\U0010FFFF"}, // utf8.RuneSelf and utf8.MaxRune
}

func TestToLower(t *testing.T) {
	for _, tc := range lowerTests {
		got := strings.ToLower(tc.in)
		if got != tc.out {
			t.Errorf("ToLower(%q) = %q; want %q", tc.in, got, tc.out)
		}
	}
}

// -----------------------------------------------------------------------------
// Tests ported from upstream go1.25.9 strings_test.go (batch 2).
// Adaptations: no unsafe, no slices.Equal (handwritten), no iter (no *Seq),
// no testing.AllocsPerRun, no rand.
// -----------------------------------------------------------------------------

var (
	abcd   = "abcd"
	faces  = "☺☻☹"
	commas = "1,2,3,4"
	dots   = "1....2....3....4"
)

func eqStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// --- TestIndex / TestLastIndex / TestIndexAny / TestLastIndexAny / TestIndexByte / TestLastIndexByte ---

type indexCase struct {
	s, sep string
	out    int
}

var indexTests = []indexCase{
	{"", "", 0},
	{"", "a", -1},
	{"", "foo", -1},
	{"fo", "foo", -1},
	{"foo", "foo", 0},
	{"oofofoofooo", "f", 2},
	{"oofofoofooo", "foo", 4},
	{"barfoobarfoo", "foo", 3},
	{"foo", "", 0},
	{"foo", "o", 1},
	{"abcABCabc", "A", 3},
	{"x", "a", -1},
	{"x", "x", 0},
	{"abc", "a", 0},
	{"abc", "b", 1},
	{"abc", "c", 2},
	{"abc", "x", -1},
	{"xab"[:2], "ab", -1},
	{"xabc"[:3], "abc", -1},
	{"oxoxoxoxoxoxoxoxoxoxoxoy", "oy", 22},
	{"oxoxoxoxoxoxoxoxoxoxoxox", "oy", -1},
	{"oxoxoxoxoxoxoxoxoxoxox☺", "☺", 22},
	// invalid UTF-8 byte sequence
	{"xx0123456789012345678901234567890123456789012345678901234567890120123456789012345678901234567890123456xxx\xed\x9f\xc0", "\xed\x9f\xc0", 105},
}

var lastIndexTests = []indexCase{
	{"", "", 0},
	{"", "a", -1},
	{"", "foo", -1},
	{"fo", "foo", -1},
	{"foo", "foo", 0},
	{"foo", "f", 0},
	{"oofofoofooo", "f", 7},
	{"oofofoofooo", "foo", 7},
	{"barfoobarfoo", "foo", 9},
	{"foo", "", 3},
	{"foo", "o", 2},
	{"abcABCabc", "A", 3},
	{"abcABCabc", "a", 6},
}

var indexAnyTests = []indexCase{
	{"", "", -1},
	{"", "a", -1},
	{"", "abc", -1},
	{"a", "", -1},
	{"a", "a", 0},
	{"\x80", "\xffb", 0},
	{"aaa", "a", 0},
	{"abc", "xyz", -1},
	{"abc", "xcz", 2},
	{"ab☺c", "x☺yz", 2},
	{"a☺b☻c☹d", "cx", len("a☺b☻")},
	{"a☺b☻c☹d", "uvw☻xyz", len("a☺b")},
	{"aRegExp*", ".(|)*+?^$[]", 7},
	{dots + dots + dots, " ", -1},
	{"012abcba210", "\xffb", 4},
	{"012\x80bcb\x80210", "\xffb", 3},
	{"0123456\xcf\x80abc", "\xcfb\x80", 10},
}

var lastIndexAnyTests = []indexCase{
	{"", "", -1},
	{"", "a", -1},
	{"", "abc", -1},
	{"a", "", -1},
	{"a", "a", 0},
	{"\x80", "\xffb", 0},
	{"aaa", "a", 2},
	{"abc", "xyz", -1},
	{"abc", "ab", 1},
	{"ab☺c", "x☺yz", 2},
	{"a☺b☻c☹d", "cx", len("a☺b☻")},
	{"a☺b☻c☹d", "uvw☻xyz", len("a☺b")},
	{"a.RegExp*", ".(|)*+?^$[]", 8},
	{dots + dots + dots, " ", -1},
	{"012abcba210", "\xffb", 6},
	{"012\x80bcb\x80210", "\xffb", 7},
	{"0123456\xcf\x80abc", "\xcfb\x80", 10},
}

func runIndexCases(t *testing.T, f func(s, sep string) int, name string, tests []indexCase) {
	for _, tt := range tests {
		got := f(tt.s, tt.sep)
		if got != tt.out {
			t.Errorf("%s(%q, %q) = %d; want %d", name, tt.s, tt.sep, got, tt.out)
		}
	}
}

func TestIndex(t *testing.T)     { runIndexCases(t, strings.Index, "Index", indexTests) }
func TestLastIndex(t *testing.T) { runIndexCases(t, strings.LastIndex, "LastIndex", lastIndexTests) }
func TestIndexAny(t *testing.T)  { runIndexCases(t, strings.IndexAny, "IndexAny", indexAnyTests) }
func TestLastIndexAny(t *testing.T) {
	runIndexCases(t, strings.LastIndexAny, "LastIndexAny", lastIndexAnyTests)
}

func TestIndexByte(t *testing.T) {
	for _, tt := range indexTests {
		if len(tt.sep) != 1 {
			continue
		}
		got := strings.IndexByte(tt.s, tt.sep[0])
		if got != tt.out {
			t.Errorf("IndexByte(%q, %q) = %d; want %d", tt.s, tt.sep[0], got, tt.out)
		}
	}
}

func TestLastIndexByte(t *testing.T) {
	cases := []indexCase{
		{"", "q", -1},
		{"abcdef", "q", -1},
		{"abcdefabcdef", "a", len("abcdef")},
		{"abcdefabcdef", "f", len("abcdefabcde")},
		{"zabcdefabcdef", "z", 0},
		{"a☺b☻c☹d", "b", len("a☺")},
	}
	for _, tt := range cases {
		got := strings.LastIndexByte(tt.s, tt.sep[0])
		if got != tt.out {
			t.Errorf("LastIndexByte(%q, %c) = %d; want %d", tt.s, tt.sep[0], got, tt.out)
		}
	}
}

// --- TestIndexRune ---

func TestIndexRune(t *testing.T) {
	tests := []struct {
		in   string
		rune rune
		want int
	}{
		{"", 'a', -1},
		{"", '☺', -1},
		{"foo", '☹', -1},
		{"foo", 'o', 1},
		{"foo☺bar", '☺', 3},
		{"foo☺☻☹bar", '☹', 9},
		{"a A x", 'A', 2},
		{"some_text=some_value", '=', 9},
		{"☺a", 'a', 3},
		{"a☻☺b", '☺', 4},

		// RuneError should match any invalid UTF-8 byte sequence.
		{"�", '�', 0},
		{"\xff", '�', 0},
		{"☻x�", '�', len("☻x")},
		{"☻x\xe2\x98", '�', len("☻x")},
		{"☻x\xe2\x98�", '�', len("☻x")},
		{"☻x\xe2\x98x", '�', len("☻x")},

		// Invalid rune values should never match.
		{"a☺b☻c☹d\xe2\x98�\xff�\xed\xa0\x80", -1, -1},
		{"a☺b☻c☹d\xe2\x98�\xff�\xed\xa0\x80", 0xD800, -1},
		{"a☺b☻c☹d\xe2\x98�\xff�\xed\xa0\x80", utf8.MaxRune + 1, -1},

		// 2-byte rune
		{"ӆ", 'ӆ', 0}, // ӆ
		{"a", 'ӆ', -1},
		{"  ӆ", 'ӆ', 2},

		// 3-byte rune
		{"Ꚃ", 'Ꚃ', 0}, // Ꚁ
		{"a", 'Ꚃ', -1},
		{"  Ꚃ", 'Ꚃ', 2},

		// 4-byte rune
		{"\U000212C0", '\U000212C0', 0}, // 𡋀
		{"a", '\U000212C0', -1},
		{"  \U000212C0", '\U000212C0', 2},
	}
	for _, tt := range tests {
		if got := strings.IndexRune(tt.in, tt.rune); got != tt.want {
			t.Errorf("IndexRune(%q, U+%04X) = %d; want %d", tt.in, tt.rune, got, tt.want)
		}
	}
}

// --- TestIndexFunc / LastIndexFunc ---

func TestIndexFunc(t *testing.T) {
	tests := []struct {
		in          string
		f           func(rune) bool
		name        string
		first, last int
	}{
		{"", unicode.IsLetter, "IsLetter", -1, -1},
		{"abc", unicode.IsDigit, "IsDigit", -1, -1},
		{"0123", unicode.IsDigit, "IsDigit", 0, 3},
		{"a1b", unicode.IsDigit, "IsDigit", 1, 1},
		{"๐๒12hello34๐๑", unicode.IsDigit, "IsDigit", 0, 18},
		{"ⱯⱯⱯⱯABCDhelloEFⱯⱯGHⱯⱯ", unicode.IsUpper, "IsUpper", 0, 34},

		// invalid UTF-8
		{"\x801", unicode.IsDigit, "IsDigit", 1, 1},
		{"\x80abc", unicode.IsDigit, "IsDigit", -1, -1},
	}
	for _, tc := range tests {
		first := strings.IndexFunc(tc.in, tc.f)
		if first != tc.first {
			t.Errorf("IndexFunc(%q, %s) = %d; want %d", tc.in, tc.name, first, tc.first)
		}
		last := strings.LastIndexFunc(tc.in, tc.f)
		if last != tc.last {
			t.Errorf("LastIndexFunc(%q, %s) = %d; want %d", tc.in, tc.name, last, tc.last)
		}
	}
}

// --- TestContains / TestContainsAny / TestContainsRune ---

func TestContains(t *testing.T) {
	tests := []struct {
		str, substr string
		expected    bool
	}{
		{"abc", "bc", true},
		{"abc", "bcd", false},
		{"abc", "", true},
		{"", "a", false},
		{"01xxxx", "01", true},
		{"xx01xx", "01", true},
		{"xxxx01", "01", true},
		{"01xxxxx"[1:], "01", false},
		{"xxxxx01"[:6], "01", false},
		{"0123456789ABCDEFG", "0123456789ABCDEFG", true},
		{"xx0123456789ABCDEFx", "0123456789ABCDEFG", false}, // issue 15679
	}
	for _, ct := range tests {
		if strings.Contains(ct.str, ct.substr) != ct.expected {
			t.Errorf("Contains(%q, %q) = %v, want %v", ct.str, ct.substr, !ct.expected, ct.expected)
		}
	}
}

func TestContainsAny(t *testing.T) {
	tests := []struct {
		str, substr string
		expected    bool
	}{
		{"", "", false},
		{"", "a", false},
		{"a", "", false},
		{"a", "a", true},
		{"aaa", "a", true},
		{"abc", "xyz", false},
		{"abc", "xcz", true},
		{"a☺b☻c☹d", "uvw☻xyz", true},
		{"aRegExp*", ".(|)*+?^$[]", true},
		{dots + dots + dots, " ", false},
	}
	for _, ct := range tests {
		if strings.ContainsAny(ct.str, ct.substr) != ct.expected {
			t.Errorf("ContainsAny(%q, %q) = %v, want %v", ct.str, ct.substr, !ct.expected, ct.expected)
		}
	}
}

func TestContainsRune(t *testing.T) {
	tests := []struct {
		str      string
		r        rune
		expected bool
	}{
		{"", 'a', false},
		{"a", 'a', true},
		{"aaa", 'a', true},
		{"abc", 'y', false},
		{"abc", 'c', true},
		{"a☺b☻c☹d", 'x', false},
		{"a☺b☻c☹d", '☻', true},
		{"aRegExp*", '*', true},
	}
	for _, ct := range tests {
		if strings.ContainsRune(ct.str, ct.r) != ct.expected {
			t.Errorf("ContainsRune(%q, %q) = %v, want %v", ct.str, ct.r, !ct.expected, ct.expected)
		}
	}
}

// --- TestCount ---

func TestCount(t *testing.T) {
	tests := []struct {
		s, sep string
		num    int
	}{
		{"", "", 1},
		{"", "notempty", 0},
		{"notempty", "", 9},
		{"smaller", "not smaller", 0},
		{"12345678987654321", "6", 2},
		{"611161116", "6", 3},
		{"notequal", "NotEqual", 0},
		{"equal", "equal", 1},
		{"abc1231231123q", "123", 3},
		{"11111", "11", 2},
	}
	for _, tt := range tests {
		if num := strings.Count(tt.s, tt.sep); num != tt.num {
			t.Errorf("Count(%q, %q) = %d, want %d", tt.s, tt.sep, num, tt.num)
		}
	}
}

// --- TestReplace ---

func TestReplace(t *testing.T) {
	tests := []struct {
		in       string
		old, new string
		n        int
		out      string
	}{
		{"hello", "l", "L", 0, "hello"},
		{"hello", "l", "L", -1, "heLLo"},
		{"hello", "x", "X", -1, "hello"},
		{"", "x", "X", -1, ""},
		{"radar", "r", "<r>", -1, "<r>ada<r>"},
		{"", "", "<>", -1, "<>"},
		{"banana", "a", "<>", -1, "b<>n<>n<>"},
		{"banana", "a", "<>", 1, "b<>nana"},
		{"banana", "a", "<>", 1000, "b<>n<>n<>"},
		{"banana", "an", "<>", -1, "b<><>a"},
		{"banana", "ana", "<>", -1, "b<>na"},
		{"banana", "", "<>", -1, "<>b<>a<>n<>a<>n<>a<>"},
		{"banana", "", "<>", 10, "<>b<>a<>n<>a<>n<>a<>"},
		{"banana", "", "<>", 6, "<>b<>a<>n<>a<>n<>a"},
		{"banana", "", "<>", 5, "<>b<>a<>n<>a<>na"},
		{"banana", "", "<>", 1, "<>banana"},
		{"banana", "a", "a", -1, "banana"},
		{"banana", "a", "a", 1, "banana"},
		{"☺☻☹", "", "<>", -1, "<>☺<>☻<>☹<>"},
	}
	for _, tt := range tests {
		if s := strings.Replace(tt.in, tt.old, tt.new, tt.n); s != tt.out {
			t.Errorf("Replace(%q, %q, %q, %d) = %q, want %q", tt.in, tt.old, tt.new, tt.n, s, tt.out)
		}
		if tt.n == -1 {
			s := strings.ReplaceAll(tt.in, tt.old, tt.new)
			if s != tt.out {
				t.Errorf("ReplaceAll(%q, %q, %q) = %q, want %q", tt.in, tt.old, tt.new, s, tt.out)
			}
		}
	}
}

// --- TestSplit / TestSplitAfter ---

type splitCase struct {
	s, sep string
	n      int
	a      []string
}

func TestSplit(t *testing.T) {
	tests := []splitCase{
		{"", "", -1, []string{}},
		{abcd, "", 2, []string{"a", "bcd"}},
		{abcd, "", 4, []string{"a", "b", "c", "d"}},
		{abcd, "", -1, []string{"a", "b", "c", "d"}},
		{faces, "", -1, []string{"☺", "☻", "☹"}},
		{faces, "", 3, []string{"☺", "☻", "☹"}},
		{faces, "", 17, []string{"☺", "☻", "☹"}},
		{"☺�☹", "", -1, []string{"☺", "�", "☹"}},
		{abcd, "a", 0, nil},
		{abcd, "a", -1, []string{"", "bcd"}},
		{abcd, "z", -1, []string{abcd}},
		{commas, ",", -1, []string{"1", "2", "3", "4"}},
		{dots, "...", -1, []string{"1", ".2", ".3", ".4"}},
		{faces, "☹", -1, []string{"☺☻", ""}},
		{faces, "~", -1, []string{faces}},
		{"1 2 3 4", " ", 3, []string{"1", "2", "3 4"}},
		{"1 2", " ", 3, []string{"1", "2"}},
		{"\xff-\xff", "", -1, []string{"\xff", "-", "\xff"}},
		{"\xff-\xff", "-", -1, []string{"\xff", "\xff"}},
	}
	for _, tt := range tests {
		a := strings.SplitN(tt.s, tt.sep, tt.n)
		if !eqStrings(a, tt.a) {
			t.Errorf("SplitN(%q, %q, %d) = %#v; want %#v", tt.s, tt.sep, tt.n, a, tt.a)
			continue
		}
		if tt.n == 0 {
			continue
		}
		s := strings.Join(a, tt.sep)
		if s != tt.s {
			t.Errorf("Join(SplitN(%q, %q, %d), %q) = %q", tt.s, tt.sep, tt.n, tt.sep, s)
		}
		if tt.n < 0 {
			b := strings.Split(tt.s, tt.sep)
			if !eqStrings(a, b) {
				t.Errorf("Split disagrees with SplitN(%q, %q, %d)", tt.s, tt.sep, tt.n)
			}
		}
	}
}

func TestSplitAfter(t *testing.T) {
	tests := []splitCase{
		{abcd, "a", -1, []string{"a", "bcd"}},
		{abcd, "z", -1, []string{abcd}},
		{abcd, "", -1, []string{"a", "b", "c", "d"}},
		{commas, ",", -1, []string{"1,", "2,", "3,", "4"}},
		{dots, "...", -1, []string{"1...", ".2...", ".3...", ".4"}},
		{faces, "☹", -1, []string{"☺☻☹", ""}},
		{faces, "~", -1, []string{faces}},
		{faces, "", -1, []string{"☺", "☻", "☹"}},
		{"1 2 3 4", " ", 3, []string{"1 ", "2 ", "3 4"}},
		{"1 2 3", " ", 3, []string{"1 ", "2 ", "3"}},
		{"1 2", " ", 3, []string{"1 ", "2"}},
		{"123", "", 2, []string{"1", "23"}},
		{"123", "", 17, []string{"1", "2", "3"}},
	}
	for _, tt := range tests {
		a := strings.SplitAfterN(tt.s, tt.sep, tt.n)
		if !eqStrings(a, tt.a) {
			t.Errorf("SplitAfterN(%q, %q, %d) = %v; want %v", tt.s, tt.sep, tt.n, a, tt.a)
			continue
		}
		s := strings.Join(a, "")
		if s != tt.s {
			t.Errorf("Join(SplitAfterN(%q, %q, %d), %q) = %q", tt.s, tt.sep, tt.n, tt.sep, s)
		}
		if tt.n < 0 {
			b := strings.SplitAfter(tt.s, tt.sep)
			if !eqStrings(a, b) {
				t.Errorf("SplitAfter disagrees with SplitAfterN(%q, %q, %d)", tt.s, tt.sep, tt.n)
			}
		}
	}
}

// --- TestFields / TestFieldsFunc ---

func TestFields(t *testing.T) {
	tests := []struct {
		s string
		a []string
	}{
		{"", []string{}},
		{" ", []string{}},
		{" \t ", []string{}},
		{" ", []string{}},
		{"  abc  ", []string{"abc"}},
		{"1 2 3 4", []string{"1", "2", "3", "4"}},
		{"1  2  3  4", []string{"1", "2", "3", "4"}},
		{"1\t\t2\t\t3\t4", []string{"1", "2", "3", "4"}},
		{"1 2 3 4", []string{"1", "2", "3", "4"}},
		{"   ", []string{}},
		{"\n™\t™\n", []string{"™", "™"}},
		{"\n 1™2    ™", []string{"1™2", "™"}},
		{"\n1� �2 3�4", []string{"1�", "�2", "3�4"}},
		{"1\xFF \xFF2\xFF \xFF", []string{"1\xFF", "\xFF2\xFF", "\xFF"}},
		{faces, []string{faces}},
	}
	for _, tt := range tests {
		a := strings.Fields(tt.s)
		if !eqStrings(a, tt.a) {
			t.Errorf("Fields(%q) = %v; want %v", tt.s, a, tt.a)
		}
	}
}

func TestFieldsFunc(t *testing.T) {
	// Same data as TestFields, predicate = unicode.IsSpace.
	tests := []struct {
		s string
		a []string
	}{
		{"", []string{}},
		{" ", []string{}},
		{"  abc  ", []string{"abc"}},
		{"1 2 3 4", []string{"1", "2", "3", "4"}},
	}
	for _, tt := range tests {
		a := strings.FieldsFunc(tt.s, unicode.IsSpace)
		if !eqStrings(a, tt.a) {
			t.Errorf("FieldsFunc(%q, IsSpace) = %v; want %v", tt.s, a, tt.a)
		}
	}
	xCases := []struct {
		s string
		a []string
	}{
		{"", []string{}},
		{"XX", []string{}},
		{"XXhiXXX", []string{"hi"}},
		{"aXXbXXXcX", []string{"a", "b", "c"}},
	}
	pred := func(c rune) bool { return c == 'X' }
	for _, tt := range xCases {
		a := strings.FieldsFunc(tt.s, pred)
		if !eqStrings(a, tt.a) {
			t.Errorf("FieldsFunc(%q, ==X) = %v; want %v", tt.s, a, tt.a)
		}
	}
}

// --- TestTrim (TrimLeft/Right/Prefix/Suffix folded in) ---

func TestTrim(t *testing.T) {
	tests := []struct {
		f            string
		in, arg, out string
	}{
		{"Trim", "abba", "a", "bb"},
		{"Trim", "abba", "ab", ""},
		{"TrimLeft", "abba", "ab", ""},
		{"TrimRight", "abba", "ab", ""},
		{"TrimLeft", "abba", "a", "bba"},
		{"TrimLeft", "abba", "b", "abba"},
		{"TrimRight", "abba", "a", "abb"},
		{"TrimRight", "abba", "b", "abba"},
		{"Trim", "<tag>", "<>", "tag"},
		{"Trim", "* listitem", " *", "listitem"},
		{"Trim", `"quote"`, `"`, "quote"},
		{"Trim", "ⱯⱯɐɐⱯⱯ", "Ɐ", "ɐɐ"},
		{"Trim", "\x80test\xff", "\xff", "test"},
		{"Trim", " Ġ ", " ", "Ġ"},
		{"Trim", " Ġİ0", "0 ", "Ġİ"},
		{"Trim", "abba", "", "abba"},
		{"Trim", "", "123", ""},
		{"Trim", "", "", ""},
		{"TrimLeft", "abba", "", "abba"},
		{"TrimLeft", "", "123", ""},
		{"TrimLeft", "", "", ""},
		{"TrimRight", "abba", "", "abba"},
		{"TrimRight", "", "123", ""},
		{"TrimRight", "", "", ""},
		{"TrimRight", "☺\xc0", "☺", "☺\xc0"},
		{"TrimPrefix", "aabb", "a", "abb"},
		{"TrimPrefix", "aabb", "b", "aabb"},
		{"TrimSuffix", "aabb", "a", "aabb"},
		{"TrimSuffix", "aabb", "b", "aab"},
	}
	for _, tc := range tests {
		var f func(string, string) string
		switch tc.f {
		case "Trim":
			f = strings.Trim
		case "TrimLeft":
			f = strings.TrimLeft
		case "TrimRight":
			f = strings.TrimRight
		case "TrimPrefix":
			f = strings.TrimPrefix
		case "TrimSuffix":
			f = strings.TrimSuffix
		default:
			t.Errorf("Undefined trim function %s", tc.f)
			continue
		}
		actual := f(tc.in, tc.arg)
		if actual != tc.out {
			t.Errorf("%s(%q, %q) = %q; want %q", tc.f, tc.in, tc.arg, actual, tc.out)
		}
	}
}

// --- TestTrimFunc ---

func TestTrimFunc(t *testing.T) {
	const space = "\t\v\r\f\n  　"
	isSpace := unicode.IsSpace
	isDigit := unicode.IsDigit
	isUpper := unicode.IsUpper
	isValidRune := func(r rune) bool { return r != utf8.RuneError }
	notValidRune := func(r rune) bool { return !isValidRune(r) }
	notSpace := func(r rune) bool { return !isSpace(r) }
	notDigit := func(r rune) bool { return !isDigit(r) }

	cases := []struct {
		name     string
		f        func(rune) bool
		in       string
		trimOut  string
		leftOut  string
		rightOut string
	}{
		{"IsSpace", isSpace,
			space + " hello " + space,
			"hello",
			"hello " + space,
			space + " hello"},
		{"IsDigit", isDigit,
			"๐๒12hello34๐๑",
			"hello",
			"hello34๐๑",
			"๐๒12hello"},
		{"IsUpper", isUpper,
			"ⱯⱯⱯⱯABCDhelloEFⱯⱯGHⱯⱯ",
			"hello",
			"helloEFⱯⱯGHⱯⱯ",
			"ⱯⱯⱯⱯABCDhello"},
		{"not IsSpace", notSpace,
			"hello" + space + "hello",
			space,
			space + "hello",
			"hello" + space},
		{"not IsDigit", notDigit,
			"hello๐๒1234๐๑helo",
			"๐๒1234๐๑",
			"๐๒1234๐๑helo",
			"hello๐๒1234๐๑"},
		{"IsValidRune", isValidRune,
			"ab\xc0a\xc0cd",
			"\xc0a\xc0",
			"\xc0a\xc0cd",
			"ab\xc0a\xc0"},
		{"not IsValidRune", notValidRune,
			"\xc0a\xc0",
			"a",
			"a\xc0",
			"\xc0a"},
		{"IsSpace empty", isSpace, "", "", "", ""},
		{"IsSpace space", isSpace, " ", "", "", ""},
	}
	for _, tc := range cases {
		got := strings.TrimFunc(tc.in, tc.f)
		if got != tc.trimOut {
			t.Errorf("TrimFunc(%q, %s) = %q; want %q", tc.in, tc.name, got, tc.trimOut)
		}
		got = strings.TrimLeftFunc(tc.in, tc.f)
		if got != tc.leftOut {
			t.Errorf("TrimLeftFunc(%q, %s) = %q; want %q", tc.in, tc.name, got, tc.leftOut)
		}
		got = strings.TrimRightFunc(tc.in, tc.f)
		if got != tc.rightOut {
			t.Errorf("TrimRightFunc(%q, %s) = %q; want %q", tc.in, tc.name, got, tc.rightOut)
		}
	}
}

// --- TestTrimSpace ---

func TestTrimSpace(t *testing.T) {
	const space = "\t\v\r\f\n  　"
	trimSpaceTests := []struct{ in, out string }{
		{"", ""},
		{"abc", "abc"},
		{space + "abc" + space, "abc"},
		{" ", ""},
		{" \t\r\n \t\t\r\r\n\n ", ""},
		{" \t\r\n x\t\t\r\r\n\n ", "x"},
		{"  \t\r\n x\t\t\r\r\ny\n 　", "x\t\t\r\r\ny"},
		{"1 \t\r\n2", "1 \t\r\n2"},
		{" x\x80", "x\x80"},
		{" x\xc0", "x\xc0"},
		{"x \xc0\xc0 ", "x \xc0\xc0"},
		{"x \xc0", "x \xc0"},
		{"x \xc0 ", "x \xc0"},
		{"x \xc0\xc0 ", "x \xc0\xc0"},
		{"x ☺\xc0\xc0 ", "x ☺\xc0\xc0"},
		{"x ☺ ", "x ☺"},
	}
	for _, tc := range trimSpaceTests {
		got := strings.TrimSpace(tc.in)
		if got != tc.out {
			t.Errorf("TrimSpace(%q) = %q; want %q", tc.in, got, tc.out)
		}
	}
}

// --- TestMap ---

func tenRunes(ch rune) string {
	r := make([]rune, 10)
	for i := range r {
		r[i] = ch
	}
	return string(r)
}

func rot13(r rune) rune {
	step := rune(13)
	if r >= 'a' && r <= 'z' {
		return ((r-'a'+step)%26 + 'a')
	}
	if r >= 'A' && r <= 'Z' {
		return ((r-'A'+step)%26 + 'A')
	}
	return r
}

func TestMap(t *testing.T) {
	// 1. Grow.
	a := tenRunes('a')
	maxRuneFn := func(rune) rune { return unicode.MaxRune }
	m := strings.Map(maxRuneFn, a)
	expect := tenRunes(unicode.MaxRune)
	if m != expect {
		t.Errorf("growing: expected %q got %q", expect, m)
	}

	// 2. Shrink.
	minRuneFn := func(rune) rune { return 'a' }
	m = strings.Map(minRuneFn, tenRunes(unicode.MaxRune))
	expect = a
	if m != expect {
		t.Errorf("shrinking: expected %q got %q", expect, m)
	}

	// 3. Rot13.
	m = strings.Map(rot13, "a to zed")
	expect = "n gb mrq"
	if m != expect {
		t.Errorf("rot13: expected %q got %q", expect, m)
	}

	// 4. Rot13^2.
	m = strings.Map(rot13, strings.Map(rot13, "a to zed"))
	expect = "a to zed"
	if m != expect {
		t.Errorf("rot13^2: expected %q got %q", expect, m)
	}

	// 5. Drop.
	dropNotLatin := func(r rune) rune {
		if unicode.Is(unicode.Latin, r) {
			return r
		}
		return -1
	}
	m = strings.Map(dropNotLatin, "Hello, 세계")
	expect = "Hello"
	if m != expect {
		t.Errorf("drop: expected %q got %q", expect, m)
	}

	// 6. Handle invalid UTF-8 sequence.
	replaceNotLatin := func(r rune) rune {
		if unicode.Is(unicode.Latin, r) {
			return r
		}
		return utf8.RuneError
	}
	m = strings.Map(replaceNotLatin, "Hello\255World")
	expect = "Hello�World"
	if m != expect {
		t.Errorf("replace invalid sequence: expected %q got %q", expect, m)
	}

	// 7. utf8.RuneSelf / utf8.MaxRune round-trip.
	encode := func(r rune) rune {
		switch r {
		case utf8.RuneSelf:
			return unicode.MaxRune
		case unicode.MaxRune:
			return utf8.RuneSelf
		}
		return r
	}
	s := string(rune(utf8.RuneSelf)) + string(utf8.MaxRune)
	r := string(utf8.MaxRune) + string(rune(utf8.RuneSelf))
	m = strings.Map(encode, s)
	if m != r {
		t.Errorf("encoding not handled correctly: expected %q got %q", r, m)
	}
	m = strings.Map(encode, r)
	if m != s {
		t.Errorf("encoding not handled correctly: expected %q got %q", s, m)
	}

	// 8. trimSpaces.
	trimSpaces := func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}
	m = strings.Map(trimSpaces, "   abc    123   ")
	expect = "abc123"
	if m != expect {
		t.Errorf("trimSpaces: expected %q got %q", expect, m)
	}
}

// --- TestTitle (Title is deprecated upstream but still present in Gno) ---

func TestTitle(t *testing.T) {
	tests := []struct {
		in, out string
	}{
		{"", ""},
		{"a", "A"},
		{" aaa aaa aaa ", " Aaa Aaa Aaa "},
		{" Aaa Aaa Aaa ", " Aaa Aaa Aaa "},
		{"123a456", "123a456"},
		{"double-blind", "Double-Blind"},
		{"ÿøû", "Ÿøû"}, // only the first rune titled
		{"with_underscore", "With_underscore"},
		{"unicode \xe2\x80\xa8 line separator", "Unicode \xe2\x80\xa8 Line Separator"},
	}
	for _, tt := range tests {
		if s := strings.Title(tt.in); s != tt.out {
			t.Errorf("Title(%q) = %q, want %q", tt.in, s, tt.out)
		}
	}
}
