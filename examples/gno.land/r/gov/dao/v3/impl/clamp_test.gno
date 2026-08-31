package impl

import (
	"strings"
	"testing"

	"gno.land/p/nt/uassert/v0"
)

// Sanitizing costs ~11,310 gas/byte, and rendering is reachable unauthenticated
// through vm/qrender under a 3,000,000,000 gas cap. ExecutorCreationRealm is
// dispatched through the public dao.Executor interface, so a hostile executor
// computes it per call and stores almost nothing: unclamped, a ~250KB value
// costs over 3G gas in one render and bricks the proposal page for everyone.
// Measured on the real render path: 2,839,117,770 gas unclamped versus
// 18,749,984 clamped. About 265KB of input crosses the query cap entirely.
func TestClampFieldBoundsSanitizerInput(t *testing.T) {
	uassert.Equal(t, "short", clampField("short", maxRenderedRealm),
		"a value within the bound must pass through untouched")

	huge := strings.Repeat("a", 250000)
	got := clampField(huge, maxRenderedRealm)
	uassert.True(t, len(got) < maxRenderedRealm+32,
		"the clamped value must be bounded by the limit, not by the input")
	uassert.True(t, strings.HasSuffix(got, "… truncated"),
		"a clamped value must say it was cut rather than look authored short")
}

// Cutting at a byte offset can land inside a multi-byte rune; the sanitizer
// tolerates invalid UTF-8, but handing it a split rune is sloppy and would
// render a replacement character mid-path.
func TestClampFieldCutsOnRuneBoundary(t *testing.T) {
	// 3-byte runes, so a 256-byte cut lands mid-rune (256 = 85*3 + 1).
	got := clampField(strings.Repeat("世", 200), maxRenderedRealm)
	body := strings.TrimSuffix(got, "… truncated")

	uassert.True(t, len(body)%3 == 0,
		"the cut must land on a rune boundary")
	uassert.Equal(t, strings.Repeat("世", len(body)/3), body,
		"every retained rune must be intact")
}

// Boundary and malformed input. clampField does byte arithmetic and backs off
// over UTF-8 continuation bytes, so the interesting cases are the ones where
// that loop has nowhere to back off to.
func TestClampFieldBoundaries(t *testing.T) {
	exact := strings.Repeat("a", maxRenderedRealm)
	uassert.Equal(t, exact, clampField(exact, maxRenderedRealm),
		"a value exactly at the bound must not be marked truncated")

	over := strings.Repeat("a", maxRenderedRealm+1)
	uassert.True(t, strings.HasSuffix(clampField(over, maxRenderedRealm), "… truncated"),
		"one byte over the bound must be cut")

	// Already-invalid UTF-8: every byte is a continuation, so the backoff
	// walks to zero. Must degrade to the marker rather than panic or loop.
	cont := strings.Repeat("\x80", maxRenderedRealm+10)
	uassert.Equal(t, "… truncated", clampField(cont, maxRenderedRealm),
		"an all-continuation string must degrade to just the marker")

	uassert.Equal(t, "… truncated", clampField("abc", 0),
		"a zero bound must still terminate")
}
