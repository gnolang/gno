package params

import (
	"fmt"
	"strings"
	"unicode"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

type paramsHandler struct {
	params ParamsKeeper
}

func NewHandler(params ParamsKeeper) paramsHandler {
	return paramsHandler{
		params: params,
	}
}

func (bh paramsHandler) Process(ctx sdk.Context, msg std.Msg) sdk.Result {
	errMsg := fmt.Sprintf("unrecognized params message type: %T", msg)
	return abciResult(std.ErrUnknownRequest(errMsg))
}

// ----------------------------------------
// Query:
// - params/prefix:key for a prefixed module parameter key.
// - params/key for an arbitrary parameter key.
func (bh paramsHandler) Query(ctx sdk.Context, req abci.RequestQuery) (res abci.ResponseQuery) {
	prefix, paramKey := parseParamKey(req.Path)
	if prefix != "" {
		if bh.params.PrefixExists(prefix) == false {
			res = sdk.ABCIResponseQueryFromError(
				std.ErrUnknownRequest(fmt.Sprintf("unknown params query endpoint %q", prefix)))
			return
		}
	}
	if paramKey == "" {
		res = sdk.ABCIResponseQueryFromError(
			std.ErrUnknownRequest("param key is empty"))
		return
	}
	val := bh.params.GetRaw(ctx, paramKey)
	res.Data = val
	return
}

//----------------------------------------
// misc

func abciResult(err error) sdk.Result {
	return sdk.ABCIResultFromError(err)
}

/*
// return the parameter key and it's prefix
func parseParamKeyWithPrefix(path string) (prefix, key string) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return "", ""
	}

	subParts := strings.SplitN(parts[1], ":", 2)
	if len(subParts) < 2 {
		return "", parts[1]
	}
	return subParts[0], parts[1]

}
*/
// paramKey may include a prefix in the format "<prefix:>" if a prefix is detected.
func parseParamKey(path string) (prefix, key string) {
	parts := strings.SplitN(path, "/", 2)
	if len(parts) < 2 {
		return "", ""
	}
	remainder := parts[1]
	// Look for the first colon.
	colonIndex := strings.Index(remainder, ":")
	if colonIndex == -1 {
		// No colon found: treat entire remainder as the key.
		return "", remainder
	}

	candidatePrefix := remainder[:colonIndex]
	// If candidatePrefix , it is not considered a valid module prefix.

	isPrefix := true
	for _, char := range candidatePrefix {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) && char != '_' {
			isPrefix = false
		}
	}

	if isPrefix == false {
		return "", remainder
	}

	// Otherwise, candidatePrefix is valid.
	return candidatePrefix, remainder
}
