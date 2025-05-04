package queryparsing

import "strings"

func splitAtCharacter(fullpath string, s string) (string, string) {
	i := strings.Index(fullpath, s)
	if i == -1 {
		return fullpath, ""
	}
	return fullpath[:i], fullpath[i+1:]
}

func Parse(fullpath string) (path string, params map[string]string) {
	var oneParam string

	params = map[string]string{}
	path, oneParam = splitAtCharacter(fullpath, "?")
	paramList := strings.Split(oneParam, "&")
	for _, v := range paramList {
		x, y := splitAtCharacter(v, "=")
		params[x] = y
	}
	return
}