package logos

import (
	"testing"

	require "github.com/jaekwon/testify/require"
)

func TestNewPage1(t *testing.T) {
	page := NewPage("this is a new string", 40, false, Style{})
	require.NotNil(t, page)
	size := page.Size
	require.Equal(t, size, Size{Width: 20, Height: 1})
}

func TestNewPage2(t *testing.T) {
	page := NewPage("this is a new string", 10, false, Style{})
	require.NotNil(t, page)
	size := page.Size
	/*
		0123456789
		this is a
		new string
	*/
	require.Equal(t, size, Size{Width: 10, Height: 2})
	require.Equal(t, page.Elems[0].GetCoord(), Coord{0, 0})
	require.Equal(t, page.Elems[1].GetCoord(), Coord{5, 0})
	require.Equal(t, page.Elems[2].GetCoord(), Coord{8, 0})
	require.Equal(t, page.Elems[3].GetCoord(), Coord{0, 1})
	require.Equal(t, page.Elems[4].GetCoord(), Coord{4, 1})
	require.Equal(t, len(page.Elems), 5)
}
