package memberstore

import (
	"net/url"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/ufmt"
)

// Common tiers list used throughout
var AllTiers = []string{T1, T2, T3}

// RenderMembers displays the full memberstore with tier filters and pagination, it writes into the provided string builder
func RenderMembers(path string, sb *strings.Builder) string {
	u, _ := url.Parse(path)
	query := u.Query()
	filter := parseFilterMap(query.Get("filter"))
	pageNum := parsePageNum(query.Get("page"))
	pageSize := 14
	allMembers := avl.NewTree()

	collectMembers(filter, allMembers)

	sb.WriteString("## Members:\n")
	renderTierFilters(sb, u, query, filter)

	pager := pager.NewPager(allMembers, pageSize, false)
	page := pager.GetPageWithSize(pageNum, pageSize)
	renderMembersPage(sb, page, allMembers)
	renderPagination(sb, u, query, page, "page")

	return sb.String()
}

// parsePageNum parses the page number from the query string, defaults to 1 if invalid, and returns it
func parsePageNum(s string) int {
	if n, err := strconv.Atoi(s); err == nil && n > 0 {
		return n
	}
	return 1
}

// parseFilterMap parses the filter string into a map of tier names, and returns it
func parseFilterMap(filter string) map[string]bool {
	m := map[string]bool{}
	for _, tn := range strings.Split(filter, ",") {
		trimTn := strings.TrimSpace(tn)
		if trimTn != "" {
			m[trimTn] = true
		}
	}
	return m
}

// collectMembers gathers the tiers and their summaries into a new avl.Tree
func collectMembers(filter map[string]bool, allMembers *avl.Tree) {
	members.Iterate("", "", func(tn string, ti interface{}) bool {
		tree, ok := ti.(*avl.Tree)
		if !ok {
			return false
		}
		tree.Iterate("", "", func(addr string, _ interface{}) bool {
			if len(filter) == 0 || filter[tn] {
				allMembers.Set(addr, tn)
			}
			return false
		})
		return false
	})
}

// renderTierFilters writes the tier filter hyperlinks into the provided string builder
func renderTierFilters(sb *strings.Builder, u *url.URL, query url.Values, filter map[string]bool) {
	sb.WriteString("Filter members by tier: \n")
	for _, tn := range AllTiers {
		newFilters := toggleFilter(filter, tn)

		// newQuery := copyQueryWithoutKey(query, "filter")
		// newQuery = copyQueryWithoutKey(newQuery, "page")
		newQuery := copyQueryWithoutKeys(query, []string{"filter", "page"})
		if len(newFilters) > 0 {
			newQuery.Set("filter", strings.Join(newFilters, ","))
		}

		urlStr := buildURL(u.Path, newQuery)
		label := "**" + tn + "**"
		if !filter[tn] {
			label = "~~" + tn + "~~"
		}
		sb.WriteString(ufmt.Sprintf(" | [%v](%v) ", label, urlStr))
	}
	sb.WriteString("\n")
}

// toggleFilter toggles the filter for a specific tier and returns the updated filter list
func toggleFilter(filter map[string]bool, toggleTn string) []string {
	result := []string{}
	for _, tn := range AllTiers {
		if tn == toggleTn {
			if !filter[tn] {
				result = append(result, tn)
			}
		} else if filter[tn] {
			result = append(result, tn)
		}
	}
	return result
}

// copyQueryWithoutKeys returns the query parameters, excluding the specified keys
func copyQueryWithoutKeys(query url.Values, keysToExclude []string) url.Values {
	exclude := map[string]bool{}
	for _, k := range keysToExclude {
		exclude[k] = true
	}
	newQuery := url.Values{}
	for k, v := range query {
		if !exclude[k] {
			for _, vv := range v {
				newQuery.Add(k, vv)
			}
		}
	}
	return newQuery
}

// buildURL returns an URL with the given path and query parameters
func buildURL(path string, query url.Values) string {
	if enc := query.Encode(); enc != "" {
		return path + "?" + enc
	}
	return path
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
func renderPagination(sb *strings.Builder, u *url.URL, query url.Values, page *pager.Page, pageKey string) {
	sb.WriteString("\n")
	baseQuery := copyQueryWithoutKeys(query, []string{pageKey})
	basePath := buildURL(u.Path, baseQuery)
	sb.WriteString(page.Picker(basePath))
	sb.WriteString("\n\n")
}
