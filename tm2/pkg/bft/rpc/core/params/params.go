package params

import (
	"encoding/hex"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
)

func get(params []any, idx int) any {
	if idx < 0 || idx >= len(params) {
		return nil
	}

	return params[idx]
}

// AsString returns the parameter at idx as a string, accepting raw strings or decoding JSON-RPC values via Amino
func AsString(params []any, idx int) (string, *spec.BaseJSONError) {
	raw := get(params, idx)
	if raw == nil {
		return "", nil
	}

	switch v := raw.(type) {
	case string:
		// Query params are strings already
		return v, nil
	default:
		// For JSON-RPC POSTs, go through Amino to preserve legacy behavior
		b, err := json.Marshal(v)
		if err != nil {
			return "", spec.GenerateInvalidParamError(idx)
		}

		var out string
		if err = amino.UnmarshalJSON(b, &out); err != nil {
			return "", spec.GenerateInvalidParamError(idx)
		}

		return out, nil
	}
}

// AsBytes returns the parameter at idx as bytes, supporting 0x-prefixed hex and Amino JSON
// semantics otherwise, optionally requiring the value to be present
func AsBytes(params []any, idx int, required bool) ([]byte, *spec.BaseJSONError) {
	raw := get(params, idx)
	if raw == nil {
		if required {
			return nil, spec.GenerateInvalidParamError(idx)
		}

		return nil, nil
	}

	switch v := raw.(type) {
	case string:
		// HTTP GET compatibility, 0x-prefixed hex
		if strings.HasPrefix(v, "0x") {
			data, err := hex.DecodeString(v[2:])
			if err != nil {
				return nil, spec.GenerateInvalidParamError(idx)
			}

			return data, nil
		}

		// For everything else, Amino semantics for []byte
		b, err := amino.MarshalJSON(v)
		if err != nil {
			return nil, spec.GenerateInvalidParamError(idx)
		}

		var out []byte
		if err := amino.UnmarshalJSON(b, &out); err != nil {
			return nil, spec.GenerateInvalidParamError(idx)
		}

		return out, nil

	default:
		// For JSON-RPC POSTs, the value is already decoded by encoding/json
		b, err := json.Marshal(v)
		if err != nil {
			return nil, spec.GenerateInvalidParamError(idx)
		}

		var out []byte
		if err := amino.UnmarshalJSON(b, &out); err != nil {
			return nil, spec.GenerateInvalidParamError(idx)
		}

		return out, nil
	}
}

// AsInt64 returns the parameter at idx as an int64, accepting native numeric types and query-string integers,
// with a JSON -> Amino fallback for legacy behavior
func AsInt64(params []any, idx int) (int64, *spec.BaseJSONError) {
	raw := get(params, idx)
	if raw == nil {
		return 0, nil
	}

	switch v := raw.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case float64:
		// JSON numbers -> int64 (old Amino expected strings, but no client should rely on that distinction)
		return int64(v), nil
	case string:
		// HTTP GET: query param is always a string.
		// Old Amino wrapped integer-looking strings in quotes and then used Amino decoding
		if v == "" {
			return 0, nil
		}

		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0, spec.GenerateInvalidParamError(idx)
		}

		return i, nil
	default:
		// Fallback, json -> amino -> int64
		b, err := json.Marshal(v)
		if err != nil {
			return 0, spec.GenerateInvalidParamError(idx)
		}

		var out int64
		if err := amino.UnmarshalJSON(b, &out); err != nil {
			return 0, spec.GenerateInvalidParamError(idx)
		}

		return out, nil
	}
}

// AsBool returns the parameter at idx as a bool, accepting native bools and
// "true"/"false" strings, with a JSON -> Amino fallback for legacy behavior
func AsBool(params []any, idx int) (bool, *spec.BaseJSONError) {
	raw := get(params, idx)
	if raw == nil {
		return false, nil
	}

	switch v := raw.(type) {
	case bool:
		return v, nil

	case string:
		// Accept "true"/"false" as HTTP query values
		switch strings.ToLower(v) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		default:
			return false, spec.GenerateInvalidParamError(idx)
		}
	default:
		// Fallback, json -> amino -> bool
		b, err := json.Marshal(v)
		if err != nil {
			return false, spec.GenerateInvalidParamError(idx)
		}

		var out bool
		if err := amino.UnmarshalJSON(b, &out); err != nil {
			return false, spec.GenerateInvalidParamError(idx)
		}

		return out, nil
	}
}
