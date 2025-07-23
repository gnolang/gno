package counter

import "strconv"

var count int64 = 0

func Increment(cur realm) {
	count++
}

func Decrement(cur realm) {
	count--
}

func RenderDemo() string {
	content := "# DEMO\n"
	content += "Counter Value : **" + strconv.FormatInt(count, 10) + "**\n\n"
	content += "[Increment](counter$help&func=Increment)\n\n"
	content += "[Decrement](counter$help&func=Decrement)\n\n"
	return content
}
