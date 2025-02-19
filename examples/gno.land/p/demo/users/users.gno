package users

import (
	"std"
	"strconv"

	"gno.land/p/demo/avl"
)

type Profile struct {
	Header  string    // title
	Profile string    // short description
	Image   string    // base64 encoded image
	Blog    *avl.Tree // blog number -> URL
	Post    *avl.Tree // posts
}

type User struct {
	Name    string
	Address std.Address
	Invites int
	Inviter std.Address
	Profile Profile
	Number  int
}

func (u *User) Render() string {
	// Start with user header
	str := "## " + u.Profile.Header + "\n\n"

	// Basic user info
	str += "### " + u.Name + "\n" +
		" * address = " + string(u.Address) + "\n" +
		" * " + strconv.Itoa(u.Invites) + " invites\n"

	// Add inviter if exists
	if u.Inviter != "" {
		str += " * invited by " + string(u.Inviter) + "\n"
	}

	// Add Profile description
	str += "\n" + u.Profile.Profile + "\n"

	// Add image if exists
	if u.Profile.Image != "" {
		str += "\n![Profile Image][data:image/png;base64," + u.Profile.Image + "]\n"
	}

	// Add blogs if exist
	str += "\n### Blogs\n"
	u.Profile.Blog.Iterate("", "", func(key string, value interface{}) bool {
		str += " * Blog #" + key + ": " + value.(string) + "\n"
		return false
	})

	// Add posts if exist
	str += "\n### Posts\n"
	u.Profile.Post.Iterate("", "", func(key string, value interface{}) bool {
		str += " * Post #" + key + ": " + value.(string) + "\n"
		return false
	})

	return str
}
