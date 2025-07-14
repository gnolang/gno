package consts

import (
	"std"
)

// INITIAL ADDRESSES
const (
	ADMIN   std.Address = "g17290cwvmrapvp869xfnhhawa8sm9edpufzat7d"
	DEV_OPS std.Address = "g1mjvd83nnjee3z2g7683er55me9f09688pd4mj9"
)

// WRAP & UNWRAP
const (
	GNOT           string = "gnot"
	UGNOT          string = "ugnot"
	WRAPPED_WUGNOT string = "gno.land/r/demo/wugnot"

	// defined in https://github.com/gnolang/gno/blob/81a88a2976ba9f2f9127ebbe7fb7d1e1f7fa4bd4/examples/gno.land/r/demo/wugnot/wugnot.gno#L19
	UGNOT_MIN_DEPOSIT_TO_WRAP uint64 = 1000
)

// INITIAL CONTRACT PATH & ADDRESS
const (
	POOL_PATH string      = "gno.land/r/gnoswap/v1/pool"
	POOL_ADDR std.Address = "g148tjamj80yyrm309z7rk690an22thd2l3z8ank"

	POSITION_PATH string      = "gno.land/r/gnoswap/v1/position"
	POSITION_ADDR std.Address = "g1q646ctzhvn60v492x8ucvyqnrj2w30cwh6efk5"

	ROUTER_PATH string      = "gno.land/r/gnoswap/v1/router"
	ROUTER_ADDR std.Address = "g1lm2l7tf49h3mykesct7rhfml30yx8dw5xrval7"

	STAKER_PATH string      = "gno.land/r/gnoswap/v1/staker"
	STAKER_ADDR std.Address = "g1cceshmzzlmrh7rr3z30j2t5mrvsq9yccysw9nu"

	GNS_PATH string      = "gno.land/r/gnoswap/v1/gns"
	GNS_ADDR std.Address = "g1jgqwaa2le3yr63d533fj785qkjspumzv22ys5m"

	GNFT_PATH string      = "gno.land/r/gnoswap/v1/gnft"
	GNFT_ADDR std.Address = "g1wxv2rdfn53qc84nt3nn646f9yh3nly8lm7j89t"

	WUGNOT_PATH string      = "gno.land/r/demo/wugnot"
	WUGNOT_ADDR std.Address = "g1pf6dv9fjk3rn0m4jjcne306ga4he3mzmupfjl6"

	EMISSION_PATH string      = "gno.land/r/gnoswap/v1/emission"
	EMISSION_ADDR std.Address = "g10xg6559w9e93zfttlhvdmaaa0er3zewcr7nh20"

	PROTOCOL_FEE_PATH string      = "gno.land/r/gnoswap/v1/protocol_fee"
	PROTOCOL_FEE_ADDR std.Address = "g1f7wpek7q67tkns27sw495u5yuu3a5wwjxw5l6l"

	COMMUNITY_POOL_PATH string      = "gno.land/r/gnoswap/v1/community_pool"
	COMMUNITY_POOL_ADDR std.Address = "g100fnnlz5eh87p5hvwt8pf279lxaelm8k8md049"

	GOV_XGNS_PATH string      = "gno.land/r/gnoswap/v1/gov/xgns"
	GOV_XGNS_ADDR std.Address = "g1wwh55uwzlz2zzr2qcvvxf83qhcvmx2t8779l9r"

	GOV_STAKER_PATH string      = "gno.land/r/gnoswap/v1/gov/staker"
	GOV_STAKER_ADDR std.Address = "g17e3ykyqk9jmqe2y9wxe9zhep3p7cw56davjqwa"

	GOV_GOVERNANCE_PATH string      = "gno.land/r/gnoswap/v1/gov/governance"
	GOV_GOVERNANCE_ADDR std.Address = "g17s8w2ve7k85fwfnrk59lmlhthkjdted8whvqxd"

	COMMON_PATH string      = "gno.land/r/gnoswap/v1/common"
	COMMON_ADDR std.Address = "g14ytarn5u7h3xywygt8hzhs3m23frljz72ta9xk"

	LAUNCHPAD_PATH string      = "gno.land/r/gnoswap/v1/launchpad"
	LAUNCHPAD_ADDR std.Address = "g122mau2lp2rc0scs8d27pkkuys4w54mdy2tuer3"

	ACCESS_PATH string      = "gno.land/r/gnoswap/v1/access"
	ACCESS_ADDR std.Address = "g1yyw4t7pywpgpq6z2p745y05qejwur97xud4arw"

	HALT_PATH string      = "gno.land/r/gnoswap/v1/halt"
	HALT_ADDR std.Address = "g1q67vstyzqycl75yv7ern98n4u8qrgc8hxkmsxt"
)

