package time

import (
	"errors"
)

var UTC *Location = &utcLoc

var utcLoc = Location{name: "UTC"}

type Location struct {
	name string
}

func (l *Location) String() string {
	if l == nil {
		return "UTC"
	}
	return l.name
}

func LoadLocation(name string) (*Location, error) {
	if name == "" || name == "UTC" {
		return UTC, nil
	}
	return nil, errors.New("time: unsupported location; only UTC is available")
}
