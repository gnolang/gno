package mux

import "strings"

type UrlQuery map[string][]string

/*
Query() returns a full map of key-listValue from request
to get a list of params by key, use: req.Query()["key"]
to get single param by key, use: req.Query().Get("key")
*/
func (r *Request) Query() UrlQuery {
	urlQueries := UrlQuery{}
	reqParts := strings.Split(r.Path, "/")
	for i := 0; i < len(reqParts); i++ {
		rFullPath := reqParts[i]
		switch {
		// find the starting of query string
		case strings.Contains(rFullPath, "?"):
			{
				index := strings.Index(rFullPath, "?")
				// get the query path and the query string
				// the rQueryPath may be used for handler for handling request path
				// rQueryPath := rFullPath[:index]
				queryString := rFullPath[index+1:]
				params := strings.Split(queryString, "&")
				for _, param := range params {
					// find key - value of query params, then append to the UrlQuery
					kv := strings.Split(param, "=")
					if len(kv) == 2 {
						urlQueries[kv[0]] = append(urlQueries[kv[0]], kv[1])
					}
				}
			}
		default:
			// continue
		}
	}
	return urlQueries
}

// get the query by key
// if there is more than one query, returns the very first param
func (qs UrlQuery) Get(key string) string {
	listMatching, ok := qs[key]
	if ok {
		return listMatching[0]
	} else {
		return ""
	}
}
