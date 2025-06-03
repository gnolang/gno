package blog

import (
	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/seqid"
	"gno.land/p/moul/md"
)

type Blog struct {
	Title        string
	Prefix       string
	Posts        *avl.Tree
	PostId       seqid.ID
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

func (b *Blog) AddPostToBlog(post *Post) { // add error
	post.ID = b.PostId
	b.PostId.Next()
	b.Posts.Set(b.PostId.String(), post)
}

func (b *Blog) UpdatePostById(id seqid.ID, post *Post) bool { // add error
	existingPost, found := b.Posts.Get(id.String())
	if !found {
		return false
	}
	existingPost.(*Post).Title = post.Title
	existingPost.(*Post).Body = post.Body
	existingPost.(*Post).Tags = post.Tags
	existingPost.(*Post).Authors = post.Authors
	existingPost.(*Post).UpdatedAt = post.UpdatedAt
	return true
}

func (b *Blog) DeletePostById(id seqid.ID) bool { // add error
	_, removed := b.Posts.Remove(id.String())
	if !removed {
		return false
	}
	return true
}

func (b Blog) Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", b.RenderPosts)
	router.HandleFunc("id/{slug}", b.RenderPost)
	router.HandleFunc("tag/{slug}", b.RenderTags)
	router.HandleFunc("author/{slug}", b.RenderAuthors)

	return router.Render(path)
}

func renderBlogHeader(prefix, rawPath string) string {
	out := md.Link("⊞ grid", updateQuery(rawPath, "mode", "grid")) + " | "
	out += md.Link("☰ list", updateQuery(rawPath, "mode", "list")) + " | "
	out += md.Link("▲ recent", updateQuery(rawPath, "sort", "recent")) + " | "
	out += md.Link("▼ oldest", updateQuery(rawPath, "sort", "oldest")) + " | "
	out += md.Link("🗘 reset", prefix) + "\n"
	// time ranges?
	return out
}

func (b Blog) RenderPosts(res *mux.ResponseWriter, req *mux.Request) {
	if b.Posts.Size() == 0 {
		res.Write("No posts found.")
		return
	}

	sort := parseQueryParams("sort", req.RawPath)
	if sort == "" {
		sort = "recent" // default to recent posts
	}
	newestFirst := sort == "recent"

	pageSize := 9
	mode := parseQueryParams("mode", req.RawPath)
	if mode == "" {
		mode = "grid" // default to grid mode
	}
	isGrid := mode == "grid"
	if !isGrid {
		pageSize = 5
	}

	p := pager.NewPager(b.Posts, pageSize, newestFirst)
	page := p.MustGetPageByPath(req.RawPath)
	out := renderBlogHeader(b.Prefix, req.RawPath)
	out += b.RenderListGrid(page, isGrid) + page.Picker(req.Path)
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
	out := md.H2(md.Link(b.Prefix, b.Prefix)+"/"+req.Path) + "\n"
	out += md.Link("▲ recent", b.Prefix+":"+updateQuery(req.RawPath, "sort", "recent")) + " | "
	out += md.Link("▼ oldest", b.Prefix+":"+updateQuery(req.RawPath, "sort", "oldest")) + " | "
	out += md.Link("🗘 reset", b.Prefix) + "\n\n"

	sort := parseQueryParams("sort", req.RawPath)
	if sort == "" {
		sort = "recent" // default to recent posts
	}
	newestFirst := sort == "recent"

	filteredPosts, exists := b.filterPostsByField("tag", tag)
	p := pager.NewPager(filteredPosts, 4, newestFirst)
	page := p.MustGetPageByPath(req.RawPath)
	for _, item := range page.Items {
		post := item.Value.(*Post)
		if hasField(post.Tags, tag) {
			out += post.RenderPreview(false, b.Prefix) + "\n"
			exists = true
		}
	}
	if !exists {
		out += "No posts found for this tag."
	}
	out += page.Picker(req.Path) + "\n\n"
	res.Write(out)
}

func (b Blog) RenderAuthors(res *mux.ResponseWriter, req *mux.Request) {
	author := req.GetVar("slug")
	out := md.H2(md.Link(b.Prefix, b.Prefix)+"/"+req.Path) + "\n"
	out += md.Link("▲ recent", b.Prefix+":"+updateQuery(req.RawPath, "sort", "recent")) + " | "
	out += md.Link("▼ oldest", b.Prefix+":"+updateQuery(req.RawPath, "sort", "oldest")) + " | "
	out += md.Link("🗘 reset", b.Prefix) + "\n\n"

	sort := parseQueryParams("sort", req.RawPath)
	if sort == "" {
		sort = "recent" // default to recent posts
	}
	newestFirst := sort == "recent"

	filteredPosts, exists := b.filterPostsByField("author", author)
	p := pager.NewPager(filteredPosts, 4, newestFirst)
	page := p.MustGetPageByPath(req.RawPath)
	for _, item := range page.Items {
		post := item.Value.(*Post)
		out += post.RenderPreview(false, b.Prefix) + "\n"
	}
	if !exists {
		out += "No posts found for this author."
	}
	out += page.Picker(req.Path)
	res.Write(out)
}

func (b Blog) filterPostsByField(field, value string) (*avl.Tree, bool) {
	filteredPosts := avl.NewTree()
	b.Posts.ReverseIterate("", "", func(k string, v interface{}) bool {
		post := v.(*Post)
		if field == "tag" && hasField(post.Tags, value) {
			filteredPosts.Set(k, post)
		} else if field == "author" && hasField(post.Authors, value) {
			filteredPosts.Set(k, post)
		}
		return false
	})
	return filteredPosts, filteredPosts.Size() > 0
}
