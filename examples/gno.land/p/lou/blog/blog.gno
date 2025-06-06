package blog

import (
	"strings"
	"time"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/mux"
	"gno.land/p/demo/seqid"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/md"
)

type Blog struct {
	Title            string
	Prefix           string
	Posts            *avl.Tree // id --> *Post
	PostId           seqid.ID
	PostsBySlug      *avl.Tree // slug --> *Post
	PostsByUpdatedAt *avl.Tree // updatedAt --> *Post
	TagsIndex        *avl.Tree // tagName --> int
	TagsSorted       *avl.Tree // "<count>::tag" --> tag
	AuthorsIndex     *avl.Tree // authorAddress --> int
	AuthorsSorted    *avl.Tree // "<count>::author" --> author
}

func NewBlog(title, prefix string) (*Blog, error) {
	if prefix == "" {
		return nil, ErrEmptyPrefix
	}
	return &Blog{
		Title:            title,
		Prefix:           prefix,
		Posts:            avl.NewTree(),
		PostsBySlug:      avl.NewTree(),
		PostsByUpdatedAt: avl.NewTree(),
		PostId:           0,
		TagsIndex:        avl.NewTree(),
		TagsSorted:       avl.NewTree(),
		AuthorsIndex:     avl.NewTree(),
		AuthorsSorted:    avl.NewTree(),
	}, nil
}

func (b *Blog) AddPost(post *Post) error {
	b.PostId.Next()
	post.ID = b.PostId
	var added bool
	if _, exists := b.PostsBySlug.Get(post.Slug); exists {
		return ErrPostAlreadyExists
	}

	b.addToIndex(post)
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
	router.HandleFunc("posts", b.RenderPosts)
	router.HandleFunc("posts/{slug}", b.RenderPost)
	router.HandleFunc("tags", b.RenderTags)
	router.HandleFunc("tags/{slug}", b.RenderTags)
	router.HandleFunc("authors", b.RenderAuthors)
	router.HandleFunc("authors/{slug}", b.RenderAuthors)
	router.HandleFunc("commenter/{slug}", b.RenderCommenters)
	return router.Render(path)
}

func RenderBlogHeader(prefix, rawPath string) string {
	out := md.Link("⊞ grid", GetGridFmtPath(rawPath)) + " | "
	out += md.Link("☰ list", GetListFmtPath(rawPath)) + " | "
	out += md.Link("▲ recent", GetSortByPublishedPath(rawPath, "asc")) + " | "
	out += md.Link("▼ oldest", GetSortByPublishedPath(rawPath, "desc")) + " | "
	out += "⏱ " + md.Link("short", GetTimeFmtPath(rawPath, "short")) + ", "
	out += md.Link("full", GetTimeFmtPath(rawPath, "full")) + ", "
	out += md.Link("relative", GetTimeFmtPath(rawPath, "relative")) + " |"

	sort := parseQueryParams("sort", rawPath)
	order := parseQueryParams("order", rawPath)
	if order == "" {
		order = "asc"
	}
	out += md.Link(orderArrow("update", sort, order)+" last updated",
		GetSortByUpdatedPath(rawPath, order)) + " | "

	alphaOrder := "(A-Z)"
	if order == "desc" && sort == "alpha" {
		alphaOrder = "(Z-A)"
	}
	out += md.Link(orderArrow("alpha", sort, order)+" alphabetical "+alphaOrder,
		GetSortByAlphabeticalPath(rawPath, order)) + " | "

	out += md.Link("🗘 reset", prefix) + "\n"
	return out
}

func (b Blog) RenderPosts(res *mux.ResponseWriter, req *mux.Request) {
	out := md.H1(b.Title) + "\n"
	if b.Posts.Size() == 0 {
		res.Write("No posts found.")
		return
	}

	options := parseRenderOptions(req.RawPath)
	pageSize := 9
	if !options.IsGrid {
		pageSize = 5
	}

	var p *pager.Pager
	if options.IsLastUpdated {
		p = pager.NewPager(b.PostsByUpdatedAt, pageSize, options.Ascending)
	} else if options.IsAlphabetical {
		p = pager.NewPager(b.PostsBySlug, pageSize, options.Ascending)
	} else {
		p = pager.NewPager(b.Posts, pageSize, options.NewestFirst)
	}
	page := p.MustGetPageByPath(req.RawPath)
	out += RenderBlogHeader(b.Prefix, req.RawPath)
	out += b.RenderListGrid(page, options.IsGrid, options.TimeFormat) + page.Picker(req.Path)
	res.Write(out)
}

