package postit

import "std"

type User struct {
	Username  string
	Address   std.Address
	Name      string
	Bio       string
	Posts     []uint64
	Followers []string
	Following []string
	AvatarUrl string // use grc721 instead?
}
