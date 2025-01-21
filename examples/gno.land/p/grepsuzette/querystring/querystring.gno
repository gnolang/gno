package querystring

import "strings"

// Break down a string of url parameters to map[string]string.
// Since #2876, uri such as /r/test:a=1&b=2?c=3&d=4 are possible.
// In case of a redefined value, the querystring has priority over the path.
// E.g.
// "a=1&b=2?c=3&b=4" -> (map[string]string{a:"1", b:"4", c:"3"}, nil)
func ToMap(querystring string) (map[string]string, error) {
	m := make(map[string]string)
	t := strings.Replace(querystring, "?", "&", 1)
	for _, s := range strings.Split(t, "&") {
		if len(strings.TrimSpace(s)) == 0 {
			continue
		}
		idx := strings.Index(s, "=")
		if idx > -1 {
			m[s[:idx]] = s[idx+1:]
		} else {
			m[s] = ""
		}
	}
	return m, nil
}
