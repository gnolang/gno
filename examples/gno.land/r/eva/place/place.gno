package place

import (
	"encoding/base64"
	"errors"
	"net/url"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
	"gno.land/p/moul/txlink"
)

type User struct {
	Address std.Address
	Rank    int
	Timer   int
}

type Register struct {
	Users *avl.Tree
}

const (
	memberRank = 0
	adminRank  = 1
	timeRefill = 15
)

var (
	table [][]string
	rows  uint64 = 15 // pour linstant
	cols  uint64 = 91

	xStr string
	yStr string

	Registered Register

	ErrInvalidCoordQuery = errors.New("Invalid coordinate value, expecting a uint64")
)

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
}

func RegisterMember() {
	newMember := User{
		Address: std.OriginCaller(),
		Rank:    memberRank,
		Timer:   0,
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

func refillTime(address std.Address) {
	user, ok := retrieveMember(address)
	if !ok {
		panic("Could not find requested user")
	}

	user.Timer = timeRefill
	Registered.Users.Set(address.String(), user)
}

func Submit(x, y uint64) {
	table[y][x] = "**O**"
}

func GenerateSVG() string {
	out := `<svg width="3600" height="600" xmlns="http://www.w3.org/2000/svg">`
	out += `<rect x="0" y="0" width="3600" height="600" fill="rgb(20,20,20)"/>`
	for y := 0; y < int(rows); y++ {
		for x := 0; x < int(cols); x++ {
			if table[y][x] == "**O**" || table[y][x] == "**X**" {
				out += ufmt.Sprintf(`<rect x="%d" y="%d" width="40" height="40" fill="rgb(255, 255, 255)"/>`, x*40, y*40)
			}
		}
	}
	out += "</svg>"
	return out
}

func Render(path string) string {
	out := md.H1("r/place") + "\n\n"
	if strings.Contains(path, "?") {
		if x, y, err := parseXY(path); err != nil {
			panic(err)
		} else {
			table[y][x] = "**X**"
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

	out += "## " + md.Link("Submit", txlink.Call("Submit", "x", xStr, "y", yStr)) + "\n\n"

	svg := GenerateSVG()
	base64SVG := base64.StdEncoding.EncodeToString([]byte(svg))
	out += ufmt.Sprintf("![Pixel Canvas](data:image/svg+xml;base64,%s)", base64SVG)

	return out
}
