package memberstore

import (
	"net/url"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/ufmt"
	"gno.land/p/samcrew/urlfilter"
	"gno.land/p/sunspirit/md"
)

// RenderMembers returns the members list with tier filters and pagination.
func RenderMembers(path string, members MembersByTier) string {
	u, _ := url.Parse(path)
	data := urlfilter.ApplyFilters(u, members.Tree)
	var sb strings.Builder

	sb.WriteString(md.Link("> Go to Tiers summary <", "/r/gov/dao/v3/memberstore"))
	sb.WriteString(md.LineBreak(1))
	sb.WriteString(md.Bold("Filter members by tiers:"))
	sb.WriteString(data.Markdown)
	sb.WriteString(md.LineBreak(1))

	const pageSize = 14
	pager := pager.NewPager(data.Items, pageSize, false)
	page := pager.MustGetPageByPath(path)

	sb.WriteString(renderMembersPage(page, data.Items))
	sb.WriteString(md.LineBreak(1))
	sb.WriteString(renderPagination(u, page))

	return sb.String()
}

// renderMembersPage returns the members of the current page as bullet points.
func renderMembersPage(page *pager.Page, members *avl.Tree) string {
	var sb strings.Builder
	for _, item := range page.Items {
		addr := item.Key
		tn, _ := members.Get(addr)
		tnStr, _ := tn.(string)
		sb.WriteString(ufmt.Sprintf("- %v %v · %v\n", tierColoredChip(tnStr), tnStr, addr))
	}
	return sb.String()
}

// renderPagination returns the pagination UI for the current page.
func renderPagination(u *url.URL, page *pager.Page) string {
	q := u.Query()
	q.Del("page")
	u.RawQuery = q.Encode()

	var sb strings.Builder
	sb.WriteString(page.Picker(u.String()))

	return sb.String()
}
