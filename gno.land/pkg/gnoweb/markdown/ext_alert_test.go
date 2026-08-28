package markdown

import (
	"bytes"
	"fmt"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/text"
)

// markerCase is a source plus how far to advance the reader before handing it
// to the parser. `at` matters: the padding that pinned the reader was only set
// when the byte AFTER the `>` was a tab, so on `>>>>>\t` it fired at the
// innermost marker, not at the head of the line. A case that starts at offset
// 0 does not reach that branch and would pass with the bug reintroduced.
type markerCase struct {
	src string
	at  int
}

func (c markerCase) reader() text.Reader {
	r := text.NewReader([]byte(c.src))
	r.Advance(c.at)
	return r
}

// process runs inside Open, before the `[!KIND]` regex has decided whether
// the line is an alert at all. A block parser that goes on to decline must
// leave the reader byte-for-byte as it found it, or the parser that handles
// the line next inherits the change — a stray padding is what let
// `>>>>>` + TAB pin the reader and spin parser.openBlocks forever.
func TestAlertProcessDoesNotMutateReader(t *testing.T) {
	for _, tc := range []markerCase{
		{src: ">\tnot an alert\n"},
		{src: ">>>>>\t\n"},
		{src: ">>>>>\t\n", at: 4}, // innermost marker: `>` + TAB
		{src: "> \tnot an alert\n"},
		{src: "> [!NOTE] real alert\n"},
		{src: ">\t[!NOTE] real alert\n"},
		{src: ">\n"},
	} {
		t.Run(fmt.Sprintf("%q@%d", tc.src, tc.at), func(t *testing.T) {
			reader := tc.reader()
			lineBefore, segBefore := reader.Position()
			offBefore := reader.LineOffset()

			defaultAlertParser.process(reader)

			lineAfter, segAfter := reader.Position()
			require.Equal(t, lineBefore, lineAfter, "line number moved")
			require.Equal(t, segBefore, segAfter, "segment moved")
			require.Equal(t, offBefore, reader.LineOffset(), "line offset moved")
		})
	}
}

// process reports the offset FROM THE START OF THE LINE of the first byte
// after the marker. Counting from the marker instead left the indent inside
// the slice Open matches `[!KIND]` against, so an indented marker — legal in
// CommonMark up to three columns, and accepted by consumeMarker on every
// continuation line — silently degraded the whole alert to a blockquote.
func TestAlertProcessOffsetIncludesIndent(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want int
	}{
		{src: ">[!NOTE] t\n", want: 1},
		{src: "> [!NOTE] t\n", want: 2},
		{src: ">\t[!NOTE] t\n", want: 2},
		{src: " > [!NOTE] t\n", want: 3},
		{src: "  > [!NOTE] t\n", want: 4},
		{src: "   > [!NOTE] t\n", want: 5},
		{src: "   >\t[!NOTE] t\n", want: 5},
	} {
		t.Run(fmt.Sprintf("%q", tc.src), func(t *testing.T) {
			ok, got := defaultAlertParser.process(text.NewReader([]byte(tc.src)))
			require.True(t, ok, "marker not recognised")
			require.Equal(t, tc.want, got)
			// This is the offset Open slices at, so it must land on the `[`.
			require.Equal(t, byte('['), tc.src[got], "offset does not land on `[`")
		})
	}

	// Four columns of indent is an indented code block, not a block start.
	ok, _ := defaultAlertParser.process(text.NewReader([]byte("    > [!NOTE] t\n")))
	require.False(t, ok, "four-column indent must not be an alert marker")
}

