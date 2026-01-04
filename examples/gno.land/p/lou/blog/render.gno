package blog

import (
	"net/url"
	"strings"

	"gno.land/p/lou/query"
	"gno.land/p/moul/md"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/avl/pager"
	"gno.land/p/nt/mux"
	"gno.land/p/nt/ufmt"
)

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
	rawPath = prefix + ":" + rawPath
	headerPresets := []HeaderPreset{
		{"mode", "grid", RenderModeToggle},
		{"time", "relative", RenderTimeFormat},
		{"alpha", "", RenderSortAlpha},
		{"recent", "", RenderSortRecent},
		{"update", "", RenderSortUpdate},
		{"time-range", "", func(p string) string { return RenderTimeRangeLinks(p, nil) }},
		{"reset", "", func(p string) string { return md.Link("⟳ reset", prefix) }},
	}
	return RenderHeader(rawPath, headerPresets)
}

// :posts
func (b Blog) RenderPosts(res *mux.ResponseWriter, req *mux.Request) {
	out := md.H1(b.Title()) + "\n"
	out += RenderBlogHeader(b.Prefix(), req.RawPath) + "\n\n"
	if b.Posts.Size() == 0 {
		res.Write(out + "No posts found.")
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
		out += post.RenderPreview(gridMode, timeFmt, b.Prefix())
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

// :posts/{slug}
func (b Blog) RenderPost(res *mux.ResponseWriter, req *mux.Request) {
	postSlug := req.GetVar("slug")
	foundKey, found := b.findPostBySlug(url.PathEscape(postSlug))
	if !found {
		res.Write("Post not found.")
		return
	}
	post, found := b.Posts.Get(foundKey)
	out := post.(*Post).RenderPost(b.Prefix())

	res.Write(out)
}

func (b Blog) RenderFilterHeader(req *mux.Request, user string, isListing bool) string {
	out := ""
	field := strings.Split(req.RawPath, "/")[0]
	switch field {
	case "commenters":
		out += md.H2(md.Link(b.Prefix(), b.Prefix()) + "/commenters/" + user)
		out += md.H4(md.Link("@"+user, "/u/"+user) + "'s profile")
	case "authors":
		out += md.H2(renderBreadcrumb(b.Prefix(), req.Path))
		if user != "" {
			out += md.H4(md.Link("@"+user, "/u/"+user) + "'s profile")
		}
	default:
		out += md.H2(renderBreadcrumb(b.Prefix(), req.Path))
	}

	var headerPresets []HeaderPreset
	if !isListing {
		headerPresets = append(headerPresets,
			HeaderPreset{"mode", "grid", RenderModeToggle},
			HeaderPreset{"time", "relative", RenderTimeFormat},
		)
	}
	headerPresets = append(headerPresets,
		HeaderPreset{"alpha", "", RenderSortAlpha},
		HeaderPreset{"recent", "", RenderSortRecent},
	)

	if isListing {
		headerPresets = append(headerPresets, HeaderPreset{"common", "", RenderSortCommon})
	} else {
		headerPresets = append(headerPresets, HeaderPreset{"update", "", RenderSortUpdate})
	}
	headerPresets = append(headerPresets,
		HeaderPreset{"time-range", "", func(p string) string { return RenderTimeRangeLinks(p, nil) }},
		HeaderPreset{"reset", "", func(p string) string { return md.Link("⟳ reset", b.Prefix()) }},
	)
	out += RenderHeader(b.Prefix()+":"+req.RawPath, headerPresets)
	return out + "\n\n"
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
	if query.GetQuery("mode", req.RawPath) == "" {
		options.IsGrid = false
	}
	if value == "" {
		return b.renderListingIndex(field, options, req)
	}
	user, err := CheckUser(value, b.UserResolver)
	if err == nil && user != "" {
		return b.renderFilteredListing(field, user, options, req)
	}
	return b.renderFilteredListing(field, value, options, req)
}

func (b Blog) renderListingIndex(field string, options RenderOptions, req *mux.Request) string {
	out := b.RenderFilterHeader(req, "", true)
	p, exists := b.determinePager(options, field, "")
	if !exists {
		return out + "No " + field + "s found."
	}
	page := p.MustGetPageByPath(req.RawPath)

	for _, item := range page.Items {
		label, count := extractListingDisplay(item.Key, item.Value)
		out += md.H3(md.Link(label, b.Prefix()+":"+field+"s/"+label)+" ("+ufmt.Sprintf("%d", count)+")") + "\n\n"
	}
	out += page.Picker(req.Path) + "\n\n"
	return out
}

func (b Blog) renderFilteredListing(field, value string, options RenderOptions, req *mux.Request) string {
	out := b.RenderFilterHeader(req, value, false)
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
			out += post.RenderPreview(options.IsGrid, options.TimeFormat, b.Prefix()) + "\n"
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
	var found bool
	if post, found = b.Posts.Get(postKey); !found {
		if post, found = b.PostsBySlug.Get(postKey); !found {
			post, found = b.PostsByUpdatedAt.Get(postKey)
		}
	}
	if !found {
		return ""
	}
	p := post.(*Post)

	user, _ := CheckUser(comment.Author(), b.UserResolver)
	out := md.H4(b.Mention("commenter", user)) + "\n\n"
	out += comment.Content() + "\n\n"
	out += md.Italic(formatTime(comment.CreatedAt(), options.TimeFormat)) + "\n\n"
	out += "in " + md.Link(p.Title(), b.Prefix()+":posts/"+p.Slug()) + "\n\n"
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

func orderArrow(targetSort, currentSort, order string) string {
	if targetSort != currentSort {
		return "↕"
	}
	if order == "asc" {
		return "▲"
	}
	return "▼"
}
