package urlfilter

import (
	"net/url"
	"strings"
	"testing"

	"gno.land/p/demo/avl"
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

	// --- Case 1: No filter selected, custom parameter name
	u, _ := url.Parse("/test")
	res := ApplyFilters(u, TreeParent, "custom-param-name")
	expectedMarkdown := " | [~~keyF1~~](/test?custom-param-name=keyF1)  | [~~keyF2~~](/test?custom-param-name=keyF2) "
	if res.Markdown != expectedMarkdown {
		t.Errorf("Expected Markdown %q, got %q", expectedMarkdown, res.Markdown)
	}
	// All items should be present
	count := 0
	res.Items.Iterate("", "", func(k string, _ interface{}) bool {
		count++
		return false
	})
	if count != 4 {
		t.Errorf("Expected 4 items, got %d", count)
	}
	// The parameter name should be "custom-param-name"
	if !strings.Contains(expectedMarkdown, "custom-param-name") {
		t.Errorf("Expected 'custom-param-name' parameter in Markdown, got %q", res.Markdown)
	}

	// --- Case 2: One filter active (keyF1)
	u, _ = url.Parse("/test?filter=keyF1")
	res = ApplyFilters(u, TreeParent)
	expectedMarkdown = " | [**keyF1**](/test)  | [~~keyF2~~](/test?filter=keyF1%2CkeyF2) "
	if res.Markdown != expectedMarkdown {
		t.Errorf("Expected Markdown %q, got %q", expectedMarkdown, res.Markdown)
	}
	// Only keyF1 items should be present
	keys := map[string]bool{}
	res.Items.Iterate("", "", func(k string, _ interface{}) bool {
		keys[k] = true
		return false
	})
	if len(keys) != 2 || !keys["key1_1"] || !keys["key1_2"] {
		t.Errorf("Unexpected items in filtered result: %#v", keys)
	}

	// --- Case 3: Multiple filters active (keyF1, keyF2)
	u, _ = url.Parse("/test?filter=keyF1,keyF2")
	res = ApplyFilters(u, TreeParent)
	// Both filters should be bold, no remove query for last one
	if res.Items.Size() != 4 {
		t.Errorf("Expected 4 items, got %d", res.Items.Size())
	}

	// --- Case 4: Filter not existing
	u, _ = url.Parse("/test?filter=unknown")
	res = ApplyFilters(u, TreeParent)
	// No matching items
	if res.Items.Size() != 0 {
		t.Errorf("Expected 0 items for unknown filter, got %d", res.Items.Size())
	}
}
