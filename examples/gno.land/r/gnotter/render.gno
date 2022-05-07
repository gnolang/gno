package gnotter

import "strings"

func Render(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		// /r/profile:PROFILE_NAME
		name := parts[0]
		profileI, exists := gProfilesByName.Get(name)
		if !exists {
			return "profile does not exist: " + name
		}
		return profileI.(*Profile).RenderProfile()
	}
	return "unreacheable"
}
