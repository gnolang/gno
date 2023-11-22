package blog

import "strings"

func breadcrumb(parts []string) string {
	return "# " + strings.Join(parts, " / ") + "\n\n"
}

func (p postSlice) GetPost(slug string) *Post {
	for _, post := range p {
		if post.Slug == slug {
			return post
		}
	}
	return nil
}

// Sort interface implementation for postSlice

func (p postSlice) Len() int {
	return len(p)
}

func (p postSlice) Less(i, j int) bool {
	return p[i].CreatedAt.Before(p[j].CreatedAt)
}

func (p postSlice) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
