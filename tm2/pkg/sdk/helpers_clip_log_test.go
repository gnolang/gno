package sdk

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClipLog_FastPath_Short(t *testing.T) {
	in := "short message"
	assert.Equal(t, in, clipLog(in))
}

func TestClipLog_FastPath_ExactCap(t *testing.T) {
	// Length exactly maxLogLineBytes, no newlines → fast path.
	in := strings.Repeat("a", maxLogLineBytes)
	assert.Equal(t, in, clipLog(in))
}

func TestClipLog_LongLine_Truncated(t *testing.T) {
	in := strings.Repeat("a", maxLogLineBytes*2)
	out := clipLog(in)
	assert.True(t, strings.HasSuffix(out, "...<truncated>"))
	assert.Equal(t, maxLogLineBytes+len("...<truncated>"), len(out))
}

func TestClipLog_PerLineCap_MultiLine(t *testing.T) {
	long := strings.Repeat("a", maxLogLineBytes*2)
	in := long + "\nshort line"
	out := clipLog(in)
	parts := strings.Split(out, "\n")
	assert.Equal(t, 2, len(parts))
	assert.True(t, strings.HasSuffix(parts[0], "...<truncated>"))
	assert.Equal(t, "short line", parts[1])
}

func TestClipLog_LineCount_Cap(t *testing.T) {
	var b strings.Builder
	for i := 0; i < maxLogLines+5; i++ {
		b.WriteString("line\n")
	}
	in := b.String()
	out := clipLog(in)
	parts := strings.Split(out, "\n")
	// Expect maxLogLines + 1 (the elision marker)
	assert.Equal(t, maxLogLines+1, len(parts))
	assert.Contains(t, parts[maxLogLines], "more lines elided")
}

func TestClipLog_BothCaps(t *testing.T) {
	long := strings.Repeat("X", maxLogLineBytes*2)
	var b strings.Builder
	for i := 0; i < maxLogLines*2; i++ {
		b.WriteString(long)
		b.WriteString("\n")
	}
	out := clipLog(b.String())
	parts := strings.Split(out, "\n")
	// First maxLogLines lines (all truncated) + elision marker
	assert.Equal(t, maxLogLines+1, len(parts))
	assert.True(t, strings.HasSuffix(parts[0], "...<truncated>"))
}

func TestClipLog_EmbeddedNewline_PreservedAsBoundary(t *testing.T) {
	// A wrapped error with embedded \n in a single msgtrace's msg
	// should be split correctly so per-line cap applies on each
	// physical line.
	in := "line1\nline2\nline3"
	out := clipLog(in)
	assert.Equal(t, in, out) // no truncation needed
}

func TestClipLog_Empty(t *testing.T) {
	assert.Equal(t, "", clipLog(""))
}
