package position

import (
	"std"
	"testing"

	"gno.land/p/demo/json"
	"gno.land/p/demo/uassert"
)

func setupPositions(t *testing.T) {
	t.Helper()

	MakeMintPositionWithoutFee(t)
}

func TestApiGetPositions(t *testing.T) {
	setupPositions(t)
	std.TestSetRealm(std.NewUserRealm(adminAddr))

	result := ApiGetPositions()

	root, err := json.Unmarshal([]byte(result))
	if err != nil {
		panic(err.Error())
	}

	response, err := root.GetKey("response")
	if err != nil {
		panic(err.Error())
	}
	uassert.NoError(t, err)
	uassert.Equal(t, 1, response.Size())

	response, err = response.GetIndex(0)
	if err != nil {
		panic(err.Error())
	}

	lpPositionId, err := response.GetKey("lpPositionId")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "1", lpPositionId.String())

	burned, err := response.GetKey("burned")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "false", burned.String())

	owner, err := response.GetKey("owner")
	if err != nil {
		panic(err.Error())
	}
	ownerAddr := "\"g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d\""
	uassert.Equal(t, ownerAddr, owner.String())

	operator, err := response.GetKey("operator")
	if err != nil {
		panic(err.Error())
	}
	operatorAddr := "\"g100000000000000000000000000000000dnmcnx\""
	uassert.Equal(t, operatorAddr, operator.String())

	poolKey, err := response.GetKey("poolKey")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500\"", poolKey.String())

	tickLower, err := response.GetKey("tickLower")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "-887270", tickLower.String())

	tickUpper, err := response.GetKey("tickUpper")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "887270", tickUpper.String())

	liquidity, err := response.GetKey("liquidity")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"50000\"", liquidity.String())

	feeGrowthInside0LastX128, err := response.GetKey("feeGrowthInside0LastX128")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", feeGrowthInside0LastX128.String())

	feeGrowthInside1LastX128, err := response.GetKey("feeGrowthInside1LastX128")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", feeGrowthInside1LastX128.String())

	token0Owed, err := response.GetKey("token0Owed")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", token0Owed.String())

	token1Owed, err := response.GetKey("token1Owed")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", token1Owed.String())

	token0Balance, err := response.GetKey("token0Balance")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"50000\"", token0Balance.String())

	token1Balance, err := response.GetKey("token1Balance")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"50000\"", token1Balance.String())

	fee0Unclaimed, err := response.GetKey("fee0Unclaimed")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", fee0Unclaimed.String())

	fee1Unclaimed, err := response.GetKey("fee1Unclaimed")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", fee1Unclaimed.String())
}

func TestApiGetPositionsByPoolPath(t *testing.T) {
	result := ApiGetPositionsByPoolPath("gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500")

	root, err := json.Unmarshal([]byte(result))
	if err != nil {
		panic(err.Error())
	}

	response, err := root.GetKey("response")
	if err != nil {
		panic(err.Error())
	}
	uassert.NoError(t, err)
	uassert.Equal(t, 1, response.Size())

	response, err = response.GetIndex(0)
	if err != nil {
		panic(err.Error())
	}

	lpPositionId, err := response.GetKey("lpPositionId")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "1", lpPositionId.String())

	burned, err := response.GetKey("burned")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "false", burned.String())

	owner, err := response.GetKey("owner")
	if err != nil {
		panic(err.Error())
	}
	ownerAddr := "\"g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d\""
	uassert.Equal(t, ownerAddr, owner.String())

	operator, err := response.GetKey("operator")
	if err != nil {
		panic(err.Error())
	}
	operatorAddr := "\"g100000000000000000000000000000000dnmcnx\""
	uassert.Equal(t, operatorAddr, operator.String())

	poolKey, err := response.GetKey("poolKey")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500\"", poolKey.String())

	tickLower, err := response.GetKey("tickLower")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "-887270", tickLower.String())

	tickUpper, err := response.GetKey("tickUpper")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "887270", tickUpper.String())

	liquidity, err := response.GetKey("liquidity")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"50000\"", liquidity.String())

	feeGrowthInside0LastX128, err := response.GetKey("feeGrowthInside0LastX128")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", feeGrowthInside0LastX128.String())

	feeGrowthInside1LastX128, err := response.GetKey("feeGrowthInside1LastX128")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", feeGrowthInside1LastX128.String())

	token0Owed, err := response.GetKey("token0Owed")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", token0Owed.String())

	token1Owed, err := response.GetKey("token1Owed")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", token1Owed.String())

	token0Balance, err := response.GetKey("token0Balance")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"50000\"", token0Balance.String())

	token1Balance, err := response.GetKey("token1Balance")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"50000\"", token1Balance.String())

	fee0Unclaimed, err := response.GetKey("fee0Unclaimed")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", fee0Unclaimed.String())

	fee1Unclaimed, err := response.GetKey("fee1Unclaimed")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", fee1Unclaimed.String())
}

