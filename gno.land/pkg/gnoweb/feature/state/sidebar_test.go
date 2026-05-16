package state

import (
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShortenOID(t *testing.T) {
	t.Parallel()

	const realmA = "ff61a23bc5d8c018b6c8f29498b1b89435bbeb998:11"
	const realmB = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa:7"

	tests := []struct {
		name     string
		id, ref  string
		expected string
	}{
		{"same hashlet keeps :N suffix", "ff61a23bc5d8c018b6c8f29498b1b89435bbeb998:1", realmA, ":1"},
		{"different hashlet returns full id", realmB, realmA, realmB},
		{"id without colon untouched", "abcdef", realmA, "abcdef"},
		{"ref without colon returns full id", realmA, "noref", realmA},
		{"empty id returns empty", "", realmA, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, ShortenOID(tt.id, tt.ref))
		})
	}
}

func TestTruncMid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		head, tail int
		expected   string
	}{
		{"long string truncated head…tail", "ff61a23bc5d8c018b6c8f29498b1b89435bbeb998", 6, 4, "ff61a2…b998"},
		{"already short string untouched", "abc", 6, 4, "abc"},
		{"exactly threshold untouched", "ff61a23bc", 4, 4, "ff61a23bc"},
		{"tail zero gives head…", "abcdefghij", 3, 0, "abc…"},
		{"head zero gives …tail", "abcdefghij", 0, 3, "…hij"},
		{"empty stays empty", "", 6, 4, ""},
		{"negative bounds clamp to zero", "abcdefghij", -1, -1, "…"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, truncMid(tt.input, tt.head, tt.tail))
		})
	}
}

func TestTruncOID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		input      string
		head, tail int
		expected   string
	}{
		{
			"OID hashlet truncated, :N preserved",
			"ff61a23bc5d8c018b6c8f29498b1b89435bbeb998:11", 6, 4,
			"ff61a2…b998:11",
		},
		{
			"short OID still shows :N",
			"abc:1", 6, 4,
			"abc:1",
		},
		{
			"hash without colon truncated bare",
			"ff61a23bc5d8c018b6c8f29498b1b89435bbeb998", 6, 4,
			"ff61a2…b998",
		},
		{
			"empty returns empty",
			"", 6, 4,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, TruncOID(tt.input, tt.head, tt.tail))
		})
	}
}

// TestBuildPackageSidebarFull pins the full-sidebar contract: every
// realm decl gets an entry (icon + label + type), on-page rows resolve
// to in-page anchors, off-page rows resolve to cross-page URLs with
// the correct paginated offset, and the cap surfaces a "+N more" hint.
func TestBuildPackageSidebarFull(t *testing.T) {
	t.Parallel()

	names := []string{"A", "B", "C", "D", "E", "F", "G"} // 7 entries
	anchors := []string{
		"state-a", "state-b", "state-c", "state-d", "state-e", "state-f", "state-g",
	}
	kinds := []string{
		KindStruct, KindStruct, KindStruct, KindStruct, KindStruct, KindStruct, KindStruct,
	}
	types := []string{"T1", "T2", "T3", "T4", "T5", "T6", "T7"}

	// Page 2 with limit 2 → on-page indices = [2, 3].
	sidebar, truncated := BuildPackageSidebarFull(
		"/r/foo", names, anchors, kinds, types, 2, 2, "",
	)
	require.NotNil(t, sidebar)
	assert.False(t, truncated, "7 entries ≤ cap → not truncated")
	require.Len(t, sidebar.TOC, 7, "every decl emits an entry")

	// Entry 0 — off-page (index 0, currentOffset=2): page 1 → no offset=.
	e0 := sidebar.TOC[0]
	assert.False(t, e0.OnPage)
	assert.Equal(t, KindStruct, e0.Kind, "kind populated for off-page entries (icon-everywhere)")
	assert.Equal(t, "T1", e0.Type)
	assert.NotContains(t, string(e0.PrettyHref), "offset=",
		"index 0 lives on page 1 → canonical URL omits offset=")
	assert.Contains(t, string(e0.PrettyHref), "#state-a-pretty")
	assert.Contains(t, string(e0.TreeHref), "view=tree")
	assert.Contains(t, string(e0.TreeHref), "#state-a-tree")

	// Entry 2 — on-page: in-page anchor, OnPage=true.
	e2 := sidebar.TOC[2]
	assert.True(t, e2.OnPage)
	assert.Equal(t, template.URL("#state-c-pretty"), e2.PrettyHref)
	assert.Equal(t, template.URL("#state-c-tree"), e2.TreeHref)
	assert.Equal(t, KindStruct, e2.Kind)
	assert.Equal(t, "T3", e2.Type)

	// Entry 4 — off-page on page 3 (offset=4 with limit=2).
	e4 := sidebar.TOC[4]
	assert.False(t, e4.OnPage)
	assert.Contains(t, string(e4.PrettyHref), "offset=4")
	assert.Contains(t, string(e4.PrettyHref), "#state-e-pretty")
}
