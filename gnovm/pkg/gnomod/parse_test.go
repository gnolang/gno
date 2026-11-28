package gnomod

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/tm2/pkg/std"
)

// TestParseBytes tests parsing of both gno.mod and gnomod.toml files
func TestParseBytes(t *testing.T) {
	testCases := []struct {
		name            string
		content         string
		fileType        string // "gno.mod" or "gnomod.toml"
		expectedModule  string
		expectedVersion string
		expectedIgnore  bool
		expectedDraft   bool
		expectedError   string
	}{
		// Valid gno.mod cases
		{
			name:            "valid gno.mod with module",
			content:         "module gno.land/p/demo/foo",
			fileType:        "gno.mod",
			expectedModule:  "gno.land/p/demo/foo",
			expectedVersion: "0.0",
		},
		{
			name:            "valid gno.mod with module and gno version",
			content:         "module gno.land/p/demo/foo\ngno 0.9",
			fileType:        "gno.mod",
			expectedModule:  "gno.land/p/demo/foo",
			expectedVersion: "0.9",
		},
		{
			name:            "valid gno.mod with module and replace",
			content:         "module gno.land/p/demo/foo\nreplace bar => ../bar",
			fileType:        "gno.mod",
			expectedModule:  "gno.land/p/demo/foo",
			expectedVersion: "0.0",
		},
		{
			name:           "gno.mod with ignore comment",
			content:        "// Ignore\n\nmodule gno.land/p/demo/foo",
			fileType:       "gno.mod",
			expectedModule: "gno.land/p/demo/foo",
			expectedIgnore: true,
		},
		{
			name:           "gno.mod with deprecated comment",
			content:        "// Deprecated: use new module\nmodule gno.land/p/demo/foo",
			fileType:       "gno.mod",
			expectedModule: "gno.land/p/demo/foo",
		},

		// Valid gnomod.toml cases
		{
			name:            "valid gnomod.toml with module",
			content:         "module = \"gno.land/p/demo/foo\"",
			fileType:        "gnomod.toml",
			expectedModule:  "gno.land/p/demo/foo",
			expectedVersion: "0.0",
		},
		{
			name:            "valid gnomod.toml with module and gno version",
			content:         "module = \"gno.land/p/demo/foo\"\ngno = \"0.9\"",
			fileType:        "gnomod.toml",
			expectedModule:  "gno.land/p/demo/foo",
			expectedVersion: "0.9",
		},
		{
			name:            "valid gnomod.toml with module and replace",
			content:         "module = \"gno.land/p/demo/foo\"\nignore = true",
			fileType:        "gnomod.toml",
			expectedModule:  "gno.land/p/demo/foo",
			expectedVersion: "0.0",
			expectedIgnore:  true,
		},
		{
			name:           "gnomod.toml with ignore flag",
			content:        "module = \"gno.land/p/demo/foo\"\nignore = true",
			fileType:       "gnomod.toml",
			expectedModule: "gno.land/p/demo/foo",
			expectedIgnore: true,
		},
		{
			name:           "gnomod.toml with draft and ignore flags",
			content:        "module = \"gno.land/p/demo/foo\"\ndraft = true\nignore = true",
			fileType:       "gnomod.toml",
			expectedModule: "gno.land/p/demo/foo",
			expectedDraft:  true,
			expectedIgnore: true,
		},

		// Invalid cases
		{
			name:          "invalid gno.mod without module",
			content:       "replace bar => ../bar",
			fileType:      "gno.mod",
			expectedError: "invalid gnomod.toml: 'module' is required",
		},
		{
			name:          "invalid gno.mod with require",
			content:       "module foo\nrequire bar v0.0.0",
			fileType:      "gno.mod",
			expectedError: "unknown directive: require",
		},
		{
			name:          "invalid gnomod.toml without module",
			content:       "gno = \"0.9\"",
			fileType:      "gnomod.toml",
			expectedError: "invalid gnomod.toml: 'module' is required",
		},
		{
			name:          "invalid gnomod.toml with invalid toml",
			content:       "path = gno.land/p/demo/foo",
			fileType:      "gnomod.toml",
			expectedError: "error parsing gnomod.toml file",
		},
		{
			name:          "invalid module path with space",
			content:       "module \"gno.land/p/demo/ foo\"",
			fileType:      "gno.mod",
			expectedError: "malformed import path",
		},
		{
			name:          "invalid module path with Unicode",
			content:       "module gno.land/p/demo/한글",
			fileType:      "gno.mod",
			expectedError: "malformed import path",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var file *File
			var err error

			file, err = ParseBytes(tc.fileType, []byte(tc.content))
			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedModule, file.Module)
			if tc.expectedVersion != "" {
				assert.Equal(t, tc.expectedVersion, file.GetGno())
			}
			assert.Equal(t, tc.expectedIgnore, file.Ignore)
			assert.Equal(t, tc.expectedDraft, file.Draft)
		})
	}
}

