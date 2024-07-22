package mux

import "strings"

type UrlQuery map[string][]string

/*
Query() returns a full map of key-listValue from request
to get a list of params by key, use: req.Query()["key"]
to get single param by key, use: req.Query().Get("key")
to get request path use GetPath()
*/

func (r *Request) QueryFull() (UrlQuery, string) {
	urlQueries := UrlQuery{}
	pathQuery := ""
	// checking if there is query in r.Path
	switch {
	case strings.Contains(r.Path, "?"):
		// get baseUrl and queryString
		rParts := strings.SplitN(r.Path, "?", 2)
		if len(rParts) == 1 {
			// if this is not a `query string` query, then returns full path as queryPath
			return urlQueries, rParts[0]
		}
		baseUrl := rParts[0]
		queryString := rParts[1]

		// find the target path
		baseUrlParts := strings.Split(baseUrl, "/")
		rTargetPath := baseUrlParts[len(baseUrlParts)-1]
		pathQuery = rTargetPath

		// process the queryString
		switch {
		// find the first param index
		case strings.Contains(queryString, "="):
			for _, keyValue := range strings.Split(queryString, "&") {
				parts := strings.SplitN(keyValue, "=", 2)
				if len(parts) != 2 {
					return urlQueries, rParts[0]
				}
				key, value := parts[0], parts[1]
				urlQueries[key] = append(urlQueries[key], value)
			}
		default:
			// continue
		}

	default:
		// continue
	}
	return urlQueries, pathQuery
}

// get the query by key
// if there is more than one query, returns the very first param
func (qs UrlQuery) Get(key string) string {
	listMatching, ok := qs[key]
	if ok {
		return listMatching[0]
	}
	return ""
}

func (r *Request) Query() UrlQuery {
	uQuery, _ := r.QueryFull()
	return uQuery
}

func (r *Request) GetQueryPath() string {
	_, pQuery := r.QueryFull()
	return pQuery
}
