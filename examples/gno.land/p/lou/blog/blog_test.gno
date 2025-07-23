package blog

// import (
// 	"std"
// 	"testing"

// 	"gno.land/p/demo/uassert"
// 	"gno.land/p/demo/testutils"
// )

// Blog is created with a title and path.
// func TestCreateBlog(t *testing.T) {
// 	creator := testutils.TestAddress("creator")
// 	testing.SetRealm(std.NewUserRealm(creator))

// 	path := "gno.land/r/blog/test1"
// 	title := "My Test Blog"z
	
// 	blog, err := NewBlog(path, creator)
// 	println(creator)
// 	uassert.NoError(t, err)

// 	uassert.Equal(t, blog.title, title)
// 	uassert.Equal(t, blog.prefix, path)
// }

// Blog rejects invalid input (e.g. empty title/path).

// Blog renders with no posts (empty blog).

// Blog renders correctly with one post and no comments.
