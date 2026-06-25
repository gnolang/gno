package counter

import "strconv"

var count int64 = 0

func Increment(cur realm) {
	if !cur.IsCurrent() {
		panic("spoofed realm")
	}
	count++
}

func Decrement(cur realm) {
	if !cur.IsCurrent() {
		panic("spoofed realm")
	}
	count--
}

func GetCount() int64 {
	return count
}

func RenderDemo() string {
	content := "## Live Demo\n"
	content += "Counter Value : **" + strconv.FormatInt(count, 10) + "**\n\n"
	content += "[Increment](counter$help&func=Increment)\n\n"
	content += "[Decrement](counter$help&func=Decrement)\n\n"
	return content
}
