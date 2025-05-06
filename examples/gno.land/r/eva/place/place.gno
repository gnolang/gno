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

type Register struct {
	Users *avl.Tree
}

type Color struct {
	name string
	rgb  string
}

const (
	memberRank int = 0
	adminRank  int = 1

	timeRefill int = 2

	rows uint64 = 15
	cols uint64 = 80

	white string = "rgb(240,240,240)"
)

var (
	colors = []Color{
		{"black", "rgb(40,40,40)"},
		{"red", "rgb(255,0,0)"},
		{"green", "rgb(0,255,0)"},
		{"blue", "rgb(0,0,255)"},
		{"yellow", "rgb(255,255,0)"},
		{"cyan", "rgb(0,255,255)"},
		{"pink", "rgb(255,0,255)"},
		{"white", white},
	}

	table      [][]string // TODO: use a B tree for storage efficiency
	tableColor [][]string

	xStr string
	yStr string

	Registered      Register
	addressLoggedIn std.Address

	ErrInvalidCoordQuery = errors.New("Invalid coordinate value, expecting a uint64")
)

func generateSquareSvg(color string) string {
	canvas := svg.Canvas{Width: 30, Height: 30}
	canvas.DrawRectangle(0, 0, 30, 30, color)
	return base64.StdEncoding.EncodeToString([]byte(canvas.String()))
}

func init() {
	Registered.Users = avl.NewTree()

	table = make([][]string, rows)
	for y := 0; y < int(rows); y++ {
		table[y] = make([]string, cols)
		for x := 0; x < int(cols); x++ {
			link := ufmt.Sprintf("/r/eva/place?x=%d&y=%d", x, y)
			table[y][x] = md.Link("⠀", link)
		}
	}

	tableColor = make([][]string, rows)
	for y := 0; y < int(rows); y++ {
		tableColor[y] = make([]string, cols)
		for x := 0; x < int(cols); x++ {
			tableColor[y][x] = white
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
	_, ok := retrieveMember(std.OriginCaller())
	if ok {
		panic("User already registered")
	}

	newMember := User{
		Address: std.OriginCaller(),
		Rank:    memberRank,
		Timer:   time.Now().Add(time.Duration(-timeRefill) * time.Minute),
	}

	Registered.Users.Set(std.OriginCaller().String(), newMember)
}

func retrieveMember(address std.Address) (User, bool) {
	raw, ok := Registered.Users.Get(address.String())
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
	if !isAdmin(std.OriginCaller()) {
		panic("Must be an admin")
	}

	user, ok := retrieveMember(address)
	if !ok || user.Rank == adminRank {
		panic("User does not meet the requirement")
	}

	user.Rank = adminRank
	Registered.Users.Set(address.String(), user)
}

func DeleteUser() {
	address := std.OriginCaller()

	Registered.Users.Remove(address.String())
}

func Submit(x, y uint64, color string) {
	user, ok := retrieveMember(std.OriginCaller())
	if !ok {
		panic("User must be registered to submit a pixel")
	}

	if time.Now().Before(user.Timer) {
		panic("Wait until " + user.Timer.String())
	}

	table[y][x] = "O"
	tableColor[y][x] = color

	user.Timer = time.Now().Add(time.Duration(timeRefill) * time.Minute)
	Registered.Users.Set(std.OriginCaller().String(), user)
}

func GenerateSVG() string {
	out := ufmt.Sprintf(`<svg width="%d" height="%d" xmlns="http://www.w3.org/2000/svg">`, cols*40, rows*40)
	for y := 0; y < int(rows); y++ {
		for x := 0; x < int(cols); x++ {
			out += ufmt.Sprintf(`<rect x="%d" y="%d" width="40" height="40" fill="%s"/>`, x*40, y*40, tableColor[y][x])
		}
	}
	out += "</svg>"
	return out
}

func LogIn() {
	addressLoggedIn = std.OriginCaller()
}

func LogOut() {
	addressLoggedIn = ""
}

func Render(path string) string {
	req := realmpath.Parse(path)
	query := req.Query

	out := md.H1("r/place") + "\n\n"
	if strings.Contains(path, "?") {
		if x, y, err := parseXY(path); err != nil {
			panic(err)
		} else {
			table[y][x] = "X"
		}
	}

	out += md.HorizontalRule() + "\n\n"
	for _, row := range table {
		out += "|"
		for _, cell := range row {
			out += cell
		}
		out += "|\n\n"
	}
	out += md.HorizontalRule() + "\n\n"

	selectedColor := query.Get("color")

	if selectedColor == "" {
		selectedColor = white
	}

	if addressLoggedIn == "" {
		out += md.H2(md.Link("Log In", txlink.Call("LogIn")))
	} else {
		_, ok := retrieveMember(addressLoggedIn)
		if !ok {
			out += md.H2(md.Link("Register", txlink.Call("RegisterMember")) + "\n\n")
		} else {
			out += md.H2(md.Link("Submit", txlink.Call("Submit", "x", xStr, "y", yStr, "color", selectedColor)) + " | " +
				md.Link("Delete account", txlink.Call("DeleteUser")) + "\n\n")
		}
	}

	ssvg := GenerateSVG()
	base64SVG := base64.StdEncoding.EncodeToString([]byte(ssvg))
	out += ufmt.Sprintf("![Pixel Canvas](data:image/svg+xml;base64,%s)", base64SVG)

	for _, color := range colors {
		query.Set("color", color.name)

		out += ufmt.Sprintf("[!["+color.name+`](data:image/svg+xml;base64,%s)](`+req.String()+")", generateSquareSvg(color.rgb)) + "\n\n" // I would have liked using md.Link here but doing so replaces the svg image with its link
	}

	if addressLoggedIn != "" {
		out += md.H2(md.Link("Log out", txlink.Call("LogOut")))
	}

	return out
}