// consumeMarker must make forward progress on a `>` marker. If it sets
// padding without consuming the tab, padding units silently absorb the
// blockquote parser's own Advance calls and the reader never leaves the byte.
func TestAlertConsumeMarkerAdvances(t *testing.T) {
	for _, tc := range []markerCase{
		{src: ">\tbody\n"},
		{src: "> body\n"},
		{src: ">body\n"},
		{src: ">>>>>\t\n"},
		{src: ">>>>>\t\n", at: 4}, // innermost marker: `>` + TAB
	} {
		t.Run(fmt.Sprintf("%q@%d", tc.src, tc.at), func(t *testing.T) {
			reader := tc.reader()
			_, before := reader.Position()

			require.True(t, consumeMarker(reader), "marker not recognised")

			_, after := reader.Position()
			require.Greater(t, after.Start, before.Start,
				"consumeMarker did not advance past the marker (padding=%d)", after.Padding)
		})
	}
}

// End-to-end guard for the pin: nested `>` markers followed by a TAB must
// render as plain nested blockquotes, exactly as upstream goldmark does
// without the alert extension loaded.
//
// A regression here loops while appending an ast.Blockquote per pass, so the
// assertion is fenced: a soft memory limit makes the runtime GC-thrash
// instead of growing without bound (the leaked nodes are live, so collection
// cannot reclaim them), and a deadline turns the hang into a failure. The
// precise cause is covered by the two unit tests above; this one proves the
// user-visible output.
func TestAlertNestedMarkersWithTabTerminate(t *testing.T) {
	defer debug.SetMemoryLimit(debug.SetMemoryLimit(256 << 20))

	for _, n := range []int{4, 5, 8, 13, 32} {
		src := strings.Repeat(">", n) + "\t"

		want, err := convertWithoutAlerts(src)
		require.NoError(t, err)

		type result struct {
			out string
			err error
		}
		done := make(chan result, 1)
		go func() {
			m := goldmark.New()
			ExtAlerts.Extend(m)
			var buf bytes.Buffer
			err := m.Convert([]byte(src), &buf)
			done <- result{buf.String(), err}
		}()

		select {
		case got := <-done:
			require.NoError(t, got.err)
			require.Equal(t, want, got.out, "n=%d diverged from plain goldmark", n)
		case <-time.After(10 * time.Second):
			t.Fatalf("n=%d: alert parser did not terminate — the `>`+TAB reader pin is back", n)
		}
	}
}

func convertWithoutAlerts(src string) (string, error) {
	var buf bytes.Buffer
	err := goldmark.New().Convert([]byte(src), &buf)
	return buf.String(), err
}

// A title is whatever follows `[!KIND]` on the line, trimmed; when that is
// empty the summary falls back to the kind name. The terminator has to be
// trimmed in full: nothing normalizes CRLF on the way in from a realm's
// Render, and trimming only `\n` leaves a lone `\r` that counts as a title and
// suppresses the fallback, so `> [!NOTE]\r\n` rendered an empty <summary>.
// Kept as a unit test rather than a golden because txtar round-tripping and
// git checkout normalization both eat lone `\r` bytes.
func TestAlertTitleTerminatorTrimmed(t *testing.T) {
	for _, tc := range []struct {
		src  string
		want string
	}{
		{src: "> [!NOTE]\n> body\n", want: "Note"},
		{src: "> [!NOTE]\r\n> body\r\n", want: "Note"},
		{src: "> [!NOTE]\t\r\n> body\r\n", want: "Note"},
		{src: "> [!NOTE] \r\n> body\r\n", want: "Note"},
		{src: "> [!NOTE] title\r\n> body\r\n", want: "title"},
		{src: "> [!NOTE]\ttitle\r\n> body\r\n", want: "title"},
	} {
		t.Run(fmt.Sprintf("%q", tc.src), func(t *testing.T) {
			m := goldmark.New()
			ExtAlerts.Extend(m)
			var buf bytes.Buffer
			require.NoError(t, m.Convert([]byte(tc.src), &buf))

			// The summary holds the kind icon, then the title, then the arrow.
			_, after, ok := strings.Cut(buf.String(), `</use></svg>`)
			require.True(t, ok, "no summary in %s", buf.String())
			title, _, ok := strings.Cut(after, `<svg>`)
			require.True(t, ok, "no arrow in %s", buf.String())
			require.Equal(t, tc.want, title)
		})
	}
}
