package pageable

import (
	"testing"
)

// MockPageable implements the Pageable interface for testing
type MockPageable struct {
	items []int
}

func (m *MockPageable) Size() int {
	return len(m.items)
}

func (m *MockPageable) IterateByOffset(offset int, count int, cb func(index interface{}, value interface{}) bool) bool {
	if count < 0 {
		// Handle reversed iteration
		count = -count                     // Make count positive
		start := len(m.items) - 1 - offset // Adjust start position from the end
		if start >= len(m.items) {
			start = len(m.items) - 1
		}
		end := start - count + 1
		if end < 0 {
			end = 0
		}
		for i := start; i >= end; i-- {
			if cb(i, m.items[i]) {
				return true
			}
		}
		return false
	}

	end := offset + count
	if end > len(m.items) {
		end = len(m.items)
	}

	for i := offset; i < end; i++ {
		if cb(i, m.items[i]) {
			return true
		}
	}
	return false
}

func TestPager(t *testing.T) {
	// Create test data
	mock := &MockPageable{
		items: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	tests := []struct {
		name        string
		pageSize    int
		pageNum     int
		wantItems   int
		wantTotal   int
		wantPages   int
		wantHasPrev bool
		wantHasNext bool
	}{
		{"first page", 3, 1, 3, 10, 4, false, true},
		{"middle page", 3, 2, 3, 10, 4, true, true},
		{"last page", 3, 4, 1, 10, 4, true, false},
		{"invalid page", 3, 5, 0, 10, 4, true, false},
		{"zero page", 3, 0, 0, 10, 4, false, true},
		{"all items", 10, 1, 10, 10, 1, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pager := NewPager(mock, tt.pageSize, false)
			page := pager.GetPage(tt.pageNum)

			if len(page.Items) != tt.wantItems {
				t.Errorf("got %d items, want %d", len(page.Items), tt.wantItems)
			}
			if page.TotalItems != tt.wantTotal {
				t.Errorf("got %d total items, want %d", page.TotalItems, tt.wantTotal)
			}
			if page.TotalPages != tt.wantPages {
				t.Errorf("got %d total pages, want %d", page.TotalPages, tt.wantPages)
			}
			if page.HasPrev != tt.wantHasPrev {
				t.Errorf("got HasPrev=%v, want %v", page.HasPrev, tt.wantHasPrev)
			}
			if page.HasNext != tt.wantHasNext {
				t.Errorf("got HasNext=%v, want %v", page.HasNext, tt.wantHasNext)
			}
		})
	}
}

func TestPicker(t *testing.T) {
	mock := &MockPageable{
		items: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	tests := []struct {
		name     string
		pageSize int
		pageNum  int
		want     string
	}{
		{"first page", 3, 1, "**1** | [2](?page=2) | [3](?page=3) | [4](?page=4)"},
		{"middle page", 3, 2, "[1](?page=1) | **2** | [3](?page=3) | [4](?page=4)"},
		{"last page", 3, 4, "[1](?page=1) | [2](?page=2) | [3](?page=3) | **4**"},
		{"single page", 10, 1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pager := NewPager(mock, tt.pageSize, false)
			page := pager.GetPage(tt.pageNum)
			got := page.Picker()
			if got != tt.want {
				t.Errorf("\ngot:  %s\nwant: %s", got, tt.want)
			}
		})
	}
}

func TestReversedPager(t *testing.T) {
	mock := &MockPageable{
		items: []int{1, 2, 3, 4, 5},
	}

	pager := NewPager(mock, 2, true)
	page := pager.GetPage(1)

	if len(page.Items) != 2 {
		t.Errorf("got %d items, want 2", len(page.Items))
	}

	// Check if items are in reverse order
	if page.Items[0].Value.(int) != 5 || page.Items[1].Value.(int) != 4 {
		t.Errorf("items not in reverse order: got %v", page.Items)
	}
}

