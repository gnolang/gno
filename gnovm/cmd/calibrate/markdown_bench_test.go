package calibrate

// chain/markdown native calibration benchmarks.
//
// Each native is benched with WORST-CASE input shape (not random or
// zero-filled) — gas is an adversarial upper bound, so the calibrated
// slope must reflect what a malicious caller can force.
//
// Sizes sweep decade steps 1 .. 100,000 covering the realistic input
// range:
//   - small  (1, 10, 100):     username slugs, footnote refs, language tags
//   - medium (1000):            comment/profile-bio prose
//   - large  (10000, 100000):   post bodies, multi-paragraph proposal text
//
// Run:
//   cd gnovm/cmd/calibrate
//   go test -bench=BenchmarkNative_Markdown -benchtime=200ms -count=3 \
//       -timeout=15m . > markdown_bench_output.txt
//   python3 gen_native_table.py markdown_bench_output.txt

import (
	"strings"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// markdownSizes is the decade-step sweep used for every chain/markdown
// native. Decade steps give a clean log-linear regression across the
// realistic input range (small validators to multi-paragraph posts).
var markdownSizes = []int{1, 10, 100, 1000, 10000, 100000}

// fillWorstCase produces n bytes of input shaped to maximize a native's
// per-byte cost. The shape is chosen to defeat any fast-path the
// implementation might take.
func fillWorstCase(n int, shape string) string {
	if n <= 0 {
		return ""
	}
	out := make([]byte, 0, n)
	for len(out) < n {
		room := n - len(out)
		if room >= len(shape) {
			out = append(out, shape...)
		} else {
			out = append(out, shape[:room]...)
		}
	}
	return string(out)
}

// ----- StripBidiAndZeroWidth -----
// Worst case: 3-byte UTF-8 runes starting with 0xE2 that are NOT in the
// strip ranges (defeats the fast-path skip AND defeats the strip-and-skip
// branch). U+2010 HYPHEN encodes as 0xE2 0x80 0x90 — forces full decode
// + full copy on every rune.
func benchMarkdownStripBidi(b *testing.B, n int) {
	b.Helper()
	s := fillWorstCase(n, "\xe2\x80\x90") // U+2010 repeated
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, s)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/markdown", "StripBidiAndZeroWidth"), nReturns: 1}
	b.ResetTimer()
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Markdown_StripBidiAndZeroWidth_1(b *testing.B)   { benchMarkdownStripBidi(b, 1) }
func BenchmarkNative_Markdown_StripBidiAndZeroWidth_10(b *testing.B)  { benchMarkdownStripBidi(b, 10) }
func BenchmarkNative_Markdown_StripBidiAndZeroWidth_100(b *testing.B) { benchMarkdownStripBidi(b, 100) }
func BenchmarkNative_Markdown_StripBidiAndZeroWidth_1000(b *testing.B) {
	benchMarkdownStripBidi(b, 1000)
}
func BenchmarkNative_Markdown_StripBidiAndZeroWidth_10000(b *testing.B) {
	benchMarkdownStripBidi(b, 10000)
}
func BenchmarkNative_Markdown_StripBidiAndZeroWidth_100000(b *testing.B) {
	benchMarkdownStripBidi(b, 100000)
}

// ----- NormalizeBreaks -----
// Worst case: every byte triggers the slow-path branch (alternating
// \r\n pairs — each \r writes a \n and skips a byte).
func benchMarkdownNormalizeBreaks(b *testing.B, n int) {
	b.Helper()
	s := fillWorstCase(n, "\r\n")
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, s)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/markdown", "NormalizeBreaks"), nReturns: 1}
	b.ResetTimer()
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Markdown_NormalizeBreaks_1(b *testing.B)   { benchMarkdownNormalizeBreaks(b, 1) }
func BenchmarkNative_Markdown_NormalizeBreaks_10(b *testing.B)  { benchMarkdownNormalizeBreaks(b, 10) }
func BenchmarkNative_Markdown_NormalizeBreaks_100(b *testing.B) { benchMarkdownNormalizeBreaks(b, 100) }
func BenchmarkNative_Markdown_NormalizeBreaks_1000(b *testing.B) {
	benchMarkdownNormalizeBreaks(b, 1000)
}
func BenchmarkNative_Markdown_NormalizeBreaks_10000(b *testing.B) {
	benchMarkdownNormalizeBreaks(b, 10000)
}
func BenchmarkNative_Markdown_NormalizeBreaks_100000(b *testing.B) {
	benchMarkdownNormalizeBreaks(b, 100000)
}

// ----- EscapeInline -----
// Worst case: every byte is in the escape set (all `*` → every byte
// gets a `\` prefix, doubling output size).
func benchMarkdownEscapeInline(b *testing.B, n int) {
	b.Helper()
	s := strings.Repeat("*", n)
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, s)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/markdown", "EscapeInline"), nReturns: 1}
	b.ResetTimer()
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Markdown_EscapeInline_1(b *testing.B)      { benchMarkdownEscapeInline(b, 1) }
func BenchmarkNative_Markdown_EscapeInline_10(b *testing.B)     { benchMarkdownEscapeInline(b, 10) }
func BenchmarkNative_Markdown_EscapeInline_100(b *testing.B)    { benchMarkdownEscapeInline(b, 100) }
func BenchmarkNative_Markdown_EscapeInline_1000(b *testing.B)   { benchMarkdownEscapeInline(b, 1000) }
func BenchmarkNative_Markdown_EscapeInline_10000(b *testing.B)  { benchMarkdownEscapeInline(b, 10000) }
func BenchmarkNative_Markdown_EscapeInline_100000(b *testing.B) { benchMarkdownEscapeInline(b, 100000) }

// ----- EscapeTitle -----
// Worst case: every byte hits the wider escape set (all `"` — only in
// EscapeTitle's set, exercises the extra bytes vs EscapeInline).
func benchMarkdownEscapeTitle(b *testing.B, n int) {
	b.Helper()
	s := strings.Repeat(`"`, n)
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, s)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/markdown", "EscapeTitle"), nReturns: 1}
	b.ResetTimer()
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Markdown_EscapeTitle_1(b *testing.B)      { benchMarkdownEscapeTitle(b, 1) }
func BenchmarkNative_Markdown_EscapeTitle_10(b *testing.B)     { benchMarkdownEscapeTitle(b, 10) }
func BenchmarkNative_Markdown_EscapeTitle_100(b *testing.B)    { benchMarkdownEscapeTitle(b, 100) }
func BenchmarkNative_Markdown_EscapeTitle_1000(b *testing.B)   { benchMarkdownEscapeTitle(b, 1000) }
func BenchmarkNative_Markdown_EscapeTitle_10000(b *testing.B)  { benchMarkdownEscapeTitle(b, 10000) }
func BenchmarkNative_Markdown_EscapeTitle_100000(b *testing.B) { benchMarkdownEscapeTitle(b, 100000) }

// ----- PercentEncodeURL -----
// Worst case: mix of unsafe bytes AND bare `%` followed by non-hex,
// which exercises both the encoding branch and the lookahead branch.
func benchMarkdownPercentEncodeURL(b *testing.B, n int) {
	b.Helper()
	s := fillWorstCase(n, "%zz ")
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, s)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/markdown", "PercentEncodeURL"), nReturns: 1}
	b.ResetTimer()
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Markdown_PercentEncodeURL_1(b *testing.B)  { benchMarkdownPercentEncodeURL(b, 1) }
func BenchmarkNative_Markdown_PercentEncodeURL_10(b *testing.B) { benchMarkdownPercentEncodeURL(b, 10) }
func BenchmarkNative_Markdown_PercentEncodeURL_100(b *testing.B) {
	benchMarkdownPercentEncodeURL(b, 100)
}
func BenchmarkNative_Markdown_PercentEncodeURL_1000(b *testing.B) {
	benchMarkdownPercentEncodeURL(b, 1000)
}
func BenchmarkNative_Markdown_PercentEncodeURL_10000(b *testing.B) {
	benchMarkdownPercentEncodeURL(b, 10000)
}
func BenchmarkNative_Markdown_PercentEncodeURL_100000(b *testing.B) {
	benchMarkdownPercentEncodeURL(b, 100000)
}

// ----- MatchCharsetN -----
// Worst case: input matches the charset for its entire length — forces
// the loop to traverse every byte (mismatches short-circuit early).
// Charset: r/sys/users UserName shape (first [a-z], rest [a-z0-9_-]).
func userNameCharsets() (firstLo, firstHi, restLo, restHi uint64) {
	for c := byte('a'); c <= 'z'; c++ {
		if c < 64 {
			firstLo |= 1 << c
		} else {
			firstHi |= 1 << (c - 64)
		}
	}
	restLo, restHi = firstLo, firstHi
	for c := byte('0'); c <= '9'; c++ {
		restLo |= 1 << c
	}
	for _, c := range []byte{'_', '-'} {
		if c < 64 {
			restLo |= 1 << c
		} else {
			restHi |= 1 << (c - 64)
		}
	}
	return
}

func benchMarkdownMatchCharsetN(b *testing.B, n int) {
	b.Helper()
	s := strings.Repeat("a", n) // all-`a` fully matches the UserName charset
	firstLo, firstHi, restLo, restHi := userNameCharsets()
	m := newDispatchMachine(7)
	setBlockValueFromGo(m, 0, s)
	setBlockValueFromGo(m, 1, firstLo)
	setBlockValueFromGo(m, 2, firstHi)
	setBlockValueFromGo(m, 3, restLo)
	setBlockValueFromGo(m, 4, restHi)
	setBlockValueFromGo(m, 5, 1)
	setBlockValueFromGo(m, 6, 1<<30) // effectively no maxLen cap
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/markdown", "MatchCharsetN"), nReturns: 1}
	b.ResetTimer()
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Markdown_MatchCharsetN_1(b *testing.B)     { benchMarkdownMatchCharsetN(b, 1) }
func BenchmarkNative_Markdown_MatchCharsetN_10(b *testing.B)    { benchMarkdownMatchCharsetN(b, 10) }
func BenchmarkNative_Markdown_MatchCharsetN_100(b *testing.B)   { benchMarkdownMatchCharsetN(b, 100) }
func BenchmarkNative_Markdown_MatchCharsetN_1000(b *testing.B)  { benchMarkdownMatchCharsetN(b, 1000) }
func BenchmarkNative_Markdown_MatchCharsetN_10000(b *testing.B) { benchMarkdownMatchCharsetN(b, 10000) }
func BenchmarkNative_Markdown_MatchCharsetN_100000(b *testing.B) {
	benchMarkdownMatchCharsetN(b, 100000)
}

// ----- CodeFence -----
// CodeFence always scans the full content. All-backtick input exercises
// both the scan and the inner run-length-counter branch maximally.
func benchMarkdownCodeFence(b *testing.B, n int) {
	b.Helper()
	s := strings.Repeat("`", n)
	m := newDispatchMachine(2)
	setBlockValueFromGo(m, 0, s)
	setBlockValueFromGo(m, 1, 3)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/markdown", "CodeFence"), nReturns: 1}
	b.ResetTimer()
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Markdown_CodeFence_1(b *testing.B)      { benchMarkdownCodeFence(b, 1) }
func BenchmarkNative_Markdown_CodeFence_10(b *testing.B)     { benchMarkdownCodeFence(b, 10) }
func BenchmarkNative_Markdown_CodeFence_100(b *testing.B)    { benchMarkdownCodeFence(b, 100) }
func BenchmarkNative_Markdown_CodeFence_1000(b *testing.B)   { benchMarkdownCodeFence(b, 1000) }
func BenchmarkNative_Markdown_CodeFence_10000(b *testing.B)  { benchMarkdownCodeFence(b, 10000) }
func BenchmarkNative_Markdown_CodeFence_100000(b *testing.B) { benchMarkdownCodeFence(b, 100000) }

// ----- EscapeBlockHazards -----
// Worst case: `[a]: u\n(x\n` repeated (paren-title LRD with no `)`
// anywhere). After the bracket-walker + scanLinkText + scanLinkURL +
// scanLRDTail all share a single `scanBudget = 8 × len(s)`, the
// LRD-title scan branch is the most expensive per-byte path that
// still falls inside that cap: each `[a]:` triggers scanLRDTail, the
// title scan opens on `(`, then walks forward looking for `)` until
// the budget hits zero. After that, all subsequent LRD candidates
// short-circuit. The combination of pass-1 budget-exhausting forward
// walks, pass-2 escape of every `[`/`]` outside spans, and the
// per-line dispatch pass yields the adversarial upper bound. Picked
// for gas calibration so the slope reflects what a malicious caller
// can actually force.
func benchMarkdownEscapeBlockHazards(b *testing.B, n int) {
	b.Helper()
	s := fillWorstCase(n, "[a]: u\n(x\n")
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, s)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/markdown", "EscapeBlockHazards"), nReturns: 1}
	b.ResetTimer()
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Markdown_EscapeBlockHazards_1(b *testing.B) {
	benchMarkdownEscapeBlockHazards(b, 1)
}
func BenchmarkNative_Markdown_EscapeBlockHazards_10(b *testing.B) {
	benchMarkdownEscapeBlockHazards(b, 10)
}
func BenchmarkNative_Markdown_EscapeBlockHazards_100(b *testing.B) {
	benchMarkdownEscapeBlockHazards(b, 100)
}
func BenchmarkNative_Markdown_EscapeBlockHazards_1000(b *testing.B) {
	benchMarkdownEscapeBlockHazards(b, 1000)
}
func BenchmarkNative_Markdown_EscapeBlockHazards_10000(b *testing.B) {
	benchMarkdownEscapeBlockHazards(b, 10000)
}
func BenchmarkNative_Markdown_EscapeBlockHazards_100000(b *testing.B) {
	benchMarkdownEscapeBlockHazards(b, 100000)
}

// ----- EscapeBlockHazardsRich -----
// Same worst-case shape as EscapeBlockHazards (the bracket walker /
// LRD scan is the per-byte hotspot in both variants). Rich mode skips
// two per-line checks (line-leader, setext); slope should be slightly
// lower than the strict variant's, but the gap is dominated by the
// shared bracket-walker work.
func benchMarkdownEscapeBlockHazardsRich(b *testing.B, n int) {
	b.Helper()
	s := fillWorstCase(n, "[a]: u\n(x\n")
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, s)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "chain/markdown", "EscapeBlockHazardsRich"), nReturns: 1}
	b.ResetTimer()
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Markdown_EscapeBlockHazardsRich_1(b *testing.B) {
	benchMarkdownEscapeBlockHazardsRich(b, 1)
}
func BenchmarkNative_Markdown_EscapeBlockHazardsRich_10(b *testing.B) {
	benchMarkdownEscapeBlockHazardsRich(b, 10)
}
func BenchmarkNative_Markdown_EscapeBlockHazardsRich_100(b *testing.B) {
	benchMarkdownEscapeBlockHazardsRich(b, 100)
}
func BenchmarkNative_Markdown_EscapeBlockHazardsRich_1000(b *testing.B) {
	benchMarkdownEscapeBlockHazardsRich(b, 1000)
}
func BenchmarkNative_Markdown_EscapeBlockHazardsRich_10000(b *testing.B) {
	benchMarkdownEscapeBlockHazardsRich(b, 10000)
}
func BenchmarkNative_Markdown_EscapeBlockHazardsRich_100000(b *testing.B) {
	benchMarkdownEscapeBlockHazardsRich(b, 100000)
}

// Compile-time use of markdownSizes/gno to silence unused warnings if the
// table is later consumed by a programmatic harness (currently the explicit
// BenchmarkNative_Markdown_* shape matches the other benches in this dir).
var _ = markdownSizes
var _ gno.Name = "noop"
