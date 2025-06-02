package blog

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/mux"
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

func (b *Blog) AddPostToBlog(post *Post) {
	b.PostId.Next()
	b.Posts.Set(b.PostId.String(), post)
}

func (b *Blog) ToggleGridMode() {
	b.DisplayMode = Grid
}

func (b *Blog) ToggleListMode() {
	b.DisplayMode = List
}

func (b Blog) Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", b.RenderPosts)
	return router.Render(path)
}

func (b Blog) RenderPosts(res *mux.ResponseWriter, req *mux.Request) {
	if b.Posts.Size() == 0 {
		res.Write("No posts.")
		return
	}
	if b.DisplayMode == Grid {
		res.Write(b.RenderListGrid(true))
	}
	if b.DisplayMode == List {
		res.Write(b.RenderListGrid(false))
	}
}

func (b Blog) RenderListGrid(gridMode bool) string {
	colCount := 0
	out := ""
	out += "<gno-columns>\n"
	b.Posts.ReverseIterate("", "", func(key string, value any) bool {
		if colCount%3 == 0 {
			out += "<gno-columns>\n"
		}
		post := value.(*Post)
		out += post.RenderPreview()

		colCount++
		if colCount%3 == 0 {
			out += "</gno-columns>\n"
		} else if gridMode {
			out += "|||\n"
		}

		return false
	})
	if colCount%3 != 0 {
		out += "</gno-columns>\n"
	}

	return out
}
