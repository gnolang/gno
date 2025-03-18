package position

import (
	"std"

	"gno.land/p/demo/json"

	"gno.land/p/gnoswap/consts"
	"gno.land/r/gnoswap/v1/gnft"
)

type JsonResponse interface {
	JSON() *json.Node
}

// helper function for creating JSON response
func makeJsonResponse(stat *ResponseQueryBase, response JsonResponse) string {
	node := json.ObjectNode("", map[string]*json.Node{
		"stat":     stat.JSON(),
		"response": response.JSON(),
	})

	b, err := json.Marshal(node)
	if err != nil {
		panic(err.Error())
	}

	return string(b)
}

// Type for Positions response
type PositionsResponse struct {
	Positions []RpcPosition
}

func (pr *PositionsResponse) JSON() *json.Node {
	rsps := json.ArrayNode("", []*json.Node{})
	for _, position := range pr.Positions {
		owner, err := gnft.OwnerOf(positionIdFrom(position.LpPositionId))
		if err != nil {
			owner = consts.ZERO_ADDRESS
		}
		rsps.AppendArray(position.JSON(owner))
	}
	return rsps
}

// Type for UnclaimedFee response
type UnclaimedFeesResponse struct {
	Fees []RpcUnclaimedFee
}

func (ur *UnclaimedFeesResponse) JSON() *json.Node {
	rsps := json.ArrayNode("", []*json.Node{})
	for _, fee := range ur.Fees {
		rsps.AppendArray(fee.JSON())
	}
	return rsps
}

///////////////// RPC TYPES /////////////////

type RpcPosition struct {
	LpPositionId             uint64 `json:"lpPositionId"`
	Burned                   bool   `json:"burned"`
	Owner                    string `json:"owner"`
	Operator                 string `json:"operator"`
	PoolKey                  string `json:"poolKey"`
	TickLower                int32  `json:"tickLower"`
	TickUpper                int32  `json:"tickUpper"`
	Liquidity                string `json:"liquidity"`
	FeeGrowthInside0LastX128 string `json:"feeGrowthInside0LastX128"`
	FeeGrowthInside1LastX128 string `json:"feeGrowthInside1LastX128"`
	TokensOwed0              string `json:"token0Owed"`
	TokensOwed1              string `json:"token1Owed"`

	Token0Balance           string `json:"token0Balance"`
	Token1Balance           string `json:"token1Balance"`
	CalculatedToken0Balance string `json:"calculatedToken0Balance"`
	CalculatedToken1Balance string `json:"calculatedToken1Balance"`
	FeeUnclaimed0           string `json:"fee0Unclaimed"`
	FeeUnclaimed1           string `json:"fee1Unclaimed"`
}

func (p RpcPosition) JSON(owner std.Address) *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"lpPositionId":             json.NumberNode("lpPositionId", float64(p.LpPositionId)),
		"burned":                   json.BoolNode("burned", p.Burned),
		"owner":                    json.StringNode("owner", owner.String()),
		"operator":                 json.StringNode("operator", p.Operator),
		"poolKey":                  json.StringNode("poolKey", p.PoolKey),
		"tickLower":                json.NumberNode("tickLower", float64(p.TickLower)),
		"tickUpper":                json.NumberNode("tickUpper", float64(p.TickUpper)),
		"liquidity":                json.StringNode("liquidity", p.Liquidity),
		"feeGrowthInside0LastX128": json.StringNode("feeGrowthInside0LastX128", p.FeeGrowthInside0LastX128),
		"feeGrowthInside1LastX128": json.StringNode("feeGrowthInside1LastX128", p.FeeGrowthInside1LastX128),
		"token0Owed":               json.StringNode("token0Owed", p.TokensOwed0),
		"token1Owed":               json.StringNode("token1Owed", p.TokensOwed1),
		"token0Balance":            json.StringNode("token0Balance", p.Token0Balance),
		"token1Balance":            json.StringNode("token1Balance", p.Token1Balance),
		"fee0Unclaimed":            json.StringNode("fee0Unclaimed", p.FeeUnclaimed0),
		"fee1Unclaimed":            json.StringNode("fee1Unclaimed", p.FeeUnclaimed1),
	})
}

type RpcUnclaimedFee struct {
	LpPositionId uint64 `json:"lpPositionId"`
	Fee0         string `json:"fee0"`
	Fee1         string `json:"fee1"`
}

func (p RpcUnclaimedFee) JSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"lpPositionId": json.NumberNode("lpPositionId", float64(p.LpPositionId)),
		"fee0":         json.StringNode("fee0", p.Fee0),
		"fee1":         json.StringNode("fee1", p.Fee1),
	})
}

type ResponseQueryBase struct {
	Height    int64 `json:"height"`
	Timestamp int64 `json:"timestamp"`
}

func NewResponseQueryBase(height int64, timestamp int64) ResponseQueryBase {
	return ResponseQueryBase{
		Height:    height,
		Timestamp: timestamp,
	}
}

func (r ResponseQueryBase) JSON() *json.Node {
	return json.ObjectNode("", map[string]*json.Node{
		"height":    json.NumberNode("height", float64(r.Height)),
		"timestamp": json.NumberNode("timestamp", float64(r.Timestamp)),
	})
}

type ResponseApiGetPositions struct {
	Stat     ResponseQueryBase `json:"stat"`
	Response []RpcPosition     `json:"response"`
}

func NewResponseApiGetPositions(stat ResponseQueryBase, response []RpcPosition) ResponseApiGetPositions {
	return ResponseApiGetPositions{
		Stat:     stat,
		Response: response,
	}
}
