package blog

import (
	"std"
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
	PostId           seqid.ID
	Posts            *avl.Tree // id --> *Post
	PostsBySlug      *avl.Tree // slug --> *Post
	PostsByCreatedAt *avl.Tree // createdAt --> *Post
	PostsByUpdatedAt *avl.Tree // updatedAt --> *Post
	TagsIndex        *avl.Tree // tagName --> int
	TagsSorted       *avl.Tree // "<count>::tag" --> tag
	AuthorsIndex     *avl.Tree // authorAddress --> int
	AuthorsSorted    *avl.Tree // "<count>::author" --> author

	DisableLikes  bool
	PreviewFooter string // additional text, e.g. "via @lou" or txlink calls
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
		PostsByCreatedAt: avl.NewTree(),
		PostsByUpdatedAt: avl.NewTree(),
		PostId:           0,
		TagsIndex:        avl.NewTree(),
		TagsSorted:       avl.NewTree(),
		AuthorsIndex:     avl.NewTree(),
		AuthorsSorted:    avl.NewTree(),
		DisableLikes:     false,
		PreviewFooter:    "",
	}, nil
}

func (b *Blog) AddPost(post *Post) error {
	b.PostId.Next()
	post.ID = b.PostId
	var added bool
	if _, exists := b.PostsBySlug.Get(post.Slug); exists {
		return ErrPostAlreadyExists
	}
	post.SetDisableLikes(b.DisableLikes)

	b.addToIndex(post)
	added = b.Posts.Set(b.PostId.String(), post)
	added = b.PostsBySlug.Set(post.Slug, post)
	added = b.PostsByCreatedAt.Set(post.CreatedAt.String(), post)
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
	postByCreated, _ := b.PostsByCreatedAt.Get(post.CreatedAt.String())
	postByUpdated, _ := b.PostsByUpdatedAt.Get(post.UpdatedAt.String())
	UpdatePost(post, newPost)
	UpdatePost(postBySlug.(*Post), newPost)
	UpdatePost(postByCreated.(*Post), newPost)
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
	postByCreated, _ := b.PostsByCreatedAt.Get(post.CreatedAt.String())
	postByUpdated, _ := b.PostsByUpdatedAt.Get(post.UpdatedAt.String())
	UpdatePost(post, newPost)
	UpdatePost(postById.(*Post), newPost)
	UpdatePost(postByUpdated.(*Post), newPost)
	UpdatePost(postByCreated.(*Post), newPost)
	return nil
}

func UpdatePost(existingPost, newPost *Post) {
	existingPost.Slug = newPost.Slug
	existingPost.Title = newPost.Title
	existingPost.Body = newPost.Body
	existingPost.Tags = newPost.Tags
	existingPost.Authors = newPost.Authors
	existingPost.UpdatedAt = time.Now()
	// likes later
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
	_, removed = b.PostsByCreatedAt.Remove(post.CreatedAt.String())
	_, removed = b.PostsByUpdatedAt.Remove(post.UpdatedAt.String())
	if !removed {
		return ErrDeleteFailed
	}
	return nil
}

