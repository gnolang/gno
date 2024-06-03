package logos

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewPage1(t *testing.T) {
	page := NewPage("this is a new string", 40, false, nil)
	require.NotNil(t, page)
	size := page.Size
	require.Equal(t, Size{Width: 20, Height: 1}, size)
}

func TestNewPage2(t *testing.T) {
	page := NewPage("this is a new string", 10, false, nil)
	require.NotNil(t, page)
	size := page.Size
	/*
		0123456789
		this is a
		new string
	*/
	require.Equal(t, Size{Width: 10, Height: 2}, size)
	require.Equal(t, Coord{0, 0}, page.Elems[0].GetCoord())
	require.Equal(t, Coord{5, 0}, page.Elems[1].GetCoord())
	require.Equal(t, Coord{8, 0}, page.Elems[2].GetCoord())
	require.Equal(t, Coord{0, 1}, page.Elems[3].GetCoord())
	require.Equal(t, Coord{4, 1}, page.Elems[4].GetCoord())
	require.Equal(t, 5, len(page.Elems))
}

func TestNewPageSprint(t *testing.T) {
	page := NewPage("this is a new string", 10, false, nil)
	require.NotNil(t, page)
	/*
		0123456789
		this is a
		new string
	*/
	bpv := NewBufferedElemView(page, Size{})
	bpv.Render()
	out := bpv.Sprint()
	require.Equal(t, "this is a \nnew string", out)
}
