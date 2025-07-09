package query

import (
	"net/url"
)

// ParseQueryParams retrieves the value of a specific query parameter from the given raw URL path.
// If the URL is invalid or the parameter does not exist, it returns an empty string.
//
//	Example:
//	  value := ParseQueryParams("user", "https://example.com/?user=john")
//	  // value == "john"
func ParseQueryParams(query, rawPath string) (string, error) {
	u, err := url.Parse(rawPath)
	if err != nil {
		return "", err
	}
	return u.Query().Get(query), nil
}

// ParseQueryMap parses the query string from the given raw URL path and returns
// a map containing the first value for each query parameter.
//
//	Example:
//	  m := ParseQueryMap("https://example.com/?foo=1&bar=2")
//	  // m == map[string]string{"foo": "1", "bar": "2"}
func ParseQueryMap(rawPath string) (map[string]string, error) {
	u, err := url.Parse(rawPath)
	if err != nil {
		return map[string]string{}, err
	}
	params := make(map[string]string)
	for key, vals := range u.Query() {
		if len(vals) > 0 {
			params[key] = vals[0]
		}
	}
	return params, nil
}

// UpdateQueryMulti updates multiple query parameters in the given raw URL path.
// Existing parameters are overwritten, and new ones are added. Returns the modified URL as a string.
//
//	Example:
//	  url := UpdateQueryMulti("https://example.com/?foo=1", map[string]string{"foo": "2", "bar": "3"})
//	  // url == "https://example.com/?bar=3&foo=2"
func UpdateQueryMulti(rawPath string, updates map[string]string) (string, error) {
	params, err := ParseQueryMap(rawPath)
	if err != nil {
		return "", err
	}
	for k, v := range updates {
		params[k] = v
	}
	return buildQueryPath(rawPath, params)
}

// UpdateQuery updates a single query parameter in the given raw URL path.
// It adds the parameter if it doesn't exist and overwrites it if it does.
//
//	Example:
//	  url := UpdateQuery("https://example.com/?foo=1", "foo", "2")
//	  // url == "https://example.com/?foo=2"
func UpdateQuery(rawPath, key, value string) (string, error) {
	u, err := url.Parse(rawPath)
	if err != nil {
		return "", err
	}
	query := u.Query()
	query.Set(key, value)
	u.RawQuery = query.Encode()
	return u.String(), nil
}

// DeleteQuery removes a query parameter from the URL.
//
//	Example:
//	  url := DeleteQuery("https://example.com/?foo=1&bar=2", "foo")
//	  // url == "https://example.com/?bar=2"
func DeleteQuery(rawURL, key string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Del(key)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// buildQueryPath rebuilds a URL with updated query parameters.
func buildQueryPath(rawPath string, params map[string]string) (string, error) {
	u, err := url.Parse(rawPath)
	if err != nil {
		return "", err
	}
	q := url.Values{}
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}
