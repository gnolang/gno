// Package urlfilter provides functionality to filter items based on URL query parameters.
// It is designed to work with an avl.Tree structure where each key represents a filter
// and each value is an avl.Tree containing items associated with that filter.
package urlfilter

import (
	"net/url"
	"strings"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ufmt/v0"
	"gno.land/p/sunspirit/md"
)

// ApplyFilters filters items based on the "filter" query parameter in the given URL
// and generates a Markdown representation of all available filters.
//
// Expected `items` structure:
//   - `items` is an *avl.Tree where each key is a filter name (e.g., "T1", "size:XL", "on_sale")
//     and each value is an *avl.Tree containing the items for that filter.
//   - Each item tree uses:
//     Key   (string): Unique item identifier
//     Value (any)   : Optional associated item data
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
//	mdFilters, items := ApplyFilters(u, filters, "filter")
//
//	// mdFilters	→ Markdown links for toggling filters
//	// items    	→ AVL tree containing the filtered items
func ApplyFilters(u *url.URL, items *avl.Tree, paramName string) (string, *avl.Tree) {
	active := parseFilterMap(u.Query(), paramName)
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
		q := toggleFilterQuery(active, f, allFilters, paramName)
		urlStr := buildURL(u.Path, q)
		sb.WriteString(ufmt.Sprintf(" | %v ", md.Link(formatLabel(f, active[f]), urlStr)))
	}

	return sb.String(), resultTree
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
func parseFilterMap(query url.Values, paramName string) map[string]bool {
	filterStr := strings.TrimSpace(query.Get(paramName))
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
func toggleFilterQuery(active map[string]bool, toggled string, all []string, paramName string) url.Values {
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
		q.Set(paramName, strings.Join(newFilters, ","))
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
