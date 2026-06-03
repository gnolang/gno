package showcase

import "testing"

func TestNewAndPublish(t *testing.T) {
	it := New("  Widget  ", 1500)
	if it.Title != "Widget" {
		t.Errorf("Title = %q, want %q", it.Title, "Widget")
	}
	if it.Status != StatusDraft {
		t.Errorf("Status = %q, want %q", it.Status, StatusDraft)
	}
	it.Publish()
	if it.Status != StatusActive {
		t.Errorf("after Publish, Status = %q, want %q", it.Status, StatusActive)
	}
}

func TestCatalogTotal(t *testing.T) {
	c := Catalog{
		"a": New("A", 100),
		"b": New("B", 250),
	}
	if got := c.Total(); got != 350 {
		t.Errorf("Total() = %d, want 350", got)
	}
}

func TestTagsHas(t *testing.T) {
	tags := Tags{"new", "featured"}
	if !tags.Has("new") {
		t.Error("Has(new) = false, want true")
	}
	if tags.Has("missing") {
		t.Error("Has(missing) = true, want false")
	}
}
