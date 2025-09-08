package blog

import (
	"testing"

	"gno.land/p/nt/uassert"
)

func TestParseRenderOptions(t *testing.T) {
	t.Run("default options with empty path", func(t *testing.T) {
		opts := ParseRenderOptions("/blog")
		uassert.Equal(t, "", opts.Sort)
		uassert.Equal(t, "", opts.TimeFormat)
		uassert.True(t, opts.Ascending)
		uassert.True(t, opts.IsGrid)
		uassert.False(t, opts.IsAlphabetical)
		uassert.False(t, opts.IsLastUpdated)
		uassert.False(t, opts.IsCommon)
		uassert.False(t, opts.HasOrder)
	})

	t.Run("full options parsing", func(t *testing.T) {
		url := "/blog?sort=alpha&order=desc&mode=list&time=full"
		opts := ParseRenderOptions(url)
		uassert.Equal(t, "alpha", opts.Sort)
		uassert.Equal(t, "full", opts.TimeFormat)
		uassert.False(t, opts.Ascending)
		uassert.False(t, opts.IsGrid)
		uassert.True(t, opts.IsAlphabetical)
		uassert.False(t, opts.IsLastUpdated)
		uassert.False(t, opts.IsCommon)
		uassert.True(t, opts.HasOrder)
	})
}

func TestToggle(t *testing.T) {
	uassert.Equal(t, "asc", toggle("desc"))
	uassert.Equal(t, "desc", toggle("asc"))
	uassert.Equal(t, "desc", toggle("")) // fallback
}

func TestToggleTimeFormat(t *testing.T) {
	uassert.Equal(t, "short", toggleTimeFormat("relative"))
	uassert.Equal(t, "full", toggleTimeFormat("short"))
	uassert.Equal(t, "relative", toggleTimeFormat("full"))
	uassert.Equal(t, "short", toggleTimeFormat("")) // fallback
}

func TestGetSortHelpers(t *testing.T) {
	base := "/blog?sort=recent&order=desc"

	uassert.Equal(t, "/blog?mode=grid", GetGridFmtPath("/blog"))
	uassert.Equal(t, "/blog?mode=list", GetListFmtPath("/blog"))

	uassert.Equal(t, "/blog?order=asc&sort=recent", GetSortByPublishedPath(base, "desc"))
	uassert.Equal(t, "/blog?order=asc&sort=update", GetSortByUpdatedPath(base, "desc"))
	uassert.Equal(t, "/blog?order=asc&sort=alpha", GetSortByAlphabeticalPath(base, "desc"))
	uassert.Equal(t, "/blog?order=asc&sort=common", GetSortByCommonPath(base, "desc"))
}

func TestTimeFormatAndRangeHelpers(t *testing.T) {
	base := "/blog"

	uassert.Equal(t, "/blog?time=full", GetTimeFmtPath(base, "full"))
	uassert.Equal(t, "/blog?start=2024-01-01", GetTimeStartPath(base, "2024-01-01"))
	uassert.Equal(t, "/blog?end=2025-01-01", GetTimeEndPath(base, "2025-01-01"))
}

func TestParseTimeRange(t *testing.T) {
	path := "/blog?start=2024-01-01&end=2024-12-31"
	start, end := parseTimeRange(path)
	uassert.NotEqual(t, nil, start)
	uassert.NotEqual(t, nil, end)

	uassert.Equal(t, true, start.Before(*end))
}
