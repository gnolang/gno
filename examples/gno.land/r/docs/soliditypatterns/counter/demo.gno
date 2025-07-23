package counter

import "strconv"

var Count int64 = 0

func Increment(cur realm) {
	Count++
}

func Decrement(cur realm) {
	Count--
}

func GetCount(cur realm) int64 {
	return Count
}

func RenderDemo() string {
	content := "# DEMO\n"
	content += "Counter Value : **" + strconv.FormatInt(Count, 10) + "**\n\n"
	content += "[Increment](counter$help&func=Increment)\n\n"
	content += "[Decrement](counter$help&func=Decrement)\n\n"
	return content
}
