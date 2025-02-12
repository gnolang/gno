package common

import (
	"testing"

	"gno.land/p/demo/uassert"

	u256 "gno.land/p/gnoswap/uint256"
)

var (
	MIN_TICK       = int32(-887272)
	MIN_SQRT_RATIO = "4295128739"

	MAX_TICK       = int32(887272)
	MAX_SQRT_RATIO = "1461446703485210103287273052203988822378723970342"
)

func TestTickMathGetSqrtRatioAtTick(t *testing.T) {
	t.Run("throws for too low", func(t *testing.T) {
		tick := MIN_TICK - 1

		uassert.PanicsWithMessage(
			t,
			"[GNOSWAP-COMMON-003] value out of range || abs tick is out of range (larger than 887272), abs tick: 887273",
			func() {
				TickMathGetSqrtRatioAtTick(tick)
			},
		)
	})

	t.Run("throws for too high", func(t *testing.T) {
		tick := MAX_TICK + 1

		uassert.PanicsWithMessage(
			t,
			"[GNOSWAP-COMMON-003] value out of range || abs tick is out of range (larger than 887272), abs tick: 887273",
			func() {
				TickMathGetSqrtRatioAtTick(tick)
			},
		)
	})

	t.Run("min tick", func(t *testing.T) {
		tick := MIN_TICK
		sqrtPriceX96 := TickMathGetSqrtRatioAtTick(tick)
		uassert.Equal(t, "4295128739", sqrtPriceX96.ToString())
	})

	t.Run("min tick + 1", func(t *testing.T) {
		tick := MIN_TICK + 1
		sqrtPriceX96 := TickMathGetSqrtRatioAtTick(tick)
		uassert.Equal(t, "4295343490", sqrtPriceX96.ToString())
	})

	t.Run("min tick ratio is less than js implementation", func(t *testing.T) {
		// encodePriceSqrt(1, BigNumber.from(2).pow(127)) = 6085630636
		sqrtPriceX96 := TickMathGetSqrtRatioAtTick(MIN_TICK)
		uassert.False(t, sqrtPriceX96.Cmp(u256.MustFromDecimal("6085630636")) > 0, "should be less than 6085630636")
	})

	t.Run("max tick ratio is greater than js implementation", func(t *testing.T) {
		// encodePriceSqrt(BigNumber.from(2).pow(127),1) = 1033437718471923706666374484006904511252097097914
		sqrtPriceX96 := TickMathGetSqrtRatioAtTick(MAX_TICK)
		uassert.False(t, sqrtPriceX96.Cmp(u256.MustFromDecimal("1033437718471923706666374484006904511252097097914")) < 0, "should be greater than 1033437718471923706666374484006904511252097097914")
	})

	t.Run("max tick ratio is equal to js implementation", func(t *testing.T) {
		sqrtPriceX96 := TickMathGetSqrtRatioAtTick(MAX_TICK)
		uassert.Equal(t, "1461446703485210103287273052203988822378723970342", sqrtPriceX96.ToString())
	})

	t.Run("multiple ticks", func(t *testing.T) {
		absTicks := []int32{50, 100, 250, 500, 1000, 2500, 3000, 4000, 5000, 50000, 150000, 250000, 500000, 738203}
		expectedResults := []string{
			"79030349367926598376800521322",
			"79426470787362580746886972461",
			"78833030112140176575862854579",
			"79625275426524748796330556128",
			"78244023372248365697264290337",
			"80224679980005306637834519095",
			"77272108795590369356373805297",
			"81233731461783161732293370115",
			"75364347830767020784054125655",
			"83290069058676223003182343270",
			"69919044979842180277688105136",
			"89776708723587163891445672585",
			"68192822843687888778582228483",
			"92049301871182272007977902845",
			"64867181785621769311890333195",
			"96768528593268422080558758223",
			"61703726247759831737814779831",
			"101729702841318637793976746270",
			"6504256538020985011912221507",
			"965075977353221155028623082916",
			"43836292794701720435367485",
			"143194173941309278083010301478497",
			"295440463448801648376846",
			"21246587762933397357449903968194344",
			"1101692437043807371",
			"5697689776495288729098254600827762987878",
			"7409801140451",
			"847134979253254120489401328389043031315994541",
		}

		for i, absTick := range absTicks {
			for j, tick := range []int32{-absTick, absTick} {
				result := TickMathGetSqrtRatioAtTick(tick)
				uassert.Equal(t, expectedResults[i*2+j], result.ToString())
			}
		}
	})

	t.Run("min_sqrt_ratio", func(t *testing.T) {
		sqrtPriceX96 := TickMathGetSqrtRatioAtTick(MIN_TICK)
		uassert.Equal(t, "4295128739", sqrtPriceX96.ToString())
	})

	t.Run("max_sqrt_ratio", func(t *testing.T) {
		sqrtPriceX96 := TickMathGetSqrtRatioAtTick(MAX_TICK)
		uassert.Equal(t, MAX_SQRT_RATIO, sqrtPriceX96.ToString())
	})

	t.Run("throws for too low sqrt_ratio", func(t *testing.T) {
		sqrtPriceX96 := u256.MustFromDecimal(MIN_SQRT_RATIO)
		sqrtPriceX96.Sub(sqrtPriceX96, u256.One())

		uassert.PanicsWithMessage(
			t,
			"[GNOSWAP-COMMON-003] value out of range || sqrtPriceX96 is out of range, sqrtPriceX96: 4295128738",
			func() {
				TickMathGetTickAtSqrtRatio(sqrtPriceX96)
			},
		)
	})

	t.Run("throws for too high sqrt_ratio", func(t *testing.T) {
		sqrtPriceX96 := u256.MustFromDecimal(MAX_SQRT_RATIO)
		sqrtPriceX96.Add(sqrtPriceX96, u256.One())

		uassert.PanicsWithMessage(
			t,
			"[GNOSWAP-COMMON-003] value out of range || sqrtPriceX96 is out of range, sqrtPriceX96: 1461446703485210103287273052203988822378723970343",
			func() {
				TickMathGetTickAtSqrtRatio(sqrtPriceX96)
			},
		)
	})

	t.Run("ratio of min tick", func(t *testing.T) {
		sqrtPriceX96 := TickMathGetSqrtRatioAtTick(MIN_TICK)
		uassert.Equal(t, MIN_SQRT_RATIO, sqrtPriceX96.ToString())
	})

	t.Run("ratio of min tick + 1", func(t *testing.T) {
		sqrtPriceX96 := u256.MustFromDecimal("4295343490")
		uassert.Equal(t, MIN_TICK+1, TickMathGetTickAtSqrtRatio(sqrtPriceX96))
	})

	t.Run("multiple sqrt_ratios", func(t *testing.T) {
		ratios := []string{
			"4295128739",
			"79228162514264337593543950336000000",
			"79228162514264337593543950336000",
			"9903520314283042199192993792",
			"28011385487393069959365969113",
			"56022770974786139918731938227",
			"79228162514264337593543950336",
			"112045541949572279837463876454",
			"224091083899144559674927752909",
			"633825300114114700748351602688",
			"79228162514264337593543950",
			"79228162514264337593543",
			"1461446703485210103287273052203988822378723970341",
		}
		expectedResults := []int32{
			-887272,
			276324,
			138162,
			-41591,
			-20796,
			-6932,
			0, // got -1, expected 0
			6931,
			20795,
			41590,
			-138163,
			-276325,
			887271,
		}

		for i, ratio := range ratios {
			sqrtPriceX96 := u256.MustFromDecimal(ratio)
			rst := TickMathGetTickAtSqrtRatio(sqrtPriceX96)
			uassert.Equal(t, expectedResults[i], rst)
		}
	})

	t.Run("ratio closest to max tick", func(t *testing.T) {
		maxSqrtRatio := u256.MustFromDecimal(MAX_SQRT_RATIO)
		maxSqrtRatioSub1 := maxSqrtRatio.Sub(maxSqrtRatio, u256.One())

		maxTick := MAX_TICK
		maxTickSub1 := maxTick - 1

		rst := TickMathGetTickAtSqrtRatio(maxSqrtRatioSub1)
		uassert.Equal(t, maxTickSub1, rst)
	})
}