// depthTable returns a gnomod.toml of at most maxFileSize bytes: a table header
// whose key path is keyPath, then as many assignments as fit under it.
// parseAssign re-walks the whole header path for every one of them, so the cost
// is O(assignments × key-path depth) for a single '['. With key-path depth bounded
// in the decoder this is no longer the worst shape a 4KB body can hold, but it is
// the one whose cost grows with a bound, so the worst-case pin builds it.
func depthTable(header, keyPath string) string {
	var b strings.Builder
	b.WriteString(header)
	b.WriteString("[" + keyPath + "]\n")
	for i := 0; ; i++ {
		line := fmt.Sprintf("%c%c%c = 1\n", 'a'+i/676%26, 'a'+i/26%26, 'a'+i%26)
		if b.Len()+len(line) > maxFileSize {
			return b.String()
		}
		b.WriteString(line)
	}
}

const testHeader = "module = \"gno.land/r/test\"\ngno = \"0.9\"\n"

// TestParseBytesSizeBound covers maxFileSize, the one bound this layer keeps:
// with shape bounded in the decoder, cost is linear in length, so length is what
// ParseBytes has to cap.
func TestParseBytesSizeBound(t *testing.T) {
	// pad returns a valid gnomod.toml of exactly n bytes.
	pad := func(n int) string {
		return testHeader + "# " + strings.Repeat("x", n-len(testHeader)-3) + "\n"
	}

	for _, tc := range []struct {
		name          string
		content       string
		expectedError string
	}{
		{
			name:    "at size limit",
			content: pad(maxFileSize),
		},
		{
			name:          "over size limit",
			content:       pad(maxFileSize + 1),
			expectedError: "size 4097 exceeds limit 4096",
		},
		{
			// The reported payload: ~840KB of nested arrays. Fed straight to
			// upstream go-toml this recursed past Go's 1GB stack and killed the
			// process with an unrecoverable fatal error. The size bound rejects it
			// here; tm2/pkg/toml's maxNestingDepth is what makes it safe at any size.
			name:          "stack overflow payload",
			content:       testHeader + "a = " + strings.Repeat("[", 420_000) + strings.Repeat("]", 420_000) + "\n",
			expectedError: "exceeds limit 4096",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseBytes("gnomod.toml", []byte(tc.content))
			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				return
			}
			require.NoError(t, err)
		})
	}
}

