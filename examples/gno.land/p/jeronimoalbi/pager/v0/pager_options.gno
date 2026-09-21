package pager

import "strings"

const (
	DefaultPageSize       = 50
	DefaultPageQueryParam = "page"
)

// PagerOption configures the pager.
type PagerOption func(*Pager)

// WithPageSize assigns a page size to a pager.
func WithPageSize(size int) PagerOption {
	return func(p *Pager) {
		if size < 1 {
			p.pageSize = DefaultPageSize
		} else {
			p.pageSize = size
		}
	}
}

// WithPageQueryParam assigns the name of the URL query param for the page value.
func WithPageQueryParam(name string) PagerOption {
	return func(p *Pager) {
		name = strings.TrimSpace(name)
		if name == "" {
			name = DefaultPageQueryParam
		}
		p.pageQueryParam = name
	}
}