func TestParseQuery(t *testing.T) {
	mock := &MockPageable{items: []int{1, 2, 3}}
	pager := NewPager(mock, 10, false)

	tests := []struct {
		name      string
		url       string
		wantPage  int
		wantSize  int
		wantError bool
	}{
		{"valid query", "/?page=2&size=5", 2, 5, false},
		{"missing params", "/", 1, 10, false},
		{"invalid page", "/?page=invalid", 1, 10, false},
		{"invalid size", "/?size=invalid", 1, 10, false},
		{"negative values", "/?page=-1&size=-5", 1, 10, false},
		{"invalid url", ":%invalid", 1, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, size, err := pager.ParseQuery(tt.url)
			if (err != nil) != tt.wantError {
				t.Errorf("got error=%v, want error=%v", err != nil, tt.wantError)
			}
			if page != tt.wantPage {
				t.Errorf("got page=%d, want %d", page, tt.wantPage)
			}
			if size != tt.wantSize {
				t.Errorf("got size=%d, want %d", size, tt.wantSize)
			}
		})
	}
}

func TestPager_UI_WithManyPages(t *testing.T) {
	mock := &MockPageable{
		items: make([]int, 100),
	}
	for i := 0; i < 100; i++ {
		mock.items[i] = i
	}

	pager := NewPager(mock, 10, false)

	tests := []struct {
		name     string
		pageNum  int
		pageSize int
		want     string
	}{
		{"first page", 1, 10, "**1** | [2](?page=2) | [3](?page=3) | … | [10](?page=10)"},
		{"second page", 2, 10, "[1](?page=1) | **2** | [3](?page=3) | [4](?page=4) | … | [10](?page=10)"},
		{"middle page", 5, 10, "[1](?page=1) | … | [3](?page=3) | [4](?page=4) | **5** | [6](?page=6) | [7](?page=7) | … | [10](?page=10)"},
		{"penultimate page", 9, 10, "[1](?page=1) | … | [7](?page=7) | [8](?page=8) | **9** | [10](?page=10)"},
		{"last page", 10, 10, "[1](?page=1) | … | [8](?page=8) | [9](?page=9) | **10**"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := pager.GetPageWithSize(tt.pageNum, tt.pageSize)
			got := page.Picker()
			if got != tt.want {
				t.Errorf("\ngot:  %s\nwant: %s", got, tt.want)
			}
		})
	}
}

func TestGetPageWithSize(t *testing.T) {
	mock := &MockPageable{
		items: []int{1, 2, 3, 4, 5},
	}

	pager := NewPager(mock, 10, false)

	tests := []struct {
		name      string
		pageNum   int
		pageSize  int
		wantLen   int
		wantFirst int
	}{
		{"custom size smaller", 1, 2, 2, 1},
		{"custom size exact", 1, 5, 5, 1},
		{"custom size larger", 1, 10, 5, 1},
		{"second page partial", 2, 3, 2, 4},
		{"empty page", 3, 3, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := pager.GetPageWithSize(tt.pageNum, tt.pageSize)
			if len(page.Items) != tt.wantLen {
				t.Errorf("got %d items, want %d", len(page.Items), tt.wantLen)
			}
			if tt.wantLen > 0 && page.Items[0].Value.(int) != tt.wantFirst {
				t.Errorf("first item got %d, want %d", page.Items[0].Value.(int), tt.wantFirst)
			}
		})
	}
}

func TestInvalidInputs(t *testing.T) {
	mock := &MockPageable{
		items: []int{1, 2, 3},
	}

	tests := []struct {
		name     string
		pageSize int
		pageNum  int
		wantLen  int
	}{
		{"zero page size", 0, 1, 0},
		{"negative page size", -1, 1, 0},
		{"negative page number", 10, -1, 0},
		{"too large page number", 10, 999, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pager := NewPager(mock, tt.pageSize, false)
			page := pager.GetPage(tt.pageNum)
			if len(page.Items) != tt.wantLen {
				t.Errorf("got %d items, want %d", len(page.Items), tt.wantLen)
			}
		})
	}
}

