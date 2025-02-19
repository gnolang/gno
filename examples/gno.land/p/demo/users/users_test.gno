package users

import (
	"fmt"
	"std"
	"testing"

	"gno.land/p/demo/avl"
)

func TestRender(t *testing.T) {
	// Create a new user with profile
	user := &User{
		Name:    "alice",
		Address: std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0ZX"),
		Invites: 5,
		Inviter: std.Address("g1jg8mtutu9khhfwc4nxmuhcpftf0ZX"),
		Profile: Profile{
			Header:  "Blockchain Developer",
			Profile: "Passionate about Gno and blockchain technology",
			Image:   "base64EncodedImageString...",
			Blog:    avl.NewTree(), // 이미 포인터를 반환하므로 &를 붙일 필요 없음
			Post:    avl.NewTree(),
		},
	}

	// Add blogs
	user.Profile.Blog.Set("1", "https://myblog.com/post1")
	user.Profile.Blog.Set("2", "https://myblog.com/post2")

	// Add posts
	user.Profile.Post.Set("1", "Introduction to Gno")
	user.Profile.Post.Set("2", "Building with AVL Trees")

	// Get rendered output and print it
	rendered := user.Render()
	fmt.Println(rendered)
}
