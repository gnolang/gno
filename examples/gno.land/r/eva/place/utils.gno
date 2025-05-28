package place

import (
	"errors"
	"net/url"
	"std"
	"strconv"
	"strings"
)

var ErrInvalidCoordQuery = errors.New("Invalid coordinate value, expecting a uint64")

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

func findColor(colorName string) (bool, Color) {
	for _, color := range colors {
		if color.initial == colorName {
			return true, color
		}
	}
	return false, Color{}
}

func retrieveMember(address std.Address) (User, bool) {
	raw, ok := users.Get(address.String())
	if !ok {
		return User{}, false
	}

	return raw.(User), true
}

func countAdmins() int {
	var adminCounter int

	users.Iterate("", "", func(_ string, val any) bool {
		user := val.(User)

		if user.Rank == adminRank {
			adminCounter++
		}

		return false
	})

	return adminCounter
}
