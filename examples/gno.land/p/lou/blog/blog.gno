package blog

import (
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/seqid"
	"gno.land/p/moul/md"
)

type Blog struct {
	Title        string
	Prefix       string
	Posts        *avl.Tree // id --> *Post
	PostId       seqid.ID
	PostsBySlug  *avl.Tree // slug --> *Post
	PostsByUpdatedAt *avl.Tree // updatedAt --> *Post
	DraftedPosts *avl.Tree // id --> *Post
}

func NewBlog(title, prefix string) (*Blog, error) {
	if prefix == "" {
		return nil, ErrEmptyPrefix
	}
	return &Blog{
		Title:  title,
		Prefix: prefix,
		Posts:  avl.NewTree(),
		PostsBySlug:  avl.NewTree(),
		PostsByUpdatedAt: avl.NewTree(),
		PostId: 0,
	}, nil
}

func (b *Blog) AddPost(post *Post) error { // drafted posts are also supported
	b.PostId.Next()
	post.ID = b.PostId
	var added bool
	if _, exists := b.PostsBySlug.Get(post.Slug); exists {
		return ErrPostAlreadyExists
	}
	added = b.Posts.Set(b.PostId.String(), post)
	added = b.PostsBySlug.Set(post.Slug, post)
	added = b.PostsByUpdatedAt.Set(post.UpdatedAt.String(), post)
	if !added {
		return ErrPostAddFailed
	}
	return nil
}

func (b *Blog) UpdatePostById(id seqid.ID, newPost *Post) error {
	existingPost, found := b.Posts.Get(id.String())
	if !found {
		return ErrPostNotFound
	}
	post := existingPost.(*Post)
	postBySlug, _ := b.PostsBySlug.Get(post.Slug)
	postByUpdated, _ := b.PostsByUpdatedAt.Get(post.UpdatedAt.String())
	UpdatePost(post, newPost)
	UpdatePost(postBySlug.(*Post), newPost)
	UpdatePost(postByUpdated.(*Post), newPost)
	return nil
}

func (b *Blog) UpdatePostBySlug(slug string, newPost *Post) error {
	existingPost, found := b.PostsBySlug.Get(slug)
	if !found {
		return ErrPostNotFound
	}
	post := existingPost.(*Post)
	postById, _ := b.Posts.Get(post.ID.String())
	postByUpdated, _ := b.PostsByUpdatedAt.Get(post.UpdatedAt.String())
	UpdatePost(post, newPost)
	UpdatePost(postById.(*Post), newPost)
	UpdatePost(postByUpdated.(*Post), newPost)
	return nil
}

func UpdatePost(existingPost, newPost *Post) {
	existingPost.Title = newPost.Title
	existingPost.Body = newPost.Body
	existingPost.Tags = newPost.Tags
	existingPost.Authors = newPost.Authors
	existingPost.UpdatedAt = time.Now()
}

func (b *Blog) DeletePostById(id string) error {
	post, found := b.Posts.Get(id)
	if !found {
		return ErrPostNotFound
	}
	return b.DeletePost(post.(*Post))
}

func (b *Blog) DeletePostBySlug(slug string) error {
	post, found := b.PostsBySlug.Get(slug)
	if !found {
		return ErrPostNotFound
	}
	return b.DeletePost(post.(*Post))
}

func (b *Blog) DeletePost(post *Post) error {
	var removed bool
	_, removed = b.Posts.Remove(post.ID.String())
	_, removed = b.PostsBySlug.Remove(post.Slug)
	_, removed = b.PostsByUpdatedAt.Remove(post.UpdatedAt.String())
	if !removed {
		return ErrDeleteFailed
	}
	return nil
}

func (b Blog) Render(path string) string {
	router := mux.NewRouter()
	router.HandleFunc("", b.RenderPosts)
	router.HandleFunc("post/{slug}", b.RenderPost)
	router.HandleFunc("tag/{slug}", b.RenderTags)
	router.HandleFunc("author/{slug}", b.RenderAuthors)
	router.HandleFunc("commenter/{slug}", b.RenderCommenters)

	return router.Render(path)
}

func RenderBlogHeader(prefix, rawPath string) string {
	out := md.Link("⊞ grid", updateQuery(rawPath, "mode", "grid")) + " | "
	out += md.Link("☰ list", updateQuery(rawPath, "mode", "list")) + " | "
	out += md.Link("▲ recent", updateQuery(rawPath, "sort", "recent")) + " | "
	out += md.Link("▼ oldest", updateQuery(rawPath, "sort", "oldest")) + " | "
	// last updated sort=updates order=ascending/descending (toggle between ▲ and ▼)
	// most popular sort=popular order=ascending/descending (toggle between ▲ and ▼)
	// alphabetical sort=alpha order=ascending/descending (toggle between ▲ and ▼)
	// time format
	out += md.Link("🗘 reset", prefix) + "\n"
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
	out := RenderBlogHeader(b.Prefix, req.RawPath)
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
	postSlug := req.GetVar("slug")
	if postSlug == "" {
		res.Write("Post slug is required.")
		return
	}

	foundKey, found := b.findPostByField("slug", postSlug)
	if !found {
		res.Write("Post not found.")
		return
	}
	post, found := b.Posts.Get(foundKey)
	out := post.(*Post).RenderPost(b.Prefix)
	res.Write(out)
}

func (b Blog) RenderTags(res *mux.ResponseWriter, req *mux.Request) {
	tag := req.GetVar("slug")
	out := b.RenderFilterHeader(req)

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
	out := b.RenderFilterHeader(req)

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

func (b Blog) RenderCommenters(res *mux.ResponseWriter, req *mux.Request) {
	commenter := req.GetVar("slug")
	out := b.RenderFilterHeader(req)
	// TO DO
	out += commenter
	res.Write(out)
}

func (b Blog) RenderFilterHeader(req *mux.Request) string {
	out := md.H2(md.Link(b.Prefix, b.Prefix)+"/"+req.Path) + "\n"
	out += md.Link("▲ recent", b.Prefix+":"+updateQuery(req.RawPath, "sort", "recent")) + " | "
	out += md.Link("▼ oldest", b.Prefix+":"+updateQuery(req.RawPath, "sort", "oldest")) + " | "
	out += md.Link("🗘 reset", b.Prefix) + "\n\n"
	return out
}

func (b Blog) Mention(role, recipient string) string {
	if role == "author" {
		return md.Bold(md.Link("@" + recipient, b.Prefix + "/author:" + recipient))
	}
	if role == "commenter" {
		return md.Bold(md.Link("@" + recipient, b.Prefix + "/commenter:" + recipient))
	}
	return ""
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

func (b Blog) findPostByField(field, value string) (string, bool) {
	var foundKey string
	var found bool
	b.Posts.Iterate("", "", func(k string, v interface{}) bool {
		post := v.(*Post)
		switch field {
		case "slug":
			if post.Slug == value {
				foundKey = k
				found = true
			}
		}
		return found
	})
	return foundKey, found
}