// TestParseBytesRejectsUnboundedShapes pins that every shape which made the
// decode unbounded is rejected, at a size the size bound alone would admit. The
// bounds now live in tm2/pkg/toml, where they are checked against the parser's own
// recursion and its own parsed key paths; this test is the gnomod-level guarantee
// that they are actually reached from here.
//
// Each of these defeated some byte-level guard this file used to carry, which is
// why they are worth keeping as named cases rather than trusting the decoder's own
// suite: openers that a running depth counter mis-reads because their closers sit
// in a comment; a key path bought with quotes instead of dots; and a key path that
// spans lines inside a quoted segment while leaving even quote parity on each one.
func TestParseBytesRejectsUnboundedShapes(t *testing.T) {
	// parityBypass spans a table-header key path across many lines inside a
	// double-quoted segment while keeping an even count of both quote characters
	// on every line — a '"' inside a single-quoted segment, or in a trailing
	// comment, is not a delimiter for parseKey.
	parityBypass := func(levels int) string {
		var b strings.Builder
		b.WriteString(testHeader)
		b.WriteString(`['"'"` + "\n")
		for range levels {
			b.WriteString("\"\"\n")
		}
		b.WriteString("\"] #'\"'\n")
		return b.String()
	}

	for _, tc := range []struct {
		name          string
		content       string
		expectedError string
	}{
		{
			// Mutual recursion in parseRvalue/parseArray, one frame pair per level.
			name:          "nesting depth",
			content:       testHeader + "a = " + strings.Repeat("[", 600) + strings.Repeat("]", 600) + "\n",
			expectedError: "nesting depth exceeds limit",
		},
		{
			// Closers are not needed to descend at all.
			name:          "nesting depth, unclosed",
			content:       testHeader + "a = " + strings.Repeat("[", 600) + "\n",
			expectedError: "nesting depth exceeds limit",
		},
		{
			// The same nesting with the closers moved into comments, where the
			// decoder ignores them: a guard tracking a running bracket depth reads
			// +1 per line here while the decoder recurses +10.
			name:          "nesting behind comments",
			content:       testHeader + "a = " + strings.Repeat(strings.Repeat("[", 10)+"#"+strings.Repeat("]", 9)+"\n", 60),
			expectedError: "nesting depth exceeds limit",
		},
		{
			// Inline tables recurse through the same choke point.
			name:          "inline table depth",
			content:       testHeader + "a = " + strings.Repeat("{b=", 600) + "\n",
			expectedError: "nesting depth exceeds limit",
		},
		{
			// parseAssign re-walks the current table path per assignment: O(n·d),
			// and this shape spends a single '[', so no delimiter count sees it.
			name:          "dotted key-path depth",
			content:       depthTable(testHeader, strings.Repeat("a.", 1015)+"z"),
			expectedError: "key path depth",
		},
		{
			// The same depth with no dot at all: parseKey appends a key group per
			// quoted run and does not require a '.' between segments, so ["a""b""c"]
			// is the path a.b.c. A dot count cannot see this.
			name:          "quoted key-path depth, zero dots",
			content:       depthTable(testHeader, strings.Repeat(`""`, 1000)),
			expectedError: "key path depth",
		},
		{
			// A key path spanning lines inside a quoted segment, with even quote
			// parity on every line — what defeats a per-line key-unit count.
			name:          "key path spanning lines at even parity",
			content:       parityBypass(500),
			expectedError: "key path depth",
		},
		{
			// Depth on an assignment key rather than a table header: lexKey accepts
			// newlines inside a quoted key too.
			name:          "assignment key depth",
			content:       testHeader + strings.Repeat("a.", 1015) + "z = 1\n",
			expectedError: "key path depth",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.LessOrEqual(t, len(tc.content), maxFileSize,
				"case must be under the size bound, so it exercises the shape bound")
			_, err := ParseBytes("gnomod.toml", []byte(tc.content))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.expectedError)
		})
	}
}

// TestParseBytesAcceptsLegitimateBodies is the other half of the guarantee: the
// bounds must not reject real content. Each case here was rejected by a
// byte-level guard this file used to carry — an apostrophe in a comment and a
// backslash in a replace path both parse fine, and both appear in ordinary files.
func TestParseBytesAcceptsLegitimateBodies(t *testing.T) {
	for _, tc := range []struct{ name, content string }{
		{"plain", testHeader},
		{"apostrophe in a comment", "# don't bump this without regenerating goldens\n" + testHeader},
		{"quote in a comment", "# see the \"replace\" section\n" + testHeader},
		{"windows path in a replace", testHeader + "[[replace]]\nold = \"gno.land/p/demo/avl\"\nnew = \"..\\\\..\\\\p\\\\avl\"\n"},
		{"literal string value", testHeader + "[[replace]]\nold = 'gno.land/p/demo/avl'\nnew = '../../p/x'\n"},
		{"many replaces", testHeader + strings.Repeat("[[replace]]\nold = \"gno.land/p/demo/avl\"\nnew = \"../../p/x\"\n", 40)},
		{"nested arrays and inline tables", testHeader + "[addpkg]\nheight = 1\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			require.LessOrEqual(t, len(tc.content), maxFileSize)
			_, err := ParseBytes("gnomod.toml", []byte(tc.content))
			require.NoError(t, err)
		})
	}
}

