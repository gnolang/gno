package blog

import (
	"std"

	"gno.land/p/moul/md"
)

type UserResolver func(input string) (string, bool)

type User struct {
	Addr     std.Address
	Username string
}

func (u User) renderLink() string {
	return md.Link("@"+u.Username, "/r/gnoland/users/v1:"+u.Username)
}

func CheckUser(addr string, resolver UserResolver) (*User, error) {
	if std.Address(addr).IsValid() {
		return nil, ErrInvalidCaller
	}
	userData, _ := resolver(addr)
	return &User{
		Addr:     std.Address(addr),
		Username: userData,
	}, nil
}
