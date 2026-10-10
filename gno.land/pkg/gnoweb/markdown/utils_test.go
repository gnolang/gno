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
