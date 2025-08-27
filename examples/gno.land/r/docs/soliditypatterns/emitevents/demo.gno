package emitevents

import (
	"std"
	"strconv"
)

var Count int64 = 0

func Increment(cur realm) {
	Count++
	std.Emit("Incremented", "count", strconv.FormatInt(Count, 10))
}

func RenderDemo() string {
	content := "## Live Demo\n"
	content += "Counter Value : **" + strconv.FormatInt(Count, 10) + "**\n\n"
	content += "[Increment](emitevents$help&func=Increment)\n\n"
	return content
}
