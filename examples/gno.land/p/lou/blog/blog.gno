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
	PostId           seqid.ID
	Posts            *avl.Tree // id --> *Post
	PostsBySlug      *avl.Tree // slug --> *Post
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
	existingPost.Slug = newPost.Slug
	existingPost.Title = newPost.Title
	existingPost.Body = newPost.Body
	existingPost.Authors = newPost.Authors
	existingPost.Tags = newPost.Tags
	existingPost.CommentId = newPost.CommentId
	existingPost.Comments = newPost.Comments
	existingPost.Likes = newPost.Likes
	existingPost.DisableLikes = newPost.DisableLikes
	existingPost.PreviewFooter = newPost.PreviewFooter
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
	b.PostsByUpdatedAt.ReverseIterate("", "", func(_ string, value any) bool {
		post := value.(*Post)
		post.SetDisableLikes(disable)
		return false
	})
}

func (b Blog) Mention(role, recipient string) string {
	if role == "author" {
		return md.Bold(md.Link("@"+recipient, b.Prefix+":authors/"+recipient))
	}
	if role == "commenter" {
		return md.Bold(md.Link("@"+recipient, b.Prefix+":commenters/"+recipient))
	}
	if role == "tag" {
		return md.Bold(md.Link("#"+recipient, b.Prefix+":tags/"+recipient))
	}
	return ""
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
	router.HandleFunc("commenters", b.RenderCommenters)
	router.HandleFunc("commenters/{slug}", b.RenderCommenters)
	return router.Render(path)
}

func RenderBlogHeader(prefix, rawPath string) string {
	out := md.Link("⊞ grid", GetGridFmtPath(rawPath)) + " | "
	out += md.Link("☰ list", GetListFmtPath(rawPath)) + " | "

	// ?time=relative | full | short
	time := parseQueryParams("time", rawPath)
	if time == "" {
		time = "relative"
	}
	out += md.Link("⏱ "+time, GetTimeFmtPath(rawPath, toggleTimeFormat(time))) + " | "

	sort := parseQueryParams("sort", rawPath)
	order := parseQueryParams("order", rawPath)
	if order == "" {
		order = "desc"
	}
	// ?sort=alpha
	alphaOrder := "(A-Z)"
	if order == "desc" && sort == "alpha" {
		alphaOrder = "(Z-A)"
	}
	out += md.Link(orderArrow("alpha", sort, order)+" alphabetical "+alphaOrder,
		GetSortByAlphabeticalPath(rawPath, order)) + " | "
	// ?sort=recent
	publishedOrder := " recent"
	if order == "desc" && sort == "recent" {
		publishedOrder = " oldest"
	}
	out += md.Link(orderArrow("recent", sort, order)+publishedOrder,
		GetSortByPublishedPath(rawPath, order)) + " | "
	// ?sort=update
	out += md.Link(orderArrow("update", sort, order)+" last updated",
		GetSortByUpdatedPath(rawPath, order)) + " | "
	// ?start=<date>&end=<date>
	out += renderTimeRangeLinks(rawPath) + " | "

	out += md.Link("🗘 reset", prefix) + "\n\n"
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
		pageSize = 3
	}
	order := options.Ascending
	if options.IsAlphabetical {
		order = !order
	}

	var tree *avl.Tree
	switch {
	case options.IsLastUpdated:
		tree = b.PostsByUpdatedAt
	case options.IsAlphabetical:
		tree = b.PostsBySlug
	default:
		tree = b.Posts
	}
	if options.StartTime != nil || options.EndTime != nil {
		tree = b.filterPostsStartEnd(tree, options.StartTime, options.EndTime)
	}

	p := pager.NewPager(tree, pageSize, order)
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
	out := ""
	if strings.Split(req.RawPath, "/")[0] == "commenters" {
		out += md.H2(md.Link(b.Prefix, b.Prefix) + "/commenters/" + strings.Split(req.RawPath, "/")[1])
	} else {
		out += md.H2(renderBreadcrumb(b.Prefix, req.Path)) + "\n"
	}
	if !isListing {
		out += md.Link("⊞ grid", GetGridFmtPath(b.Prefix+":"+req.RawPath)) + " | "
		out += md.Link("☰ list", GetListFmtPath(b.Prefix+":"+req.RawPath)) + " | "

		// ?time=relative | short | full
		time := parseQueryParams("time", req.RawPath)
		if time == "" {
			time = "relative"
		}
		out += md.Link("⏱ "+time, GetTimeFmtPath(b.Prefix+":"+req.RawPath, toggleTimeFormat(time))) + " | "

	}

	sort := parseQueryParams("sort", req.RawPath)
	order := parseQueryParams("order", req.RawPath)
	if order == "" {
		order = "desc"
	}
	// ?sort=alpha
	alphaOrder := " (A-Z)"
	if order == "desc" && sort == "alpha" {
		alphaOrder = " (Z-A)"
	}
	out += md.Link(orderArrow("alpha", sort, order)+" alphabetical"+alphaOrder,
		GetSortByAlphabeticalPath(b.Prefix+":"+req.RawPath, order)) + " | "
	// ?sort=recent
	publishedOrder := " recent"
	if order == "desc" && sort == "recent" {
		publishedOrder = " oldest"
	}
	out += md.Link(orderArrow("recent", sort, order)+publishedOrder,
		GetSortByPublishedPath(b.Prefix+":"+req.RawPath, order)) + " | "

	if isListing {
		// ?sort=common
		commonOrder := " most common"
		if order == "desc" && sort == "common" {
			commonOrder = " least common"
		}
		out += md.Link(orderArrow("common", sort, order)+commonOrder,
			GetSortByCommonPath(b.Prefix+":"+req.RawPath, order)) + " | "
	} else {
		// ?sort=update
		out += md.Link(orderArrow("update", sort, order)+" last updated",
			GetSortByUpdatedPath(b.Prefix+":"+req.RawPath, order)) + " | "
	}
	// ?start=<date>&end=<date>
	out += renderTimeRangeLinks(req.RawPath) + " | "
	out += md.Link("🗘 reset", b.Prefix+":"+req.Path) + "\n\n"
	return out
}

