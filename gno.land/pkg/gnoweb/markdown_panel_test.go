package gnoweb

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestPanelProductionTable renders through the production goldmark config,
// which enables GFM tables (the markdown package goldens do not): the close
// tag must end the table instead of becoming a row, and a panel tag inside
// a cell stays inline.
func TestPanelProductionTable(t *testing.T) {
	t.Parallel()
	gm := makeProductionRenderer()
	cases := map[string]string{
		"<gno-panel>\n| a | b |\n|---|---|\n| 1 | 2 |\n</gno-panel>\nafter\n": "<section class=\"gno-panel\">\n<table>\n<thead>\n<tr>\n<th>a</th>\n<th>b</th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>1</td>\n<td>2</td>\n</tr>\n</tbody>\n</table>\n</section>\n<p>after</p>\n",
		"| <gno-panel> |\n|---|\n| x |\n":                                     "<table>\n<thead>\n<tr>\n<th><!-- raw HTML omitted --></th>\n</tr>\n</thead>\n<tbody>\n<tr>\n<td>x</td>\n</tr>\n</tbody>\n</table>\n",
	}
	for in, want := range cases {
		var out bytes.Buffer
		require.NoError(t, gm.Convert([]byte(in), &out))
		require.Equal(t, want, out.String(), in)
	}
}
