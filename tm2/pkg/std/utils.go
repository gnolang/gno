package std

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// sortJSON takes any JSON and returns it sorted by keys. Also, all white-spaces
// are removed.
// This method can be used to canonicalize JSON to be returned by GetSignBytes,
// e.g. for the ledger integration.
// If the passed JSON isn't valid it will return an error.
// A repeated key collapses to its last value; unmarshalStrict refuses one.
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

// unmarshalStrict is json.Unmarshal that also refuses a repeated key in any
// object: Go keeps its last value, other parsers its first.
func unmarshalStrict(data []byte, v any) error {
	// Unmarshal first: it bounds the nesting depth rejectDuplicateKeys recurses
	// into, and refuses data after the value.
	if err := json.Unmarshal(data, v); err != nil {
		return err
	}
	return rejectDuplicateKeys(json.NewDecoder(bytes.NewReader(data)))
}

// rejectDuplicateKeys reads one JSON value from dec and fails if an object in it
// repeats a key.
func rejectDuplicateKeys(dec *json.Decoder) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}

	delim, ok := tok.(json.Delim)
	if !ok {
		return nil
	}

	var seen map[string]bool
	if delim == '{' {
		seen = map[string]bool{}
	}
	for dec.More() {
		if seen != nil {
			tok, err := dec.Token()
			if err != nil {
				return err
			}
			key := tok.(string)
			if seen[key] {
				return fmt.Errorf("duplicate key %q", key)
			}
			seen[key] = true
		}
		if err := rejectDuplicateKeys(dec); err != nil {
			return err
		}
	}

	_, err = dec.Token() // the closing delimiter
	return err
}
