package renderer

func Bold(str string) string {
	return "**" + str + "**"
}

func Italic(str string) string {
	return "*" + str + "*"
}

func CodeBlock(language string, code string) string {
	return "```" + language + "\n" + code + "\n```\n"
}

func Header(id int, str string) string {
	if id <= 0 {
		return " " + str + "\n"
	}
	return "#" + Header(id-1, str)
}

func Picture(alt string, path string) string {
	return "![" + alt + "](" + path + ")"
}

func Link(alt string, path string) string {
	return "[" + alt + "](" + path + ")"
}

func Table(tab [][]string) (rend string) {
	if len(tab) == 0 {
		return "empty table"
	}
	size := len(tab[0])
	if size == 0 {
		return "empty first line in table"
	}
	for idx, line := range tab {
		if len(tab[idx]) != size {
			return "line size missmatch in table"
		}
		rend += "| "
		for _, elem := range line {
			rend += elem
			rend += " | "
		}
		rend += "\n"
		if idx == 0 {
			rend += "| "
			for i := 0; i < size; i++ {
				rend += ":-: | "
			}
			rend += "\n"
		}
	}
	rend += "\n"
	return
}
