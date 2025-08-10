package memberstore

import (
	"net/url"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/ufmt"
	"gno.land/p/samcrew/urlfilter"
)

// RenderMembers displays the full memberstore with tier filters and pagination, it writes into the provided string builder
func RenderMembers(path string, members MembersByTier, sb *strings.Builder) string {
	u, _ := url.Parse(path)
	query := u.Query()
	pageNum, _ := strconv.Atoi(query.Get("page"))
	if pageNum < 1 {
		pageNum = 1
	}
	pageSize := 14
	data := urlfilter.ApplyFilters(u, members.Tree)

	sb.WriteString("## Members:\n")
	sb.WriteString("Filter members by tier: \n")
	sb.WriteString(data.Markdown)
	sb.WriteString("\n")

	pager := pager.NewPager(data.Items, pageSize, false)
	page := pager.GetPageWithSize(pageNum, pageSize)
	renderMembersPage(sb, page, data.Items)
	renderPagination(sb, u, page)

	return sb.String()
}

// parsePageNum parses the page number from the query string, defaults to 1 if invalid, and returns it
func parsePageNum(s string) int {
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	return 1
}

// renderMembersPage writes the members of the current page into the provided string builder
func renderMembersPage(sb *strings.Builder, page *pager.Page, allMembers *avl.Tree) {
	for _, item := range page.Items {
		addr := item.Key
		tn, _ := allMembers.Get(addr)
		sb.WriteString(ufmt.Sprintf("- %v · %v\n", tn, addr))
	}
}

// renderPagination writes the pagination UI for the current page into the provided string builder
func renderPagination(sb *strings.Builder, u *url.URL, page *pager.Page) {
	sb.WriteString("\n")
	q := u.Query()
	q.Del("page")
	u.RawQuery = q.Encode()
	sb.WriteString(page.Picker(u.String()) + "\n\n")
}
