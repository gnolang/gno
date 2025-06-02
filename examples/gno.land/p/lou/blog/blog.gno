package blog

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/seqid"
)

const (
	Grid string = "grid"
	List string = "list"
)

type Blog struct {
	Title        string
	Prefix       string
	Posts        *avl.Tree
	PostId       seqid.ID
	DisplayMode  string // Grid or List
	NewestFirst  bool   // if true, newest posts are shown first
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
	post.ID = b.PostId
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
	router.HandleFunc("id/{slug}", b.RenderPost)
	return router.Render(path)
}

func (b Blog) RenderPosts(res *mux.ResponseWriter, req *mux.Request) {
	if b.Posts.Size() == 0 {
		res.Write("No posts.")
		return
	}
	pageSize := 9
	isGrid := b.DisplayMode == Grid
	if !isGrid {
		pageSize = 5
	}

	p := pager.NewPager(b.Posts, pageSize, b.NewestFirst)
	page := p.MustGetPageByPath(req.RawPath)

	out := "" // header
	out += b.RenderListGrid(page, isGrid)
	out += page.Picker(req.Path)
	res.Write(out)
}

func (b Blog) RenderPost(res *mux.ResponseWriter, req *mux.Request) {
	id := req.GetVar("slug")
	post, found := b.Posts.Get(id)
	if !found {
		res.Write("Post not found.")
		return
	}
	out := post.(*Post).RenderPost()
	res.Write(out)
}

func (b Blog) RenderListGrid(page *pager.Page, gridMode bool) string {
	colCount := 0
	out := "<gno-columns>\n"

	for _, item := range page.Items {
		if colCount%3 == 0 {
			out += "<gno-columns>\n"
		}

		post := item.Value.(*Post)
		out += post.RenderPreview(gridMode, b.Prefix)

		colCount++
		if colCount%3 == 0 {
			out += "</gno-columns>\n"
		} else if gridMode {
			out += "|||\n"
		}
	}

	if colCount%3 != 0 {
		out += "</gno-columns>\n"
	}

	return out
}