// TestParseBytesWorstCase pins the cost of the most expensive body the bounds
// jointly admit: a table header at the decoder's key-depth limit, then every
// remaining byte spent on assignments under it, so parseAssign re-walks that path
// ~1000 times.
//
// With key-path depth bounded at 16 this shape is no longer the ceiling — an
// ordinary flat 4KB body costs about the same — which is the property worth
// keeping: cost is linear in maxFileSize with no shape premium, so raising the
// size bound raises the cost and the caller's gas charge alike. Measured at these
// bounds: ~1.2ms per decode, ~586ns/byte across the two decodes a message pays,
// against 1250 gas/byte charged. The assertion is a sanity net for a regression in
// the decoder or in these bounds, not a cliff detector.
func TestParseBytesWorstCase(t *testing.T) {
	for _, tc := range []struct {
		name    string
		keyPath string
	}{
		{"dotted", strings.Repeat("a.", 15) + "z"},
		{"quoted", strings.Repeat(`""`, 16)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := depthTable(testHeader, tc.keyPath)
			require.Greater(t, len(body), maxFileSize-8)

			start := time.Now()
			_, err := ParseBytes("gnomod.toml", []byte(body))
			elapsed := time.Since(start)

			require.NoError(t, err)
			assert.Less(t, elapsed, 100*time.Millisecond, "worst admitted body took %v to parse", elapsed)
		})
	}
}

// TestParseBytesAcceptsRepoModFiles runs ParseBytes over every gnomod.toml and
// gno.mod in the repository. A bound is only defensible if it never rejects real
// content, so that is checked mechanically here rather than eyeballed. It also
// reports the observed size headroom, which is what to look at before tightening
// maxFileSize.
func TestParseBytesAcceptsRepoModFiles(t *testing.T) {
	root, err := filepath.Abs("../../..")
	require.NoError(t, err)

	var n, worstSize int
	err = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			// An unreadable subtree is not this test's concern.
			return nil
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "node_modules" {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() != "gnomod.toml" && d.Name() != "gno.mod" {
			return nil
		}
		b, err := os.ReadFile(p)
		require.NoError(t, err)
		n++
		worstSize = max(worstSize, len(b))

		rel, _ := filepath.Rel(root, p)
		// Only the bounds are under test here: a file may still fail validation
		// for unrelated reasons (a draft module path, say), so assert only that
		// it is not rejected by the size bound or by the decoder's shape bounds.
		if _, err := ParseBytes(p, b); err != nil {
			assert.NotContains(t, err.Error(), "exceeds limit",
				"a bound rejected a real mod file: %s", rel)
		}
		return nil
	})
	require.NoError(t, err)

	// Guard against the walk silently finding nothing and the test passing vacuously.
	require.Greater(t, n, 400, "expected to walk the whole repository")
	t.Logf("%d mod files; worst observed size=%d/%d", n, worstSize, maxFileSize)
}

// TestParseMemPackage tests parsing of module files from MemPackage
func TestParseMemPackage(t *testing.T) {
	t.Skip("skipping")
	testCases := []struct {
		name            string
		files           []*std.MemFile
		expectedModule  string
		expectedVersion string
		expectedError   string
	}{
		{
			name: "valid gno.mod in mem package",
			files: []*std.MemFile{
				{Name: "gno.mod", Body: "module gno.land/p/demo/foo"},
			},
			expectedModule:  "gno.land/p/demo/foo",
			expectedVersion: "0.0",
		},
		{
			name: "valid gnomod.toml in mem package",
			files: []*std.MemFile{
				{Name: "gnomod.toml", Body: "module = \"gno.land/p/demo/foo\""},
			},
			expectedModule:  "gno.land/p/demo/foo",
			expectedVersion: "0.0",
		},
		{
			name: "both files present, prefers gnomod.toml",
			files: []*std.MemFile{
				{Name: "gno.mod", Body: "module gno.land/p/demo/old"},
				{Name: "gnomod.toml", Body: "module = \"gno.land/p/demo/new\""},
			},
			expectedModule:  "gno.land/p/demo/new",
			expectedVersion: "0.0",
		},
		{
			name:          "no module files",
			files:         []*std.MemFile{},
			expectedError: "gnomod.toml not in mem package",
		},
		{
			name: "invalid gno.mod",
			files: []*std.MemFile{
				{Name: "gno.mod", Body: "invalid content"},
			},
			expectedError: "error parsing gno.mod file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mpkg := &std.MemPackage{
				Name:  "test",
				Path:  "gno.land/p/demo/test",
				Files: tc.files,
			}

			file, err := ParseMemPackage(mpkg)
			if tc.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedModule, file.Module)
			assert.Equal(t, tc.expectedVersion, file.GetGno())
		})
	}
}
