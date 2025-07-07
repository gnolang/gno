package query

import (
	"net/url"
)

func ParseQueryParams(query, rawPath string) string {
	u, err := url.Parse(rawPath)
	if err != nil {
		return ""
	}
	return u.Query().Get(query)
}

func ParseQueryMap(rawPath string) map[string]string {
	u, err := url.Parse(rawPath)
	if err != nil {
		return map[string]string{}
	}
	params := make(map[string]string)
	for key, vals := range u.Query() {
		if len(vals) > 0 {
			params[key] = vals[0]
		}
	}
	return params
}

func buildQueryPath(rawPath string, params map[string]string) string {
	u, err := url.Parse(rawPath)
	if err != nil {
		return rawPath
	}
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func UpdateQueryMulti(rawPath string, updates map[string]string) string {
	params := parseQueryMap(rawPath)
	for k, v := range updates {
		params[k] = v
	}
	return buildQueryPath(rawPath, params)
}

func UpdateQuery(rawPath, key, value string) string {
	u, err := url.Parse(rawPath)
	if err != nil {
		return rawPath
	}
	query := u.Query()
	query.Set(key, value)
	u.RawQuery = query.Encode()
	return u.String()
}
