// Package place is a package that reproduces the concept of Reddit's r/place.
// Users can interact with a grid composed of many pixels to color them
// When combined, those pixels can form images
package place

import (
	"errors"
	"net/url"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/r/leon/hor"
)

// This Color struct will store colors
// initial represents the initial of this color used in the interactive table
// rgb represents the rgb code of this color used in the display canva
type Color struct {
	rgb     string
	initial string
}

const (
	rows = 40
	cols = 40

	defaultColor = "rgb(250,250,250)"

	// default moderators wallets
	chaeWallet = "g1dzzx5qwcmawqnhhd76ehkteas5v2dg42hqnd7g"
	ghhvWallet = "g1uz7u2aeff4z8haftcsay2zp3rcczknkqu364mm"
	louWallet  = "g1pfyhn0d7g4tnp6wft9ge4cuu88ppr9u8mdggfs"
)

var (
	colors = []Color{
		{"rgb(0,0,0)", "K"},       // black
		{"rgb(255,0,0)", "R"},     // red
		{"rgb(0,255,0)", "G"},     // green
		{"rgb(0,0,255)", "B"},     // blue
		{"rgb(255,255,0)", "Y"},   // yellow
		{"rgb(0,255,255)", "C"},   // cyan
		{"rgb(255,0,255)", "P"},   // pink
		{"rgb(250,250,250)", "W"}, // white
	}

	tableColor [][]Color // TODO: use a B tree for storage efficiency

	xStr string
	yStr string

	ErrInvalidCoordQuery = errors.New("Invalid coordinate value, expecting a uint64")
)

func init() {
	hor.Register(cross, "R/place", "")

	setDefaultAdmin(chaeWallet)
	setDefaultAdmin(ghhvWallet)
	setDefaultAdmin(louWallet)

	tableColor = make([][]Color, rows)
	for y := 0; y < int(rows); y++ {
		tableColor[y] = make([]Color, cols)
		for x := 0; x < int(cols); x++ {
			tableColor[y][x].rgb = defaultColor
			tableColor[y][x].initial = "⠀"
		}
	}
}

// Submit places a tile on the tile of choice with the selected color
// Panics if the user's cooldown timer is not 0
func Submit(cur realm, x uint64, y uint64, colorName string) {
	user, ok := retrieveMember(std.PreviousRealm().Address())
	if !ok {
		Register(cross)
		user, ok = retrieveMember(std.PreviousRealm().Address())
		if !ok {
			panic("Can't register user")
		}
	}

	if time.Now().Before(user.Timer) {
		panic("wait until " + user.Timer.String())
	}

	foundColor, color := findColor(colorName)

	if !foundColor {
		panic("invalid color")
	}

	if x >= cols || y >= rows {
		panic("invalid coordinates")
	}

	tableColor[y][x].rgb = color.rgb
	tableColor[y][x].initial = color.initial

	user.Timer = time.Now().Add(time.Duration(timeRefill) * time.Minute)
}

// parseXY takes a raw query like "?x=19&y=1" and returns the two ints.
func parseXY(rawQuery string) (x uint64, y uint64, err error) {
	vals, err := url.ParseQuery(strings.TrimPrefix(rawQuery, "?"))
	if err != nil {
		return 0, 0, err
	}

	xStr = vals.Get("x")
	yStr = vals.Get("y")

	x, err = strconv.ParseUint(xStr, 10, 64)
	if err != nil {
		return 0, 0, ErrInvalidCoordQuery
	}
	y, err = strconv.ParseUint(yStr, 10, 64)
	if err != nil {
		return 0, 0, ErrInvalidCoordQuery
	}
	return x, y, nil
}
