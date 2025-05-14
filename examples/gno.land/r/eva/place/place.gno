package place

import (
	"encoding/base64"
	"errors"
	"net/url"
	"std"
	"strconv"
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/svg"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/p/moul/realmpath"
	"gno.land/p/moul/txlink"
)

type User struct {
	Address std.Address
	Rank    int
	Timer   time.Time
}

type Color struct {
	rgb     string
	initial string
}

const (
	memberRank int = 0
	adminRank  int = 1

	timeRefill int = 2

	rows uint64 = 45
	cols uint64 = 45

	white string = "rgb(250,250,250)"
)

var (
	Users *avl.Tree = avl.NewTree()

	colors = []Color{
		{"rgb(0,0,0)", "K"},     // black
		{"rgb(255,0,0)", "R"},   // red
		{"rgb(0,255,0)", "G"},   // green
		{"rgb(0,0,255)", "B"},   // blue
		{"rgb(255,255,0)", "Y"}, // yellow
		{"rgb(0,255,255)", "C"}, // cyan
		{"rgb(255,0,255)", "P"}, // pink
		{white, "W"},            // white
	}

	tableColor [][]Color // TODO: use a B tree for storage efficiency

	xStr string
	yStr string

	ErrInvalidCoordQuery = errors.New("Invalid coordinate value, expecting a uint64")
)

func init() {
	tableColor = make([][]Color, rows)
	for y := 0; y < int(rows); y++ {
		tableColor[y] = make([]Color, cols)
		for x := 0; x < int(cols); x++ {
			tableColor[y][x].rgb = white
			tableColor[y][x].initial = "⠀"
		}
	}
}

// parseXY takes a raw query like "?x=19&y=1" and returns the two ints.
func parseXY(rawQuery string) (x, y uint64, err error) {
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

func RegisterMember() {
	_, ok := retrieveMember(std.PreviousRealm().Address())
	if ok {
		panic("User already registered")
	}

	newMember := User{
		Address: std.PreviousRealm().Address(),
		Rank:    memberRank,
		Timer:   time.Now().Add(time.Duration(-timeRefill) * time.Minute),
	}

	Users.Set(std.PreviousRealm().Address().String(), newMember)
}

func retrieveMember(address std.Address) (User, bool) {
	raw, ok := Users.Get(address.String())
	if !ok {
		return User{}, false
	}

	return raw.(User), true
}

func isAdmin(address std.Address) bool {
	user, ok := retrieveMember(address)
	if !ok {
		panic("User not registered")
	}

	if user.Rank != adminRank {
		return false
	}

	return true
}

func SetNewAdmin(address std.Address) {
	if !isAdmin(std.PreviousRealm().Address()) {
		panic("Must be an admin")
	}

	user, ok := retrieveMember(address)
	if !ok || user.Rank == adminRank {
		panic("User does not meet the requirement")
	}

	user.Rank = adminRank
	Users.Set(address.String(), user)
}

func DeleteUser() {
	address := std.PreviousRealm().Address()

	_, ok := retrieveMember(address)
	if !ok {
		panic("User not registered")
	}

	Users.Remove(address.String())
}

func findColor(colorName string) (bool, Color) {
	for _, color := range colors {
		if color.initial == colorName {
			return true, color
		}
	}
	return false, Color{}
}

func Submit(x, y uint64, colorName string) {
	user, ok := retrieveMember(std.PreviousRealm().Address())
	if !ok {
		RegisterMember()
	}

	if time.Now().Before(user.Timer) {
		panic("Wait until " + user.Timer.String())
	}

	foundColor, color := findColor(colorName)

	if !foundColor {
		panic("Invalid color")
	}

	tableColor[y][x].rgb = color.rgb
	tableColor[y][x].initial = color.initial

	user.Timer = time.Now().Add(time.Duration(timeRefill) * time.Minute)
	Users.Set(std.PreviousRealm().Address().String(), user)
}

func GenerateSVG() string {
	out := ufmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, cols*40, rows*40)
	for y := 0; y < int(rows); y++ {
		for x := 0; x < int(cols); x++ {
			out += ufmt.Sprintf(`<rect x="%d" y="%d" width="40" height="40" fill="%s"/>`, x*40, y*40, tableColor[y][x].rgb)
		}
	}
	out += "</svg>"
	return out
}

func RenderAccountManagement(selectedColor string) string {
	var out string

	out += md.H2(md.Link("Submit", txlink.Call("Submit", "x", xStr, "y", yStr, "colorName", selectedColor)) + " | " +
		md.Link("Delete account", txlink.Call("DeleteUser")) + "\n\n")

	return out
}

func Render(path string) string {
	req := realmpath.Parse(path)
	query := req.Query

	out := md.H1("r/place") + "\n\n"
	if strings.Contains(path, "?") {
		if x, y, err := parseXY(path); err != nil {
			panic(err)
		} else {
			tableColor[y][x].initial = "X"
		}
	}

	selectedColor := query.Get("color")

	if selectedColor == "" {
		selectedColor = "W"
	}

	out += md.HorizontalRule() + "\n\n"
	for y, row := range tableColor {
		out += "|"
		for x, cell := range row {
			link := ufmt.Sprintf("/r/eva/place?color=%s&x=%d&y=%d", selectedColor, x, y)
			out += md.Link(cell.initial, link)
		}
		out += "|\n\n"
	}
	out += md.HorizontalRule() + "\n\n"

	out += RenderAccountManagement(selectedColor)

	ssvg := GenerateSVG()
	base64SVG := base64.StdEncoding.EncodeToString([]byte(ssvg))
	out += ufmt.Sprintf("![Pixel Canvas](data:image/svg+xml;base64,%s)", base64SVG)

	for _, color := range colors {
		query.Set("color", color.initial)

		out += ufmt.Sprintf("[!["+color.initial+`](data:image/svg+xml;base64,%s)](`+req.String()+")", generateSquareSvg(color.rgb)) + "\n\n" // I would have liked using md.Link here but doing so replaces the svg image with its link
	}

	return out
}

// Helper

func generateSquareSvg(color string) string {
	canvas := svg.Canvas{Width: 30, Height: 30}
	canvas.DrawRectangle(0, 0, 30, 30, color)
	return base64.StdEncoding.EncodeToString([]byte(canvas.String()))
}