func TestApiGetPositionsByAddress(t *testing.T) {
	result := ApiGetPositionsByAddress(adminAddr)

	root, err := json.Unmarshal([]byte(result))
	if err != nil {
		panic(err.Error())
	}

	response, err := root.GetKey("response")
	if err != nil {
		panic(err.Error())
	}
	uassert.NoError(t, err)
	uassert.Equal(t, 1, response.Size())

	response, err = response.GetIndex(0)
	if err != nil {
		panic(err.Error())
	}

	lpPositionId, err := response.GetKey("lpPositionId")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "1", lpPositionId.String())

	burned, err := response.GetKey("burned")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "false", burned.String())

	owner, err := response.GetKey("owner")
	if err != nil {
		panic(err.Error())
	}
	ownerAddr := "\"g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d\""
	uassert.Equal(t, ownerAddr, owner.String())

	operator, err := response.GetKey("operator")
	if err != nil {
		panic(err.Error())
	}
	operatorAddr := "\"g100000000000000000000000000000000dnmcnx\""
	uassert.Equal(t, operatorAddr, operator.String())

	poolKey, err := response.GetKey("poolKey")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"gno.land/r/onbloc/bar:gno.land/r/onbloc/foo:500\"", poolKey.String())

	tickLower, err := response.GetKey("tickLower")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "-887270", tickLower.String())

	tickUpper, err := response.GetKey("tickUpper")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "887270", tickUpper.String())

	liquidity, err := response.GetKey("liquidity")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"50000\"", liquidity.String())

	feeGrowthInside0LastX128, err := response.GetKey("feeGrowthInside0LastX128")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", feeGrowthInside0LastX128.String())

	feeGrowthInside1LastX128, err := response.GetKey("feeGrowthInside1LastX128")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", feeGrowthInside1LastX128.String())

	token0Owed, err := response.GetKey("token0Owed")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", token0Owed.String())

	token1Owed, err := response.GetKey("token1Owed")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", token1Owed.String())

	token0Balance, err := response.GetKey("token0Balance")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"50000\"", token0Balance.String())

	token1Balance, err := response.GetKey("token1Balance")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"50000\"", token1Balance.String())

	fee0Unclaimed, err := response.GetKey("fee0Unclaimed")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", fee0Unclaimed.String())

	fee1Unclaimed, err := response.GetKey("fee1Unclaimed")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", fee1Unclaimed.String())
}

func TestApiGetPositionsUnclaimedFee(t *testing.T) {
	result := ApiGetPositionsUnclaimedFee()

	root, err := json.Unmarshal([]byte(result))
	if err != nil {
		panic(err.Error())
	}

	response, err := root.GetKey("response")
	if err != nil {
		panic(err.Error())
	}
	uassert.NoError(t, err)
	uassert.Equal(t, 1, response.Size())

	response, err = response.GetIndex(0)
	if err != nil {
		panic(err.Error())
	}

	lpPositionId, err := response.GetKey("lpPositionId")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "1", lpPositionId.String())

	fee0Unclaimed, err := response.GetKey("fee0")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", fee0Unclaimed.String())

	fee1Unclaimed, err := response.GetKey("fee1")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", fee1Unclaimed.String())
}

func TestApiGetPositionUnclaimedFeeByLpPositionId(t *testing.T) {
	setupPositions(t)
	result := ApiGetPositionUnclaimedFeeByLpPositionId(1)

	root, err := json.Unmarshal([]byte(result))
	if err != nil {
		panic(err.Error())
	}

	response, err := root.GetKey("response")
	if err != nil {
		panic(err.Error())
	}
	uassert.NoError(t, err)
	uassert.Equal(t, 1, response.Size())

	response, err = response.GetIndex(0)
	if err != nil {
		panic(err.Error())
	}

	lpPositionId, err := response.GetKey("lpPositionId")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "1", lpPositionId.String())

	fee0Unclaimed, err := response.GetKey("fee0")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", fee0Unclaimed.String())

	fee1Unclaimed, err := response.GetKey("fee1")
	if err != nil {
		panic(err.Error())
	}
	uassert.Equal(t, "\"0\"", fee1Unclaimed.String())
}
