package common

import (
	"strconv"

	"gno.land/p/demo/ufmt"

	"gno.land/p/demo/grc/grc721"
)

// TokenIdFrom converts tokenId to grc721.TokenID type
// NOTE: input parameter tokenId can be string, int, uint64, or grc721.TokenID
// if tokenId is nil or not supported, it will panic
// if input type is not supported, it will panic
// input: tokenId interface{}
// output: grc721.TokenID
func TokenIdFrom(tokenId interface{}) grc721.TokenID {
	if tokenId == nil {
		panic(newErrorWithDetail(
			errInvalidTokenId,
			"can not be nil",
		))
	}

	switch tokenId.(type) {
	case string:
		return grc721.TokenID(tokenId.(string))
	case int:
		return grc721.TokenID(strconv.Itoa(tokenId.(int)))
	case uint64:
		return grc721.TokenID(strconv.Itoa(int(tokenId.(uint64))))
	case grc721.TokenID:
		return tokenId.(grc721.TokenID)
	default:
		estimatedType := ufmt.Sprintf("%T", tokenId)
		panic(newErrorWithDetail(
			errInvalidTokenId,
			ufmt.Sprintf("unsupported tokenId type: %s", estimatedType),
		))
	}
}
