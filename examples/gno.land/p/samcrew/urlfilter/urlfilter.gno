package urlfilter

import (
	"net/url"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/p/sunspirit/md"
)

// FilterItemsTree is a type alias for an AVL tree that holds filter items.
type FilterItemsTree = *avl.Tree

// Result holds the outcome of filtering operations.
type Result struct {
	Markdown string          // Filters represented as a Markdown string (for UI rendering)
	Items    FilterItemsTree // The filtered set of items
}

// ApplyFilters filters items based on the "filter" query parameter in the given URL
// and generates a Markdown representation of all available filters.
//
// Expected `items` structure:
//   - items is an *avl.Tree where:
//     Key   (string): The filter name (e.g., "T1", "size:XL", "on_sale")
//     Value (*avl.Tree): A nested tree containing the actual items for that filter
//   - Key   (string): Unique item identifier
//   - Value (any)   : Associated item data (optional)
//
// Example:
//
//	// Build the main filters tree
//	filters := avl.NewTree()
//
//	// Subtree for filter "T1"
//	t1Items := avl.NewTree()
//	t1Items.Set("item1", nil)
//	t1Items.Set("item2", nil)
//	filters.Set("T1", t1Items)
//
//	// URL with active filter "T1"
//	u, _ := url.Parse("/shop?filter=T1")
//
//	result := ApplyFilters(u, filters)
//
//	// result.Markdown → Markdown links for toggling filters
//	// result.Items    → AVL tree containing only the filtered items
func ApplyFilters(u *url.URL, items FilterItemsTree) Result {
	active := parseFilterMap(u.Query())
	allFilters := make([]string, 0)
	resultTree := avl.NewTree()

	// Iterate over each filter group in the items tree
	items.Iterate("", "", func(filterKey string, subtree interface{}) bool {
		allFilters = append(allFilters, filterKey)

		tree, ok := subtree.(*avl.Tree)
		if !ok {
			return false
		}

		// Add items to result if there are no active filters
		// or if the current filter is active
		tree.Iterate("", "", func(itemKey string, _ interface{}) bool {
			if len(active) == 0 || active[filterKey] {
				resultTree.Set(itemKey, filterKey)
			}
			return false
		})
		return false
	})

	// Build Markdown links for toggling each filter
	var sb strings.Builder
	for _, f := range allFilters {
		q := toggleFilterQuery(active, f, allFilters)
		urlStr := buildURL(u.Path, q)
		sb.WriteString(ufmt.Sprintf(" | %v ", md.Link(formatLabel(f, active[f]), urlStr)))
	}

	return Result{
		Markdown: sb.String(),
		Items:    resultTree,
	}
}

// buildURL returns a path + query string, omitting the "?" if no query exists.
func buildURL(path string, query url.Values) string {
	if enc := query.Encode(); enc != "" {
		return path + "?" + enc
	}
	return path
}

// parseFilterMap reads the "filter" query parameter and converts it into a map
// where keys are filter names and values are true for active filters.
//
// Example:
//
//	"filter=T1,T2" -> map[string]bool{"T1": true, "T2": true}
func parseFilterMap(query url.Values) map[string]bool {
	filterStr := strings.TrimSpace(query.Get("filter"))
	if filterStr == "" {
		return map[string]bool{}
	}
	m := make(map[string]bool)
	for _, f := range strings.Split(filterStr, ",") {
		if f = strings.TrimSpace(f); f != "" {
			m[f] = true
		}
	}
	return m
}

// toggleFilterQuery returns a new query string with the given filter toggled.
// - If the filter is currently active, it will be removed.
// - If it is inactive, it will be added.
// The order of filters follows the `all` list for consistency.
func toggleFilterQuery(active map[string]bool, toggled string, all []string) url.Values {
	newFilters := []string{}
	for _, f := range all {
		if f == toggled {
			if !active[f] { // Add if it was inactive
				newFilters = append(newFilters, f)
			}
		} else if active[f] { // Keep other active filters
			newFilters = append(newFilters, f)
		}
	}
	q := url.Values{}
	if len(newFilters) > 0 {
		q.Set("filter", strings.Join(newFilters, ","))
	}
	return q
}

// formatLabel returns the Markdown-formatted label for a filter,
// showing active filters in bold (**filter**) and inactive filters
// with strikethrough (~~filter~~).
func formatLabel(name string, active bool) string {
	if active {
		return md.Bold(name)
	}
	return md.Strikethrough(name)
}
