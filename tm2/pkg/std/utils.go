package std

import "encoding/json"

// sortJSON takes any JSON and returns it sorted by keys. Also, all white-spaces
// are removed.
// This method can be used to canonicalize JSON to be returned by GetSignBytes,
// e.g. for the ledger integration.
// If the passed JSON isn't valid it will return an error
func sortJSON(toSortJSON []byte) ([]byte, error) {
	var c any

	if err := json.Unmarshal(toSortJSON, &c); err != nil {
		return nil, err
	}

	js, err := json.Marshal(c)
	if err != nil {
		return nil, err
	}

	return js, nil
}

// MustSortJSON is like sortJSON but panic if an error occurs, e.g., if
// the passed JSON isn't valid.
func MustSortJSON(toSortJSON []byte) []byte {
	js, err := sortJSON(toSortJSON)
	if err != nil {
		panic(err)
	}
	return js
}
