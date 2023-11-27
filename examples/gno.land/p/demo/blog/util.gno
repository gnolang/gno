package blog

import (
	"sort"
	"strings"
)

func breadcrumb(parts []string) string {
	return "# " + strings.Join(parts, " / ") + "\n\n"
}

func (p postSlice) GetPostBySlug(slug string) *Post {
	for _, post := range p {
		if post.Slug == slug {
			return post
		}
	}
	return nil
}

func (c commentSlice) RemoveByID(id int) (commentSlice, error) {
	for i, comment := range c {
		if comment.ID == id {
			return append(c[:i], c[i+1:]...), nil
		}
	}
	return nil, ErrNoSuchComment
}

// Sort interface implementations for postSlice & commentSlice

func (p postSlice) Len() int {
	return len(p)
}

func (p postSlice) Less(i, j int) bool {
	return p[i].CreatedAt.Before(p[j].CreatedAt)
}

func (p postSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (c commentSlice) Len() int {
	return len(c)
}

func (c commentSlice) Less(i, j int) bool {
	return c[i].CreatedAt.Before(c[j].CreatedAt)
}

func (c commentSlice) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}

var _ sort.Interface = postSlice(nil)
var _ sort.Interface = commentSlice(nil)