func (b Blog) RenderListGrid(page *pager.Page, gridMode bool, timeFmt string) string {
	colCount := 0
	out := "<gno-columns>\n"
	for _, item := range page.Items {
		if colCount%3 == 0 {
			out += "<gno-columns>\n"
		}
		post := item.Value.(*Post)
		out += post.RenderPreview(gridMode, timeFmt, b.Prefix)
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

// :post
func (b Blog) RenderPost(res *mux.ResponseWriter, req *mux.Request) {
	postSlug := req.GetVar("slug")
	if postSlug == "" {
		res.Write("Post slug is required.")
		return
	}

	foundKey, found := b.findPostBySlug(postSlug)
	if !found {
		res.Write("Post not found.")
		return
	}
	post, found := b.Posts.Get(foundKey)
	out := post.(*Post).RenderPost(b.Prefix)

	res.Write(out)
}

func (b Blog) RenderFilterHeader(req *mux.Request) string {
	out := md.H2(renderBreadcrumb(b.Prefix, req.Path)) + "\n"
	out += md.Link("▲ recent", GetSortByPublishedPath(b.Prefix+":"+req.RawPath, "asc")) + " | "
	out += md.Link("▼ oldest", GetSortByPublishedPath(b.Prefix+":"+req.RawPath, "desc")) + " | "
	out += md.Link("▲ most common", GetSortByCommonPath(b.Prefix+":"+req.RawPath, "asc")) + " | "
	out += md.Link("▼ least common", GetSortByCommonPath(b.Prefix+":"+req.RawPath, "desc")) + " | "

	sort := parseQueryParams("sort", req.RawPath)
	order := parseQueryParams("order", req.RawPath)
	if order == "" {
		order = "asc"
	}
	alphaOrder := "(A-Z)"
	if order == "desc" && sort == "alpha" {
		alphaOrder = "(Z-A)"
	}
	out += md.Link(orderArrow("alpha", sort, order)+" alphabetical "+alphaOrder,
		GetSortByAlphabeticalPath(b.Prefix+":"+req.RawPath, order)) + " | "
	out += md.Link("🗘 reset", b.Prefix+":"+req.Path) + "\n\n"
	return out
}

// :tags
func (b Blog) RenderTags(res *mux.ResponseWriter, req *mux.Request) {
	tag := req.GetVar("slug")
	out := b.RenderFilterHeader(req)
	options := parseRenderOptions(req.RawPath)

	var p *pager.Pager
	var page *pager.Page
	if tag == "" {
		if options.IsAlphabetical && options.HasOrder {
			p = pager.NewPager(b.TagsIndex, 12, options.Ascending)
		} else if options.IsCommon && options.HasOrder {
			p = pager.NewPager(b.TagsSorted, 12, options.Ascending)
		} else {
			p = pager.NewPager(b.TagsIndex, 12, options.NewestFirst)
		}
		page = p.MustGetPageByPath(req.RawPath)
		for _, t := range page.Items {
			split := strings.Split(t.Key, "::")
			if len(split) > 1 {
				out += md.H3(md.Link(split[1], b.Prefix+":tags/"+split[1])+" ("+split[0]+")") + "\n\n"
				continue
			}
			out += md.H3(md.Link(t.Key, b.Prefix+":tags/"+t.Key)+" ("+ufmt.Sprintf("%d", t.Value)+")") + "\n\n"
		}
		out += page.Picker(req.Path) + "\n\n"
		res.Write(out)
		return
	}

	filteredPosts, exists := b.filterPostsByField("tag", tag)
	p = pager.NewPager(filteredPosts, 4, options.NewestFirst)
	page = p.MustGetPageByPath(req.RawPath)
	for _, item := range page.Items {
		post := item.Value.(*Post)
		if hasField(post.Tags, tag) {
			out += post.RenderPreview(false, options.TimeFormat, b.Prefix) + "\n"
			exists = true
		}
	}
	if !exists {
		out += "No posts found for this tag."
	}
	out += page.Picker(req.Path) + "\n\n"
	res.Write(out)
}

// :authors
func (b Blog) RenderAuthors(res *mux.ResponseWriter, req *mux.Request) {
	author := req.GetVar("slug")
	out := b.RenderFilterHeader(req)
	options := parseRenderOptions(req.RawPath)

	var p *pager.Pager
	var page *pager.Page
	if author == "" {
		if options.IsAlphabetical && options.HasOrder {
			p = pager.NewPager(b.AuthorsIndex, 12, options.Ascending)
		} else if options.IsCommon && options.HasOrder {
			p = pager.NewPager(b.AuthorsSorted, 12, options.Ascending)
		} else {
			p = pager.NewPager(b.AuthorsIndex, 12, options.NewestFirst)
		}
		page = p.MustGetPageByPath(req.RawPath)
		for _, a := range page.Items {
			split := strings.Split(a.Key, "::")
			if len(split) > 1 {
				out += md.H3(md.Link(split[1], b.Prefix+":authors/"+split[1])+" ("+split[0]+")") + "\n\n"
				continue
			}
			out += md.H3(md.Link(a.Key, b.Prefix+":authors/"+a.Key)+" ("+ufmt.Sprintf("%d", a.Value)+")") + "\n\n"
		}
		out += page.Picker(req.Path) + "\n\n"
		res.Write(out)
		return
	}

	filteredPosts, exists := b.filterPostsByField("author", author)
	p = pager.NewPager(filteredPosts, 4, options.NewestFirst)
	page = p.MustGetPageByPath(req.RawPath)
	for _, item := range page.Items {
		post := item.Value.(*Post)
		out += post.RenderPreview(false, options.TimeFormat, b.Prefix) + "\n"
	}
	if !exists {
		out += "No posts found for this author."
	}
	out += page.Picker(req.Path)
	res.Write(out)
}

// :commenter
func (b Blog) RenderCommenters(res *mux.ResponseWriter, req *mux.Request) {
	commenter := req.GetVar("slug")
	out := b.RenderFilterHeader(req)
	// TO DO
	out += commenter
	res.Write(out)
}

func (b Blog) Mention(role, recipient string) string {
	if role == "author" {
		return md.Bold(md.Link("@"+recipient, b.Prefix+"/author:"+recipient))
	}
	if role == "commenter" {
		return md.Bold(md.Link("@"+recipient, b.Prefix+"/commenter:"+recipient))
	}
	if role == "tag" {
		return md.Bold(md.Link("#"+recipient, b.Prefix+"/tag:"+recipient))
	}
	return ""
}

func (b *Blog) addToIndex(post *Post) {
	for _, tag := range post.Tags {
		oldCount := 0
		if val, found := b.TagsIndex.Get(tag); found {
			oldCount = val.(int)
			b.TagsSorted.Remove(ufmt.Sprintf("%05d::%s", oldCount, tag))
		}
		newCount := oldCount + 1
		b.TagsIndex.Set(tag, newCount)
		b.TagsSorted.Set(ufmt.Sprintf("%05d::%s", newCount, tag), tag)
	}
	for _, author := range post.Authors {
		oldCount := 0
		if val, found := b.AuthorsIndex.Get(author); found {
			oldCount = val.(int)
			b.AuthorsSorted.Remove(ufmt.Sprintf("%05d::%s", oldCount, author))
		}
		newCount := oldCount + 1
		b.AuthorsIndex.Set(author, newCount)
		b.AuthorsSorted.Set(ufmt.Sprintf("%05d::%s", newCount, author), author)
	}
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

func (b Blog) findPostBySlug(value string) (string, bool) {
	var foundKey string
	var found bool
	b.Posts.Iterate("", "", func(k string, v interface{}) bool {
		post := v.(*Post)
		if post.Slug == value {
			foundKey = k
			found = true
		}
		return found
	})
	return foundKey, found
}
