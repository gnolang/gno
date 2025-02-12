package common

import (
	"strconv"

	"gno.land/p/demo/ufmt"

	"gno.land/p/demo/grc/grc721"
)

// PositionIdFrom converts positionId to grc721.TokenID type
// NOTE: input parameter positionId can be string, int, uint64, or grc721.TokenID
// if positionId is nil or not supported, it will panic
// if input type is not supported, it will panic
// input: positionId interface{}
// output: grc721.TokenID
func PositionIdFrom(positionId interface{}) grc721.TokenID {
	if positionId == nil {
		panic(newErrorWithDetail(
			errInvalidPositionId,
			"can not be nil",
		))
	}

	switch positionId.(type) {
	case string:
		return grc721.TokenID(positionId.(string))
	case int:
		return grc721.TokenID(strconv.Itoa(positionId.(int)))
	case uint64:
		return grc721.TokenID(strconv.Itoa(int(positionId.(uint64))))
	case grc721.TokenID:
		return positionId.(grc721.TokenID)
	default:
		estimatedType := ufmt.Sprintf("%T", positionId)
		panic(newErrorWithDetail(
			errInvalidPositionId,
			ufmt.Sprintf("unsupported positionId type: %s", estimatedType),
		))
	}
}