// :tags
func (b Blog) RenderTags(res *mux.ResponseWriter, req *mux.Request) {
	out := b.RenderListings("tag", req.GetVar("slug"), req)
	res.Write(out)
}

// :authors
func (b Blog) RenderAuthors(res *mux.ResponseWriter, req *mux.Request) {
	out := b.RenderListings("author", req.GetVar("slug"), req)
	res.Write(out)
}

// :commenters
func (b Blog) RenderCommenters(res *mux.ResponseWriter, req *mux.Request) {
	if req.GetVar("slug") == "" {
		res.Write("Commenter slug is required.")
		return
	}
	out := b.RenderListings("commenter", req.GetVar("slug"), req)
	res.Write(out)
}

func (b Blog) RenderListings(field, value string, req *mux.Request) string {
	options := ParseRenderOptions(req.RawPath)
	options.IsGrid = false
	if value == "" {
		return b.renderListingIndex(field, options, req)
	}
	return b.renderFilteredListing(field, value, options, req)
}

func (b Blog) renderListingIndex(field string, options RenderOptions, req *mux.Request) string {
	out := b.RenderFilterHeader(req, true)
	p, exists := b.determinePager(options, field, "")
	if !exists {
		return out + "No " + field + "s found."
	}
	page := p.MustGetPageByPath(req.RawPath)

	for _, item := range page.Items {
		label, count := extractListingDisplay(item.Key, item.Value)
		out += md.H3(md.Link(label, b.Prefix+":"+field+"s/"+label)+" ("+ufmt.Sprintf("%d", count)+")") + "\n\n"
	}
	out += page.Picker(req.Path) + "\n\n"
	return out
}

func (b Blog) renderFilteredListing(field, value string, options RenderOptions, req *mux.Request) string {
	out := b.RenderFilterHeader(req, false)
	p, exists := b.determinePager(options, field, value)
	if !exists {
		out += "No posts found for this " + field + "."
	}
	page := p.MustGetPageByPath(req.RawPath)
	out += "<gno-columns>\n"
	colCount := 0
	for _, item := range page.Items {
		if colCount%3 == 0 {
			out += "<gno-columns>\n"
		}
		colCount++
		if field == "commenter" {
			out += b.renderCommenterItem(item, options)
		} else {
			post := item.Value.(*Post)
			out += post.RenderPreview(options.IsGrid, options.TimeFormat, b.Prefix) + "\n"
		}
		if colCount%3 == 0 {
			out += "</gno-columns>\n"
		} else if options.IsGrid {
			out += "|||\n"
		}
	}
	if colCount%3 != 0 {
		out += "</gno-columns>\n"
	}
	out += page.Picker(req.Path) + "\n\n"
	return out
}

