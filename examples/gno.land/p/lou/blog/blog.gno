package blog

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/seqid"
)

const (
	Grid string = "grid"
	List string = "list"
)

type Blog struct {
	Title        string
	Prefix       string    // e.g. r/gnoland/blog:
	Posts        *avl.Tree // slug -> *Post
	PostId       seqid.ID
	DisplayMode  string // Grid or List
	NoBreadcrumb bool
}

func NewBlog(title, prefix string) *Blog {
	return &Blog{
		Title:       title,
		Prefix:      prefix,
		Posts:       avl.NewTree(),
		PostId:      0,
		DisplayMode: Grid, // default to grid mode
	}
}

func (b *Blog) AddPostToBlog(post *Post) *Blog {
	b.PostId.Next()
	b.Posts.Set(b.PostId.String(), post)
	return b
}

func (b *Blog) ToggleGridMode() *Blog {
	b.DisplayMode = Grid
	return b
}

func (b *Blog) ToggleListMode() *Blog {
	b.DisplayMode = List
	return b
}

func (b Blog) RenderPosts() string {
	out := ""

	if b.Posts.Size() == 0 {
		out += "No posts."
		return out
	}
	if b.DisplayMode == Grid {
		return b.RenderListGrid(true)
	}
	if b.DisplayMode == List {
		return b.RenderListGrid(false)
	}
	return out
}

func (b Blog) RenderListGrid(gridMode bool) string {
	colCount := 0
	out := ""
	out += "<gno-columns>\n"
	b.Posts.ReverseIterate("", "", func(key string, value any) bool {
		post := value.(*Post)
		out += post.RenderPost()
		colCount++
		if colCount % 4 == 0 {
			out += "</gno-columns>\n"
			if colCount != b.Posts.Size() {
				out += "<gno-columns>\n"
			}
		} else if gridMode {
			out += "|||\n"
		}
		return false
	})
	out += "</gno-columns>\n"
	return out
}
