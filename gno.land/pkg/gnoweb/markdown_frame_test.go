package gnoweb

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestFrameProductionTable renders through the production goldmark config,
// which enables GFM tables (the markdown package goldens do not): the close
// tag must end the table instead of becoming a row, and a frame tag inside
// a cell stays inline.
func TestFrameProductionTable(t *testing.T) {
	t.Parallel()
	gm := makeProductionRenderer()
	cases := map[string]string{
		"<gno-frame>\n| a | b |\n|---|---|\n| 1 | 2 |\n</gno-frame>\nafter\n": "<section class=\"gno-frame\">\n<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>1</td>\n<td>2</td>\n</tr>\n</tbody>\n</table>\n</section>\n<p>after</p>\n",
		"| <gno-frame> |\n|---|\n| x |\n":                                     "<table>\n<thead>\n<tr>\n<th><!-- raw HTML omitted --></th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>x</td>\n</tr>\n</tbody>\n</table>\n",
	}
	for in, want := range cases {
		var out bytes.Buffer
		require.NoError(t, gm.Convert([]byte(in), &out))
		require.Equal(t, want, out.String(), in)
	}
}
