package avl_pager_params

import (
	"gno.land/p/moul/realmpath"
	"gno.land/p/nt/avl"
	"gno.land/p/nt/avl/pager"
	"gno.land/p/nt/seqid"
	"gno.land/p/nt/ufmt"
)

// We'll keep some demo data in an AVL tree to showcase pagination.
var (
	items     *avl.Tree
	idCounter seqid.ID
)

func init() {
	items = avl.NewTree()
	// Populate the tree with 15 sample items for demonstration.
	for i := 1; i <= 15; i++ {
		id := idCounter.Next().String()
		items.Set(id, "Some item value: "+id)
	}
}

func Render(path string) string {
	// 1) Parse the incoming path to split route vs. query.
	req := realmpath.Parse(path)
	//    - req.Path contains everything *before* ? or $ (? - query params, $ - gnoweb params)
	//    - The remaining part (page=2, size=5, etc.) is not in req.Path.

	// 2) If no specific route is provided (req.Path == ""), we’ll show a “home” page
	//    that displays a list of configs in paginated form.
	if req.Path == "" {
		return renderHome(path)
	}

	// 3) If a route *is* provided (e.g. :SomeKey),
	//    we will interpret it as a request for a specific page.
	return renderConfigItem(req.Path)
}

// renderHome shows a paginated list of config items if route == "".
func renderHome(fullPath string) string {
	// Create a Pager for our config tree, with a default page size of 5.
	p := pager.NewPager(items, 5, false)

	// MustGetPageByPath uses the *entire* path (including query parts: ?page=2, etc.)
	page := p.MustGetPageByPath(fullPath)

	// Start building the output (plain text or markdown).
	out := "# AVL Pager + Render paths\n\n"
	out += `This realm showcases how to maintain a paginated list while properly parsing render paths. 
You can see how a single page can include a paginated element (like the example below), and how clicking 
an item can take you to a dedicated page for that specific item.

No matter how you browse through the paginated list, the introductory text (this section) remains the same.

`

	out += ufmt.Sprintf("Showing page %d of %d\n\n", page.PageNumber, page.TotalPages)

	// List items for this page.
	for _, item := range page.Items {
		// Link each item to a details page: e.g. ":Config01"
		out += ufmt.Sprintf("- [Item %s](/r/docs/avl_pager_params:%s)\n", item.Key, item.Key)
	}

	// Insert pagination controls (previous/next links, etc.).
	out += "\n" + page.Picker(fullPath) + "\n\n"
	out += "### [Go back to r/docs](/r/docs/home)"

	return out
}

// renderConfigItem shows details for a single item, e.g. ":item001".
func renderConfigItem(itemName string) string {
	value, ok := items.Get(itemName)
	if !ok {
		return ufmt.Sprintf("**No item found** for key: %s", itemName)
	}

	out := ufmt.Sprintf("# Item %s\n\n%s\n\n", itemName, value.(string))
	out += "[Go back](/r/docs/avl_pager_params)"
	return out
}
