package memberstore

import (
	"net/url"
	"strings"

	"gno.land/p/moul/md"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/avl/pager"
	"gno.land/p/nt/ufmt"
	"gno.land/p/samcrew/gauge"
	"gno.land/p/samcrew/urlfilter"
)

// RenderMembers returns the members list with tier filters and pagination.
func RenderMembers(path string, members MembersByTier) string {
	u, _ := url.Parse(path)
	mdFilters, items := urlfilter.ApplyFilters(u, members.Tree, "filter")
	var sb strings.Builder

	sb.WriteString("\n\n\n")
	sb.WriteString("\n\n\n" + gauge.Render(2, 1, "a", "#4caf50", false, 300))
	sb.WriteString("\n\n\n")

	sb.WriteString(md.Link("> Go to Tiers summary <", "/r/gov/dao/v3/memberstore") + "\n\n")
	sb.WriteString(md.Bold("Filter members by tiers:"))
	sb.WriteString(mdFilters + "\n")

	const pageSize = 14
	pager := pager.NewPager(items, pageSize, false)
	page := pager.MustGetPageByPath(path)

	sb.WriteString(renderMembersPages(page, items) + "\n")
	sb.WriteString(renderPagination(u, page))

	return sb.String()
}

// renderMembersPages returns the members of each page.
func renderMembersPages(page *pager.Page, members *avl.Tree) string {
	var sb strings.Builder

	sb.WriteString("| **Tier** | **Address** |\n")
	sb.WriteString("|----------|-------------|\n")

	for _, item := range page.Items {
		addr := item.Key
		tn, _ := members.Get(addr)
		tnStr, _ := tn.(string)

		sb.WriteString(ufmt.Sprintf("| %s %s | %s |\n", tierColoredChip(tnStr), tn, addr))
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