// NUMBER
const (
	// calculated by https://mathiasbynens.be/demo/integer-range
	MAX_UINT8 string = "255"
	UINT8_MAX uint8  = 255

	MAX_UINT16 string = "65535"
	UINT16_MAX uint16 = 65535

	MAX_UINT32 string = "4294967295"
	UINT32_MAX uint32 = 4294967295

	MAX_UINT64 string = "18446744073709551615"
	UINT64_MAX uint64 = 18446744073709551615

	MAX_INT64 string = "9223372036854775807"
	INT64_MAX int64  = 9223372036854775807

	MAX_UINT128 string = "340282366920938463463374607431768211455"
	MAX_UINT160 string = "1461501637330902918203684832716283019655932542975"
	MAX_UINT256 string = "115792089237316195423570985008687907853269984665640564039457584007913129639935"

	MAX_INT128 string = "170141183460469231731687303715884105727"
	MAX_INT256 string = "57896044618658097711785492504343953926634992332820282019728792003956564819967"

	// Tick Related
	MIN_TICK int32 = -887272
	MAX_TICK int32 = 887272

	MIN_SQRT_RATIO string = "4295128739"                                        // same as TickMathGetSqrtRatioAtTick(MIN_TICK)
	MAX_SQRT_RATIO string = "1461446703485210103287273052203988822378723970342" // same as TickMathGetSqrtRatioAtTick(MAX_TICK)

	MIN_PRICE string = "4295128740"                                        // MIN_SQRT_RATIO + 1
	MAX_PRICE string = "1461446703485210103287273052203988822378723970341" // MAX_SQRT_RATIO - 1

	// ETC
	Q64  string = "18446744073709551616"                    // 2 ** 64
	Q96  string = "79228162514264337593543950336"           // 2 ** 96
	Q128 string = "340282366920938463463374607431768211456" // 2 ** 128

	Q96_RESOLUTION  uint = 96
	Q128_RESOLUTION uint = 128
	Q160_RESOLUTION uint = 160
)

// TIME
const (
	SECONDS_PER_DAY = 86400
)

// BLOCK TIME
const (
	// Default block generation interval, in milliseconds, used to estimate
	// how many blocks will be produced over a given timeframe.
	// GnoSwap’s emission logic relies on this as an initial value,
	// but actual block times may vary (e.g. due to network conditions).
	// Governance or an admin can adjust it dynamically by calling
	// [SetAvgBlockTimeInMs](https://github.com/gnoswap-labs/gnoswap/blob/a29e0f994466430618548ae992cca11a52f5102a/contract/r/gnoswap/gns/halving.gno#L359) to keep emission schedules accurate.
	BLOCK_GENERATION_INTERVAL int64 = 2000
	MILLISECONDS_PER_SECOND   int64 = 1000
)

// ETCs
const (
	// REF: https://github.com/gnolang/gno/pull/2401#discussion_r1648064219
	ZERO_ADDRESS std.Address = "g100000000000000000000000000000000dnmcnx"
)