func (b *Blog) SetDisablePostLikes(disable bool) {
	b.DisableLikes = disable
	b.Posts.ReverseIterate("", "", func(_ string, value any) bool {
		post := value.(*Post)
		post.SetDisableLikes(disable)
		return false
	})
	b.PostsBySlug.ReverseIterate("", "", func(_ string, value any) bool {
		post := value.(*Post)
		post.SetDisableLikes(disable)
		return false
	})
	b.PostsByCreatedAt.ReverseIterate("", "", func(_ string, value any) bool {
		post := value.(*Post)
		post.SetDisableLikes(disable)
		return false
	})
	b.PostsByUpdatedAt.ReverseIterate("", "", func(_ string, value any) bool {
		post := value.(*Post)
		post.SetDisableLikes(disable)
		return false
	})
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
	router.HandleFunc("commenters", b.RenderCommenters) // TODO
	router.HandleFunc("commenters/{slug}", b.RenderCommenters)
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
		order = "desc"
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

	options := ParseRenderOptions(req.RawPath)
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

func (b Blog) RenderFilterHeader(req *mux.Request, isListing bool) string {
	out := md.H2(renderBreadcrumb(b.Prefix, req.Path)) + "\n"
	out += md.Link("▲ recent", GetSortByPublishedPath(b.Prefix+":"+req.RawPath, "asc")) + " | "
	out += md.Link("▼ oldest", GetSortByPublishedPath(b.Prefix+":"+req.RawPath, "desc")) + " | "
	out += md.Link("⏱ short", GetTimeFmtPath(b.Prefix+":"+req.RawPath, "short")) + ", "      // TODO
	out += md.Link("full", GetTimeFmtPath(b.Prefix+":"+req.RawPath, "full")) + ", "          // TODO
	out += md.Link("relative", GetTimeFmtPath(b.Prefix+":"+req.RawPath, "relative")) + " | " // TODO

	if !isListing {
		out += md.Link("⊞ grid", GetGridFmtPath(b.Prefix+":"+req.RawPath)) + " | " // TODO
		out += md.Link("☰ list", GetListFmtPath(b.Prefix+":"+req.RawPath)) + " | " // TODO
		out += md.Link(orderArrow("update", "update", "desc")+" last updated",     // TODO
			GetSortByUpdatedPath(b.Prefix+":"+req.RawPath, "desc")) + " | "
	}
	sort := parseQueryParams("sort", req.RawPath)
	order := parseQueryParams("order", req.RawPath)
	if order == "" {
		order = "desc"
	}

	commonOrder := " most common"
	if order == "desc" && sort == "common" {
		commonOrder = " least common"
	}
	if isListing {
		out += md.Link(orderArrow("common", sort, order)+commonOrder,
			GetSortByCommonPath(b.Prefix+":"+req.RawPath, order)) + " | "
	}

	alphaOrder := " (A-Z)"
	if order == "desc" && sort == "alpha" {
		alphaOrder = " (Z-A)"
	}
	out += md.Link(orderArrow("alpha", sort, order)+" alphabetical"+alphaOrder,
		GetSortByAlphabeticalPath(b.Prefix+":"+req.RawPath, order)) + " | "
	out += md.Link("🗘 reset", b.Prefix+":"+req.Path) + "\n\n"
	return out
}

// :tags
func (b Blog) RenderTags(res *mux.ResponseWriter, req *mux.Request) {
	tag := req.GetVar("slug")
	out := b.RenderFilterHeader(req, false)
	options := ParseRenderOptions(req.RawPath)

	if tag == "" {
		p, _ := b.determinePager(options, "tag", tag)
		page := p.MustGetPageByPath(req.RawPath)
		for _, t := range page.Items {
			if strings.Contains(t.Key, "::") {
				split := strings.Split(t.Key, "::")
				if len(split) > 1 {
					out += md.H3(md.Link(split[1], b.Prefix+":tags/"+split[1])+" ("+ufmt.Sprintf("%d", t.Value)+")") + "\n\n"
					continue
				}
			}
			out += md.H3(md.Link(t.Key, b.Prefix+":tags/"+t.Key)+" ("+ufmt.Sprintf("%d", t.Value)+")") + "\n\n"
		}
		out += page.Picker(req.Path) + "\n\n"
		res.Write(out)
		return
	}

	p, exists := b.determinePager(options, "tag", tag)
	page := p.MustGetPageByPath(req.RawPath)
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
	out := b.RenderFilterHeader(req, false)
	options := ParseRenderOptions(req.RawPath)

	if author == "" {
		p, _ := b.determinePager(options, "author", author)
		page := p.MustGetPageByPath(req.RawPath)
		for _, a := range page.Items {
			split := strings.Split(a.Key, "::")
			if len(split) > 1 {
				out += md.H3(md.Link(split[1], b.Prefix+":authors/"+split[1])+" ("+ufmt.Sprintf("%d", a.Value)+")") + "\n\n"
				continue
			}
			out += md.H3(md.Link(a.Key, b.Prefix+":authors/"+a.Key)+" ("+ufmt.Sprintf("%d", a.Value)+")") + "\n\n"
		}
		out += page.Picker(req.Path) + "\n\n"
		res.Write(out)
		return
	}

	p, exists := b.determinePager(options, "author", author)
	page := p.MustGetPageByPath(req.RawPath)
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

// :commenters
func (b Blog) RenderCommenters(res *mux.ResponseWriter, req *mux.Request) {
	commenter := req.GetVar("slug")
	out := b.RenderFilterHeader(req, false)
	comments, exists := b.filterPostsByField("commenter", commenter)
	if !exists {
		out += "No comments found for this commenter."
		res.Write(out)
		return
	}
	p := pager.NewPager(comments, 4, true) // TO DO
	page := p.MustGetPageByPath(req.RawPath)
	for _, comment := range page.Items {
		c := comment.Value.(*Comment)
		k := strings.Split(comment.Key, "::")[1]
		post, _ := b.Posts.Get(k)
		out += c.Content + "\n\n"
		out += md.Italic(formatTime(c.CreatedAt, "relative")) + "\n\n"
		out += "in " + (md.Link(post.(*Post).Title, b.Prefix+":posts/"+post.(*Post).Slug)) + "\n\n"
		out += md.HorizontalRule()
	}
	out += page.Picker(req.Path) + "\n\n"
	res.Write(out)
}

func (b Blog) Mention(role, recipient string) string {
	if role == "author" {
		return md.Bold(md.Link("@"+recipient, b.Prefix+"/authors:"+recipient))
	}
	if role == "commenter" {
		return md.Bold(md.Link("@"+recipient, b.Prefix+"/commenters:"+recipient))
	}
	if role == "tag" {
		return md.Bold(md.Link("#"+recipient, b.Prefix+"/tags:"+recipient))
	}
	return ""
}

func (b Blog) determinePager(options RenderOptions, field, value string) (*pager.Pager, bool) {
	if options.HasOrder && value == "" {
		if options.IsAlphabetical && field == "author" {
			return pager.NewPager(b.AuthorsIndex, 12, !options.Ascending), true
		} else if options.IsCommon && field == "author" {
			return pager.NewPager(b.AuthorsSorted, 12, options.Ascending), true
		} else if options.IsAlphabetical && field == "tag" {
			return pager.NewPager(b.TagsIndex, 12, !options.Ascending), true
		} else if options.IsCommon && field == "tag" {
			return pager.NewPager(b.TagsSorted, 12, options.Ascending), true
		} else if field == "author" {
			return pager.NewPager(b.AuthorsIndex, 12, options.NewestFirst), true
		} else if field == "tag" {
			return pager.NewPager(b.TagsIndex, 12, options.NewestFirst), true
		}
	} else if !options.HasOrder && value == "" {
		if field == "author" {
			return pager.NewPager(b.AuthorsIndex, 12, options.NewestFirst), true
		} else if field == "tag" {
			return pager.NewPager(b.TagsIndex, 12, options.NewestFirst), true
		}
	}
	// TODO: add grid mode
	if value != "" {
		filteredPosts, exists := b.filterPostsByField(field, value)
		return pager.NewPager(filteredPosts, 4, options.NewestFirst), exists
	}
	return nil, false
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
		b.TagsSorted.Set(ufmt.Sprintf("%05d::%s", newCount, tag), newCount)
	}
	for _, author := range post.Authors {
		oldCount := 0
		if val, found := b.AuthorsIndex.Get(author); found {
			oldCount = val.(int)
			b.AuthorsSorted.Remove(ufmt.Sprintf("%05d::%s", oldCount, author))
		}
		newCount := oldCount + 1
		b.AuthorsIndex.Set(author, newCount)
		b.AuthorsSorted.Set(ufmt.Sprintf("%05d::%s", newCount, author), newCount)
	}
}

func (b Blog) filterPostsStartEnd(tree *avl.Tree, start, end *time.Time) *avl.Tree {
	filtered := avl.NewTree()
	tree.Iterate("", "", func(k string, v interface{}) bool {
		post := v.(*Post)
		if (start == nil || post.CreatedAt.After(*start)) &&
			(end == nil || post.CreatedAt.Before(*end)) {
			filtered.Set(k, post)
		}
		return false
	})
	return filtered
}

func (b Blog) filterPostsByField(field, value string) (*avl.Tree, bool) {
	filteredPosts := avl.NewTree()
	commenterId := seqid.ID(0)
	b.Posts.ReverseIterate("", "", func(k string, v interface{}) bool {
		post := v.(*Post)
		if field == "tag" && hasField(post.Tags, value) {
			filteredPosts.Set(k, post)
		} else if field == "author" && hasField(post.Authors, value) {
			filteredPosts.Set(k, post)
		} else if field == "commenter" {
			comments := post.GetCommentsByAuthor(std.Address(value))
			if len(comments) == 0 {
				return false
			}
			for _, comment := range comments {
				commenterId.Next()
				filteredPosts.Set(commenterId.String()+"::"+post.ID.String(), comment)
			}
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
