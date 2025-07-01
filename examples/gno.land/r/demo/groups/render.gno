package groups

import (
	"strings"
)

//----------------------------------------
// Render functions

func RenderGroup(gid GroupID) string {
	group := getGroup(gid)
	if group == nil {
		return "missing Group"
	}
	return group.RenderGroup()
}

func Render(path string) string {
	if path == "" {
		str := "List of all Groups:\n\n"
		gGroups.Iterate("", "", func(key string, value any) bool {
			group := value.(*Group)
			str += " * [" + group.name + "](" + group.url + ")\n"
			return false
		})
		return str
	}
	parts := strings.Split(path, "/")
	if len(parts) == 1 {
		// /r/demo/groups:Group_NAME
		name := parts[0]
		groupI, exists := gGroupsByName.Get(name)
		if !exists {
			return "Group does not exist: " + name
		}
		return groupI.(*Group).RenderGroup()
	} else {
		return "unrecognized path " + path
	}
}
