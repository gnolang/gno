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
	NewestFirst  bool // if true, newest posts are shown first
	NoBreadcrumb bool
}

func NewBlog(title, prefix string) *Blog {
	return &Blog{
		Title:  title,
		Prefix: prefix,
		Posts:  avl.NewTree(),
		PostId: 0,
	}
}

func (b *Blog) AddPostToBlog(post *Post) {
	post.ID = b.PostId
	b.PostId.Next()
	b.Posts.Set(b.PostId.String(), post)
}

func (b Blog) Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", b.RenderPosts)
	router.HandleFunc("id/{slug}", b.RenderPost)
	router.HandleFunc("tag/{slug}", b.RenderTags)
	router.HandleFunc("author/{slug}", b.RenderAuthors)

	return router.Render(path)
}

func (b Blog) RenderPosts(res *mux.ResponseWriter, req *mux.Request) {
	if b.Posts.Size() == 0 {
		res.Write("No posts.")
		return
	}

	pageSize := 9
	mode := parseQueryParams("mode", req.RawPath)
	if mode == "" {
		mode = Grid // default to grid mode
	}
	isGrid := mode == Grid
	if !isGrid {
		pageSize = 5
	}

	p := pager.NewPager(b.Posts, pageSize, b.NewestFirst)
	page := p.MustGetPageByPath(req.RawPath)

	res.Write(b.RenderListGrid(page, isGrid) + page.Picker(req.Path))
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

func (b Blog) RenderPost(res *mux.ResponseWriter, req *mux.Request) {
	id := req.GetVar("slug")
	post, found := b.Posts.Get(id)
	if !found {
		res.Write("Post not found.")
		return
	}
	out := post.(*Post).RenderPost(b.Prefix)
	res.Write(out)
}

func (b Blog) RenderTags(res *mux.ResponseWriter, req *mux.Request) {
	tag := req.GetVar("slug")
	res.Write(tag)
}

func (b Blog) RenderAuthors(res *mux.ResponseWriter, req *mux.Request) {
	author := req.GetVar("slug")
	res.Write(author)
}
