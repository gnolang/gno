package avl_pager_params

import (
	"gno.land/p/moul/realmpath"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/pager"
	"gno.land/p/nt/markdown/sanitize/v0"
	"gno.land/p/nt/seqid/v0"
	"gno.land/p/nt/ufmt/v0"
)

// items holds the demo config items; p is a pre-built pager over it.
var (
	items *avl.Tree
	p     *pager.Pager
)

func init() {
	items = avl.NewTree()
	// Populate the tree with 15 sample items for demonstration.
	var idCounter seqid.ID
	for i := 1; i <= 15; i++ {
		id := idCounter.Next().String()
		items.Set(id, "Some item value: "+id)
	}
	// Build the pager once; Render() only calls methods on it.
	p = pager.NewPager(items, 5, false)
}

// Render routes the request:
//   - no path   → paginated home list
//   - non-empty path → single-item detail page
//
// The pager reads the `page` and `size` query params from the raw path;
// all other query params (e.g. sort, order) are preserved by Picker.
func Render(path string) string {
	// Parse splits route (before ?) from query params.
	req := realmpath.Parse(path)

	if req.Path == "" {
		return renderHome(path)
	}

	// req.Path is the item key, e.g. ":0000001" becomes "0000001".
	return renderConfigItem(req.Path)
}

// renderHome shows a paginated list of all config items.
func renderHome(fullPath string) string {
	page := p.MustGetPageByPath(fullPath)

	out := "# AVL Pager + Render paths\n\n"
	out += `This realm demonstrates how to combine realmpath routing with
avl pager URL-query pagination. A single Render() handles both
the paginated list (home) and per-item detail pages.

Static introductory content like this section stays constant
no matter which page of the list you are viewing.

`

	out += ufmt.Sprintf("Showing page %d of %d\n\n", page.PageNumber, page.TotalPages)

	for _, item := range page.Items {
		// Link each row to its detail page using a realm-relative path.
		out += ufmt.Sprintf("- [Item %s](/r/docs/avl_pager_params:%s)\n", item.Key, item.Key)
	}

	out += "\n" + page.Picker(fullPath) + "\n\n"
	out += "### [Go back to r/docs](/r/docs/home)"

	return out
}

// renderConfigItem shows the stored value for a single item key.
//
// itemName comes from the URL path and is user-controlled; sanitize
// before emitting into markdown to prevent inline injection.
func renderConfigItem(itemName string) string {
	safeKey := sanitize.InlineText(itemName)

	value, ok := items.Get(itemName)
	if !ok {
		return ufmt.Sprintf("**No item found** for key: %s", safeKey)
	}

	out := ufmt.Sprintf("# Item %s\n\n%s\n\n", safeKey, value.(string))
	out += "[Go back](/r/docs/avl_pager_params)"
	return out
}
