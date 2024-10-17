package ui

import (
	"strings"
	"testing"
)

func TestPagerSinglePage(t *testing.T) {
	pager := Pager{
		CurrentPage: 1,
		TotalPages:  1,
		PathPrefix:  "/articles/page/",
	}

	dom := DOM{Pager: pager}
	result := dom.Pager.String(dom)

	if result != "" {
		t.Errorf("Expected no output for a single page, got %s", result)
	}
}

func TestPagerMultiplePages(t *testing.T) {
	pager := Pager{
		CurrentPage: 2,
		TotalPages:  5,
		PathPrefix:  "/articles/page/",
	}

	dom := DOM{Pager: pager}
	result := dom.Pager.String(dom)

	// Verify the presence of the Previous link
	expectedPrevLink := `/articles/page/1`
	if !strings.Contains(result, expectedPrevLink) {
		t.Errorf("Expected pager to contain link %s, got %s", expectedPrevLink, result)
	}

	// Verify the presence of the active page link (CurrentPage 2)
	expectedActiveLink := `<li class="page-item active"><span class="page-link">2</span></li>`
	if !strings.Contains(result, expectedActiveLink) {
		t.Errorf("Expected pager to contain active link %s, got %s", expectedActiveLink, result)
	}

	// Verify the presence of the Next link
	expectedNextLink := `/articles/page/3`
	if !strings.Contains(result, expectedNextLink) {
		t.Errorf("Expected pager to contain link %s, got %s", expectedNextLink, result)
	}
}


func TestPagerFirstPage(t *testing.T) {
	pager := Pager{
		CurrentPage: 1,
		TotalPages:  5,
		PathPrefix:  "/articles/page/",
	}

	dom := DOM{Pager: pager}
	result := dom.Pager.String(dom)

	unexpected := `aria-label="Previous"`
	if strings.Contains(result, unexpected) {
		t.Errorf("Did not expect %s, got %s", unexpected, result)
	}

	expected := `<li class="page-item active"><span class="page-link">1</span></li>`
	if !strings.Contains(result, expected) {
		t.Errorf("Expected active link for first page, got %s", result)
	}
}

func TestPagerLastPage(t *testing.T) {
	pager := Pager{
		CurrentPage: 5,
		TotalPages:  5,
		PathPrefix:  "/articles/page/",
	}

	dom := DOM{Pager: pager}
	result := dom.Pager.String(dom)

	unexpected := `aria-label="Next"`
	if strings.Contains(result, unexpected) {
		t.Errorf("Did not expect %s, got %s", unexpected, result)
	}

	expected := `<li class="page-item active"><span class="page-link">5</span></li>`
	if !strings.Contains(result, expected) {
		t.Errorf("Expected active link for last page, got %s", result)
	}
}
