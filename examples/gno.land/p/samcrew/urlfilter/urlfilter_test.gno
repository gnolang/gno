package urlfilter

import (
	"net/url"
	"strings"
	"testing"

	"gno.land/p/nt/avl/v0"
)

func buildTestTree() *avl.Tree {
	root := avl.NewTree()

	// keyF1 subtree
	t1 := avl.NewTree()
	t1.Set("key1_1", nil)
	t1.Set("key1_2", nil)
	root.Set("keyF1", t1)

	// keyF2 subtree
	t2 := avl.NewTree()
	t2.Set("key2_1", nil)
	t2.Set("key2_2", nil)
	root.Set("keyF2", t2)

	return root
}

func TestApplyFilters(t *testing.T) {
	TreeParent := buildTestTree()

	// --- Case 1: No filter selected
	u, _ := url.Parse("/test")
	mdFilters, items := ApplyFilters(u, TreeParent, "filter")
	expectedMarkdown := " | [~~keyF1~~](/test?filter=keyF1)  | [~~keyF2~~](/test?filter=keyF2) "
	if mdFilters != expectedMarkdown {
		t.Errorf("Expected Markdown %q, got %q", expectedMarkdown, mdFilters)
	}
	// All items should be present
	count := 0
	items.Iterate("", "", func(k string, _ interface{}) bool {
		count++
		return false
	})
	if count != 4 {
		t.Errorf("Expected 4 items, got %d", count)
	}

	// Try using ApplyFilters with a different name.
	withCustom, _ := ApplyFilters(u, TreeParent, "custom-param-name")
	if !strings.Contains(withCustom, "custom-param-name=keyF1") {
		t.Errorf("Expected 'custom-param-name' parameter in Markdown, got %q", mdFilters)
	}

	// --- Case 2: One filter active (keyF1)
	u, _ = url.Parse("/test?filter=keyF1")
	mdFilters, items = ApplyFilters(u, TreeParent, "filter")
	expectedMarkdown = " | [**keyF1**](/test)  | [~~keyF2~~](/test?filter=keyF1%2CkeyF2) "
	if mdFilters != expectedMarkdown {
		t.Errorf("Expected Markdown %q, got %q", expectedMarkdown, mdFilters)
	}
	// Only keyF1 items should be present
	keys := map[string]bool{}
	items.Iterate("", "", func(k string, _ interface{}) bool {
		keys[k] = true
		return false
	})
	if len(keys) != 2 || !keys["key1_1"] || !keys["key1_2"] {
		t.Errorf("Unexpected items in filtered result: %#v", keys)
	}

	// --- Case 3: Multiple filters active (keyF1, keyF2)
	u, _ = url.Parse("/test?filter=keyF1,keyF2")
	mdFilters, items = ApplyFilters(u, TreeParent, "filter")
	// Both filters should be bold, no remove query for last one
	if items.Size() != 4 {
		t.Errorf("Expected 4 items, got %d", items.Size())
	}

	// --- Case 4: Filter not existing
	u, _ = url.Parse("/test?filter=unknown")
	mdFilters, items = ApplyFilters(u, TreeParent, "filter")
	// No matching items
	if items.Size() != 0 {
		t.Errorf("Expected 0 items for unknown filter, got %d", items.Size())
	}
}
