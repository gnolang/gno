package render

import (
	"strings"
	"gno.land/p/moul/md"
)

const pageSize = 10

func Render(path string) string {
	if path == "" || strings.HasPrefix(path, "?page=") {
		return renderHome(path)
	}

	switch {
	case strings.HasPrefix(path, "?user="):
		return renderUserPage(path)
	case strings.HasPrefix(path, "?market="):
		return renderMarketPage(path)
	default:
	}

	return md.Paragraph("Unknown page.")
}
