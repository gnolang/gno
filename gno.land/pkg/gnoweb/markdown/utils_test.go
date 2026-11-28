package markdown

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScanGnoTag(t *testing.T) {
	prefix := []byte("<gno-x")
	type attr struct{ key, val string }
	cases := []struct {
		in          string
		size        int
		selfClosing bool
		attrs       []attr
	}{
		{`<gno-x a="1" b='2' c=3 d /> tail`, 27, true, []attr{{"a", "1"}, {"b", "2"}, {"c", "3"}, {"d", ""}}},
		{`<GNO-X A="&amp;"/>`, 18, true, []attr{{"A", "&amp;"}}}, // raw value, case kept
		{`<gno-x a="/>" />`, 16, true, []attr{{"a", "/>"}}},      // `/>` inside a value
		{`<gno-x a = "1">body`, 15, false, []attr{{"a", "1"}}},
		{`<gno-x>`, 7, false, nil},
		{`<gno-xy />`, 0, false, nil},                                    // another tag
		{`<gno-x a="1"`, 0, false, nil},                                  // unterminated
		{`<gno-x a="1" <gno-x b="2" />`, 0, false, nil},                  // stops at the next tag
		{`<gno-x a="<b>" />`, 17, true, []attr{{"a", "<b>"}}},            // '<' inside a value is fine
		{`<gno-x a="1` + "\n" + `" />`, 0, false, nil},                   // one line only
		{`<gno-x a="` + strings.Repeat("a", 64) + `" />`, 0, false, nil}, // over maxLen
		{`<gno`, 0, false, nil},
	}
	for _, c := range cases {
		var got []attr
		size, sc := scanGnoTag([]byte(c.in), prefix, 64, func(k, v []byte) {
			got = append(got, attr{string(k), string(v)})
		})
		require.Equal(t, c.size, size, c.in)
		if size == 0 {
			continue
		}
		require.Equal(t, c.selfClosing, sc, c.in)
		require.Equal(t, c.attrs, got, c.in)
	}
}

// FuzzScanGnoTag: no panic, a size within the input and the bound, a found
// tag that ends in '>', and attributes that are slices of the input.
func FuzzScanGnoTag(f *testing.F) {
	for _, seed := range []string{
		`<gno-x a="1" />`, `<gno-x a='/>' b=c d>`, `<gno-x a="`, `<GNO-X/>`,
		`<gno-x a="1" <gno-x />`, "<gno-x\ta = \"\n\" />", `<gno-x =/>`, `<gno-`,
	} {
		f.Add([]byte(seed))
	}
	prefix := []byte("<gno-x")
	const maxLen = 64
	f.Fuzz(func(t *testing.T, src []byte) {
		size, _ := scanGnoTag(src, prefix, maxLen, func(key, val []byte) {
			if len(key) > len(src) || len(val) > len(src) {
				t.Fatalf("attribute longer than the input")
			}
		})
		if size < 0 || size > min(len(src), maxLen) {
			t.Fatalf("size %d out of [0, %d]", size, min(len(src), maxLen))
		}
		if size > 0 && src[size-1] != '>' {
			t.Fatalf("tag %q does not end in '>'", src[:size])
		}
	})
}
