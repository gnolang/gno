package querystring

import "strings"

// Break down path and querystring into a map[string]string.
// Since #2876, realm paths include the querystring,
// as in `/r/test:a?c=3` (question mark introduces the querystring).
//
// In case of a name defined several times, the last value wins.
//
// "a=1"             -> {a:"1"}, nil
// "?a=1"            -> {a:"1"}, nil
// "a=1&b=2?c=3&b=4" -> {a:"1", b:"4", c:"3"}, nil
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
