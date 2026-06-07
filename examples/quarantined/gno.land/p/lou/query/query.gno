package query

import (
	"net/url"
)

// GetQueryValues returns all values for a given key in query parameters.
func GetQueryValues(params url.Values, key string) []string {
	return params[key]
}

// GetQueryFirstValue returns the first value for a given key in query parameters.
func GetQueryFirstValue(params url.Values, key string) string {
	if vals, ok := params[key]; ok && len(vals) > 0 {
		return vals[0]
	}
	return ""
}

// HasQueryKey checks if the given key exists in query parameters.
func HasQueryKey(params url.Values, key string) bool {
	_, exists := params[key]
	return exists
}

// GetQueryValueFromURL retrieves a single query value from the raw URL.
// If the URL is invalid or the parameter does not exist, it returns an empty string.
//
//	 value := GetQueryValueFromURL("user", "https://example.com/?user=john")
//		// value == "john"
func GetQueryValueFromURL(key, rawPath string) (string, error) {
	u, err := url.Parse(rawPath)
	if err != nil {
		return "", err
	}
	return u.Query().Get(key), nil
}

// GetQuery retrieves a single query parameter value from the raw URL.
// If the URL is invalid or the parameter does not exist, it returns an empty string.
func GetQuery(key, rawPath string) string {
	u, err := url.Parse(rawPath)
	if err != nil {
		return ""
	}
	return u.Query().Get(key)
}

// ParseQueryFirstValuesMap returns a map of the first values for all query parameters.
//
//	m := ParseQueryFirstValuesMap("https://example.com/?foo=1&bar=2")
//	// m == map[string]string{"foo": "1", "bar": "2"}
func ParseQueryFirstValuesMap(rawPath string) (map[string]string, error) {
	u, err := url.Parse(rawPath)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for key, vals := range u.Query() {
		if len(vals) > 0 {
			result[key] = vals[0]
		}
	}
	return result, nil
}

// UpdateQueryValue sets a single query key to one value, replacing existing values.
//
//	url := UpdateQueryValue("https://example.com/?foo=1", "foo", "2")
//	// url == "https://example.com/?foo=2"
func UpdateQueryValue(rawPath, key, value string) (string, error) {
	return UpdateQueryValues(rawPath, key, []string{value})
}

// UpdateQueryValues sets a single query key to multiple values.
// Existing values are replaced, and new ones are added. Returns the modified URL as a string.
//
//	url := UpdateQueryValues("https://example.com/?foo=1", "foo", []string{"2", "3"})
//	// url == "https://example.com/?foo=2&foo=3"
func UpdateQueryValues(rawPath, key string, values []string) (string, error) {
	u, err := url.Parse(rawPath)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q[key] = values
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// UpdateQueryFirstValues sets multiple query parameters to single values.
// Existing parameters are overwritten, and new ones are added. Returns the modified URL as a string.
//
//	url := UpdateQueryFirstValues("https://example.com/?foo=1", map[string]string{"foo": "2", "bar": "3"})
//	// url == "https://example.com/?bar=3&foo=2"
func UpdateQueryFirstValues(rawPath string, updates map[string]string) (string, error) {
	u, err := url.Parse(rawPath)
	if err != nil {
		return "", err
	}
	q := u.Query()
	for k, v := range updates {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// UpdateQueryAllValues sets multiple query parameters, each with multiple values.
// Existing parameters are replaced, and new ones are added. Returns the modified URL as a string.
//
//	url := UpdateQueryAllValues("https://example.com/?foo=1", map[string][]string{"foo": {"2", "3"}, "bar": {"4"}})
//	// url == "https://example.com/?bar=4&foo=2&foo=3"
func UpdateQueryAllValues(rawPath string, updates map[string][]string) (string, error) {
	u, err := url.Parse(rawPath)
	if err != nil {
		return "", err
	}
	q := u.Query()
	for k, v := range updates {
		q[k] = v
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// SetQueries sets multiple query parameters in the URL.
// It replaces existing parameters and adds new ones, returning the modified URL as a string.
//
// url := SetQueries("https://example.com/?fa=1", map[string]string{"foo": "2", "bar": "3"})
// // url == "https://example.com/?bar=3&foo=2"
func SetQueries(rawPath string, queries map[string]string) (string, error) {
	u, err := url.Parse(rawPath)
	if err != nil {
		return "", err
	}
	q := u.Query()
	for k, v := range queries {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// SetQueriesMulti sets multiple query parameters with multiple values.
// It replaces existing parameters and adds new ones, returning the modified URL as a string.
//
// url := SetQueriesMulti("https://example.com/?fa=1", map[string][]string{"foo": {"2", "3"}, "bar": {"4"}})
// // url == "https://example.com/?bar=4&foo=2&foo=3"
func SetQueriesMulti(rawPath string, queries map[string][]string) (string, error) {
	u, err := url.Parse(rawPath)
	if err != nil {
		return "", err
	}
	q := u.Query()
	for k, v := range queries {
		q[k] = v
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// DeleteQuery removes a key from the query parameters.
//
//	url := DeleteQuery("https://example.com/?foo=1&bar=2", "foo")
//	// url == "https://example.com/?bar=2"
func DeleteQuery(rawPath, key string) (string, error) {
	u, err := url.Parse(rawPath)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Del(key)
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// ResetQuery clears all query parameters from the URL path.
//
//	url := ResetQuery("https://example.com/?foo=1&bar=2")
//	// url == "https://example.com/"
func ResetQuery(rawPath string) (string, error) {
	u, err := url.Parse(rawPath)
	if err != nil {
		return "", err
	}
	u.RawQuery = ""
	return u.String(), nil
}
