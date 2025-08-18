package memberstore

import (
	"net/url"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/ufmt"
	"gno.land/p/samcrew/urlfilter"
	"gno.land/p/sunspirit/md"
	"gno.land/p/sunspirit/table"
)

// RenderMembers returns the members list with tier filters and pagination.
func RenderMembers(path string, members MembersByTier) string {
	u, _ := url.Parse(path)
	mdFilters, items := urlfilter.ApplyFilters(u, members.Tree, "filter")
	var sb strings.Builder

	sb.WriteString(md.Link("> Go to Tiers summary <", "/r/gov/dao/v3/memberstore"))
	sb.WriteString(md.LineBreak(1))
	sb.WriteString(md.Bold("Filter members by tiers:"))
	sb.WriteString(mdFilters)
	sb.WriteString(md.LineBreak(1))

	const pageSize = 14
	pager := pager.NewPager(items, pageSize, false)
	page := pager.MustGetPageByPath(path)

	sb.WriteString(renderMembersPages(page, items))
	sb.WriteString(md.LineBreak(1))
	sb.WriteString(renderPagination(u, page))

	return sb.String()
}

// renderMembersPages returns the members of each page.
func renderMembersPages(page *pager.Page, members *avl.Tree) string {
	var sb strings.Builder
	mdBuilder := md.NewBuilder()
	tb, _ := table.New([]string{"Tier", "Address"}, [][]string{})

	for _, item := range page.Items {
		addr := item.Key
		tn, _ := members.Get(addr)
		tnStr, _ := tn.(string)

		tb.AddRow([]string{
			ufmt.Sprintf(
				"%s %s",
				tn,
				tierColoredChip(tnStr),
			),
			addr,
		})
	}

	mdBuilder.Add(md.Table(tb))
	sb.WriteString(mdBuilder.Render("\n"))

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