func (b Blog) renderCommenterItem(item pager.Item, options RenderOptions) string {
	comment := item.Value.(*Comment)
	parts := strings.Split(item.Key, "::")
	if len(parts) < 2 {
		return ""
	}
	postKey := parts[1]

	var post interface{}
	var ok bool
	if post, ok = b.Posts.Get(postKey); !ok {
		if post, ok = b.PostsBySlug.Get(postKey); !ok {
			post, ok = b.PostsByUpdatedAt.Get(postKey)
		}
	}
	if !ok {
		return ""
	}
	p := post.(*Post)

	out := md.H4(b.Mention("commenter", comment.Author)) + "\n\n"
	out += comment.Content + "\n\n"
	out += md.Italic(formatTime(comment.CreatedAt, options.TimeFormat)) + "\n\n"
	out += "in " + md.Link(p.Title, b.Prefix+":posts/"+p.Slug) + "\n\n"
	out += md.HorizontalRule()
	return out
}

func (b Blog) determinePager(options RenderOptions, field, value string) (*pager.Pager, bool) {
	if value == "" {
		if options.IsAlphabetical && field == "author" {
			return pager.NewPager(b.AuthorsIndex, 12, !options.Ascending), true
		} else if options.IsCommon && field == "author" {
			return pager.NewPager(b.AuthorsSorted, 12, options.Ascending), true
		} else if options.IsAlphabetical && field == "tag" {
			return pager.NewPager(b.TagsIndex, 12, !options.Ascending), true
		} else if options.IsCommon && field == "tag" {
			return pager.NewPager(b.TagsSorted, 12, options.Ascending), true
		} else if field == "author" {
			return pager.NewPager(b.AuthorsIndex, 12, options.Ascending), true
		} else if field == "tag" {
			return pager.NewPager(b.TagsIndex, 12, options.Ascending), true
		}
	}
	pageSize := 9
	if !options.IsGrid {
		pageSize = 3
	}

	filteredPosts, exists := b.filterPostsByField(field, value, options.Sort)
	if options.StartTime != nil || options.EndTime != nil {
		filteredPosts = b.filterPostsStartEnd(filteredPosts, options.StartTime, options.EndTime)
	}
	if options.Sort == "alpha" {
		return pager.NewPager(filteredPosts, pageSize, !options.Ascending), exists
	}
	return pager.NewPager(filteredPosts, pageSize, options.Ascending), exists
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

func (b Blog) filterPostsByField(field, value, sort string) (*avl.Tree, bool) {
	commenterId := seqid.ID(0)
	recentPosts := avl.NewTree()
	alphaPosts := avl.NewTree()
	updatedPosts := avl.NewTree()

	switch field {
	case "tag", "author":
		b.Posts.ReverseIterate("", "", func(k string, v interface{}) bool {
			post := v.(*Post)
			var match bool
			if field == "tag" {
				match = hasField(post.Tags, value)
			} else {
				match = hasField(post.Authors, value)
			}
			if match {
				recentPosts.Set(k, post)
				alphaPosts.Set(post.Slug, post)
				updatedPosts.Set(post.UpdatedAt.String(), post)
			}
			return false
		})

	case "commenter":
		b.Posts.ReverseIterate("", "", func(k string, v interface{}) bool {
			post := v.(*Post)
			comments := post.GetCommentsByAuthor(value)
			for _, comment := range comments {
				commenterId.Next()
				keyPrefix := commenterId.String() + "::"
				recentPosts.Set(keyPrefix+post.ID.String(), comment)
				alphaPosts.Set(keyPrefix+post.Slug, comment)
				updatedPosts.Set(keyPrefix+post.UpdatedAt.String(), comment)
			}
			return false
		})
	}

	switch sort {
	case "alpha":
		return alphaPosts, alphaPosts.Size() > 0
	case "update":
		return updatedPosts, updatedPosts.Size() > 0
	default:
		return recentPosts, recentPosts.Size() > 0
	}
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