func TestEmptySource(t *testing.T) {
	mock := &MockPageable{items: []int{}}
	pager := NewPager(mock, 10, false)

	page := pager.GetPage(1)
	if page.TotalItems != 0 {
		t.Errorf("got total items %d, want 0", page.TotalItems)
	}
	if page.TotalPages != 0 {
		t.Errorf("got total pages %d, want 0", page.TotalPages)
	}
	if page.HasNext {
		t.Error("got HasNext=true, want false")
	}
	if page.HasPrev {
		t.Error("got HasPrev=true, want false")
	}
	if len(page.Items) != 0 {
		t.Errorf("got %d items, want 0", len(page.Items))
	}
}

func TestReversedPagerEdgeCases(t *testing.T) {
	mock := &MockPageable{
		items: []int{1, 2, 3, 4, 5},
	}
	pager := NewPager(mock, 2, true)

	tests := []struct {
		name     string
		pageNum  int
		pageSize int
		want     []int
	}{
		{"last partial page", 3, 2, []int{1}},
		{"oversized page", 1, 10, []int{5, 4, 3, 2, 1}},
		{"empty result", 4, 2, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page := pager.GetPageWithSize(tt.pageNum, tt.pageSize)
			got := make([]int, len(page.Items))
			for i, item := range page.Items {
				got[i] = item.Value.(int)
			}

			if len(got) != len(tt.want) {
				t.Errorf("got length %d, want length %d", len(got), len(tt.want))
				return
			}

			for i := 0; i < len(got); i++ {
				if got[i] != tt.want[i] {
					t.Errorf("at index %d: got %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestCustomQueryParams(t *testing.T) {
	mock := &MockPageable{items: []int{1, 2, 3}}
	pager := NewPager(mock, 10, false)
	pager.PageQueryParam = "p"
	pager.SizeQueryParam = "s"

	tests := []struct {
		name     string
		url      string
		wantPage int
		wantSize int
	}{
		{"custom params", "/?p=2&s=1", 2, 1},
		{"old params ignored", "/?page=3&size=2", 1, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, size, err := pager.ParseQuery(tt.url)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if page != tt.wantPage {
				t.Errorf("got page=%d, want %d", page, tt.wantPage)
			}
			if size != tt.wantSize {
				t.Errorf("got size=%d, want %d", size, tt.wantSize)
			}
		})
	}
}

func TestMustGetPageByPath(t *testing.T) {
	mock := &MockPageable{items: []int{1, 2, 3}}
	pager := NewPager(mock, 10, false)

	t.Run("panic on invalid url", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic but got none")
			}
		}()
		pager.MustGetPageByPath(":%invalid")
	})
}

func TestPickerWithPath(t *testing.T) {
	mock := &MockPageable{
		items: []int{1, 2, 3, 4, 5, 6},
	}

	tests := []struct {
		name     string
		pageSize int
		pageNum  int
		path     string
		want     string
	}{
		{"with query params", 2, 1, "/test?foo=bar&baz=qux", "**1** | [2](?page=2&baz=qux&foo=bar) | [3](?page=3&baz=qux&foo=bar)"},
		{"existing page param", 2, 2, "/test?page=999&other=value", "[1](?page=1&other=value) | **2** | [3](?page=3&other=value)"},
		{"no query params", 2, 3, "/test", "[1](?page=1) | [2](?page=2) | **3**"},
		{"empty path", 2, 2, "", "[1](?page=1) | **2** | [3](?page=3)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pager := NewPager(mock, tt.pageSize, false)
			page := pager.GetPageWithSize(tt.pageNum, tt.pageSize)
			got := page.PickerWithPath(tt.path)
			if got != tt.want {
				t.Errorf("\ngot:  %s\nwant: %s", got, tt.want)
			}
		})
	}
}
