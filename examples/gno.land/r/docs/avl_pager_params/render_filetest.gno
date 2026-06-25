// PKGPATH: gno.land/r/docs/avl_pager_params/filetests/render

package render

import "gno.land/r/docs/avl_pager_params"

func main() {
	// Home page: first page of the paginated list.
	println(avl_pager_params.Render(""))
	println("---")
	// Second page via query param.
	println(avl_pager_params.Render("?page=2"))
	println("---")
	// Item detail: existing key.
	println(avl_pager_params.Render("0000001"))
	println("---")
	// Unknown key: error fallback with sanitized path.
	println(avl_pager_params.Render("nonexistent"))
}

// Output:
// # AVL Pager + Render paths
//
// This realm demonstrates how to combine realmpath routing with
// avl pager URL-query pagination. A single Render() handles both
// the paginated list (home) and per-item detail pages.
//
// Static introductory content like this section stays constant
// no matter which page of the list you are viewing.
//
// Showing page 1 of 3
//
// - [Item 0000001](/r/docs/avl_pager_params:0000001)
// - [Item 0000002](/r/docs/avl_pager_params:0000002)
// - [Item 0000003](/r/docs/avl_pager_params:0000003)
// - [Item 0000004](/r/docs/avl_pager_params:0000004)
// - [Item 0000005](/r/docs/avl_pager_params:0000005)
//
// **1** | [2](?page=2) | [3](?page=3)
//
// ### [Go back to r/docs](/r/docs/home)
// ---
// # AVL Pager + Render paths
//
// This realm demonstrates how to combine realmpath routing with
// avl pager URL-query pagination. A single Render() handles both
// the paginated list (home) and per-item detail pages.
//
// Static introductory content like this section stays constant
// no matter which page of the list you are viewing.
//
// Showing page 2 of 3
//
// - [Item 0000006](/r/docs/avl_pager_params:0000006)
// - [Item 0000007](/r/docs/avl_pager_params:0000007)
// - [Item 0000008](/r/docs/avl_pager_params:0000008)
// - [Item 0000009](/r/docs/avl_pager_params:0000009)
// - [Item 000000a](/r/docs/avl_pager_params:000000a)
//
// [1](?page=1) | **2** | [3](?page=3)
//
// ### [Go back to r/docs](/r/docs/home)
// ---
// # Item 0000001
//
// Some item value: 0000001
//
// [Go back](/r/docs/avl_pager_params)
// ---
// **No item found** for key: nonexistent
