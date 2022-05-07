package gnotter

import "strings"

func Render(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		// /r/boards:BOARD_NAME
		name := parts[0]
		boardI, exists := gProfilesByName.Get(name)
		if !exists {
			return "profile does not exist: " + name
		}
		return boardI.(*Profile).RenderProfile()
	}
	return "tmp"
}
