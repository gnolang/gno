package calibrate

// Native function calibration benchmarks for the IBC crypto stdlibs (bn254,
// cometbls, keccak256, merkle, modexp). Same harness conventions as
// native_bench_test.go — drive the GnoVM native dispatcher end-to-end so the
// measured ns/op feeds gen_native_table.py without special-casing.

import (
	"encoding/hex"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// ----- crypto/keccak256.sum256(data []byte) [32]byte -----

func benchKeccak256(b *testing.B, n int) {
	b.Helper()
	data := make([]byte, n)
	for i := range data {
		data[i] = byte(i)
	}
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, data)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "crypto/keccak256", "sum256"), nReturns: 1}
	b.ResetTimer()
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Keccak256_Sum256_0(b *testing.B)     { benchKeccak256(b, 0) }
func BenchmarkNative_Keccak256_Sum256_64(b *testing.B)    { benchKeccak256(b, 64) }
func BenchmarkNative_Keccak256_Sum256_256(b *testing.B)   { benchKeccak256(b, 256) }
func BenchmarkNative_Keccak256_Sum256_1024(b *testing.B)  { benchKeccak256(b, 1024) }
func BenchmarkNative_Keccak256_Sum256_4096(b *testing.B)  { benchKeccak256(b, 4096) }
func BenchmarkNative_Keccak256_Sum256_16384(b *testing.B) { benchKeccak256(b, 16384) }

// ----- crypto/modexp.modExp(base, exp, modulus []byte) []byte -----
//
// Modular exponentiation performs one modular squaring per exponent bit, each
// costing O(words(mod)²), so its cost is a product of two operands rather than
// a sum. It is priced with SizeModExpWork, which folds that product into a
// single metric (see gnovm/pkg/gnolang/native_gas.go):
//
//	work = (modular multiplications expNN performs) · (floor + ceil(len(mod)/8)²)
//
// Cost is NOT linear in 8·len(exp)·words², which is what this row charged on
// before the grid was extended. Two effects that metric did not model dominate
// at small exponents:
//
//   - The exponent regime. big.Int only takes the windowed Montgomery path when
//     the exponent exceeds one machine word (nat.go: `if len(y) > 1 && !slow`).
//     At 8 bytes or fewer it runs the generic loop, which divides on every
//     iteration rather than reducing in Montgomery form — ~2.5x the cost per
//     exponent bit. And the Montgomery loop walks whole words, so 9 bytes of
//     exponent costs what 16 does. Under the old metric 8/256 measured MORE
//     than 9/256 while being charged less.
//   - The dispatcher floor. Converting the operands and building the result is
//     linear in len(modulus) and entirely independent of the exponent, so at
//     small exponents it dominates a work term that is near zero. The exp=0 row
//     isolates it: 55us at a 1024-byte modulus for work=0.
//
// Both are now modelled rather than fitted around: the work metric counts the
// operations each expNN routine performs (see gnovm/pkg/gnolang/native_gas.go)
// and the row carries a second slope on len(modulus) for the dispatcher. So a
// re-fit is a single number — ns per modular multiplication — and the grid
// exists to check that it bounds every point rather than to fit a shape.
// gen_native_table.py's fit_modexp does exactly that; pass --hw-factor unless
// the run was taken on the reference Xeon.
//
// Both operands must stay at or below crypto/modexp.maxOperandLen (1024): above
// it X_modExp returns immediately and the bench would time the rejection path.
//
// RECORDED RUN — this is the calibration input for the shipped row, committed
// so the row can be re-derived rather than taken on faith. AMD Ryzen 7 7840U,
// -benchtime=100ms -count=5, median; run-to-run spread was 1.02-1.19x at every
// point but one (9/32, 1.32x). This is NOT the reference Xeon and is materially
// faster than it: rows in this same table calibrated on a reference-class Xeon
// measure 2.19-2.33x faster here (cometbls.verifyZKP 2632556 -> 1204572,
// bn254.pairingCheck 792691 -> 361591 at one pair and 1798041 -> 772655 at
// four, and modexp's own superseded anchor 6219920 -> 2686662). The shipped row
// projects by 2.3, the HIGH end: a larger factor charges more, so the high end
// is the safe choice, not the low one.
//
// These supersede a pre-#97 run. Data-backing byte slices changed what the
// dispatcher costs: Go2GnoValue no longer builds one TypedValue per byte, so
// the exponent-independent per-modulus-byte term fell from ~134 ns/byte to
// ~2.9. Any grid recorded before that commit overstates the dispatcher by ~50x
// and must not be mixed with these.
//
//	 exp    mod    work            ns/op      ns/work
//	0      32     0                      520          -
//	0      256    0                      768          -
//	0      1024   0                    1,716          -
//	1      256    43560               30,436      0.699
//	1      1024   657960             240,084      0.365
//	3      256    130680              98,887      0.757
//	3      1024   1973880            792,236      0.401
//	4      32     12960                8,454      0.652
//	4      256    174240             123,890      0.711
//	4      1024   2631840          1,102,211      0.419
//	8      32     25920               15,168      0.585
//	8      256    348480             239,457      0.687
//	8      1024   5263680          2,088,169      0.397
//	9      32     28998               17,203      0.593
//	9      256    389862             210,814      0.541
//	9      1024   5888742          2,612,594      0.444
//	16     256    389862             198,277      0.509
//	24     256    564102             278,681      0.494
//	32     32     54918               24,218      0.441
//	32     256    738342             352,113      0.477
//	32     1024   11152422         4,641,969      0.416
//	256    32     417798             165,587      0.396
//	256    1024   84843942        35,314,846      0.416
//	1024   32     1661958            724,654      0.436
//	1024   256    22344102        10,771,770      0.482
//	1024   1024   337500582       131,967,026      0.391
//
// The ns/work column still carries the dispatcher cost, which the row's second
// slope handles separately. Net of it the Montgomery points (exp > 8) hold to
// within a small factor of each other across a 128x range of exponent sizes,
// which is the check that the operation count matches the algorithm. The
// generic points (exp <= 8) are the ragged ones — nat.div's cost per word is
// set by hardware divide latency rather than a multiplication count, so that
// branch's constant is calibrated rather than counted.

// benchModExp is the symmetric diagonal, len(base)==len(exp)==len(mod)==n. Kept
// because the shipped slope is anchored to its N=256 point.
func benchModExp(b *testing.B, n int) {
	b.Helper()
	b.SetBytes(int64(n))
	benchModExpGrid(b, n, n)
}

func BenchmarkNative_ModExp_32(b *testing.B)  { benchModExp(b, 32) }
func BenchmarkNative_ModExp_64(b *testing.B)  { benchModExp(b, 64) }
func BenchmarkNative_ModExp_128(b *testing.B) { benchModExp(b, 128) }
func BenchmarkNative_ModExp_256(b *testing.B) { benchModExp(b, 256) }
func BenchmarkNative_ModExp_512(b *testing.B) { benchModExp(b, 512) }

// benchModExpGrid varies the exponent and modulus lengths independently. The
// base is held at the modulus length: a base wider than the modulus only adds
// the initial reduction, which the work metric deliberately does not model
// separately (maxOperandLen bounds it instead).
func benchModExpGrid(b *testing.B, expLen, modLen int) {
	b.Helper()
	base := make([]byte, modLen)
	exp := make([]byte, expLen)
	mod := make([]byte, modLen)
	for i := range base {
		base[i] = byte(i + 1)
	}
	for i := range exp {
		// Worst case for the metric, which charges on len(exp). Below the
		// Montgomery crossover big.Int walks the exponent bit by bit from the
		// highest set one (nat.go expNN: shift := nlz(v)+1), squaring every
		// iteration and multiplying again on each set bit — so an all-ones
		// exponent maximizes both the iteration count and the multiplies, and
		// is the only fill whose true bit length equals the charged one. A
		// patterned fill measures up to 1.75x cheaper at 4 bytes, which would
		// fit a slope that is not an upper bound. Above the crossover the
		// windowed Montgomery loop does 4 squarings plus one multiply per
		// 4-bit window whatever the window holds, so the fill is irrelevant
		// there and costs those points nothing.
		exp[i] = 0xFF
	}
	for i := range mod {
		mod[i] = 0xFF // large odd modulus
	}
	if modLen > 0 {
		mod[modLen-1] = 0xFD
	}
	m := newDispatchMachine(3)
	setBlockValueFromGo(m, 0, base)
	setBlockValueFromGo(m, 1, exp)
	setBlockValueFromGo(m, 2, mod)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "crypto/modexp", "modExp"), nReturns: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

// Grid naming is ModExpGrid_<expLen>_<modLen>. Sweeps small-exp/large-mod (the
// RSA-verify shape), large-exp/small-mod (the shape the old slope priced at
// zero), and the symmetric corner at the 1024-byte cap.
func BenchmarkNative_ModExpGrid_4_32(b *testing.B)      { benchModExpGrid(b, 4, 32) }
func BenchmarkNative_ModExpGrid_4_256(b *testing.B)     { benchModExpGrid(b, 4, 256) }
func BenchmarkNative_ModExpGrid_4_1024(b *testing.B)    { benchModExpGrid(b, 4, 1024) }
func BenchmarkNative_ModExpGrid_32_32(b *testing.B)     { benchModExpGrid(b, 32, 32) }
func BenchmarkNative_ModExpGrid_32_256(b *testing.B)    { benchModExpGrid(b, 32, 256) }
func BenchmarkNative_ModExpGrid_32_1024(b *testing.B)   { benchModExpGrid(b, 32, 1024) }
func BenchmarkNative_ModExpGrid_256_32(b *testing.B)    { benchModExpGrid(b, 256, 32) }
func BenchmarkNative_ModExpGrid_256_1024(b *testing.B)  { benchModExpGrid(b, 256, 1024) }
func BenchmarkNative_ModExpGrid_1024_32(b *testing.B)   { benchModExpGrid(b, 1024, 32) }
func BenchmarkNative_ModExpGrid_1024_256(b *testing.B)  { benchModExpGrid(b, 1024, 256) }
func BenchmarkNative_ModExpGrid_1024_1024(b *testing.B) { benchModExpGrid(b, 1024, 1024) }

// The sub-word band. big.Int only takes the windowed Montgomery path when the
// exponent exceeds one machine word (nat.go: `if len(y) > 1 && !slow`), so an
// exponent of 8 bytes or fewer runs the generic loop, which reduces by division
// on every iteration instead. That is the more expensive regime per exponent
// bit, and the grid above jumps 4 -> 32 straight over it, so the shipped slope
// was fit without a single measurement of the band where it is thinnest. 8 and
// 9 bracket the crossover; 3 is the RSA-verify shape (e=65537).
//
// expLen=0 is charged nothing by the work metric, while the native still
// allocates and fills len(modulus) bytes and the dispatcher still converts
// each operand. These points isolate that cost so it can be priced on its own
// slope rather than hidden in Base.
func BenchmarkNative_ModExpGrid_0_32(b *testing.B)   { benchModExpGrid(b, 0, 32) }
func BenchmarkNative_ModExpGrid_0_256(b *testing.B)  { benchModExpGrid(b, 0, 256) }
func BenchmarkNative_ModExpGrid_0_1024(b *testing.B) { benchModExpGrid(b, 0, 1024) }
func BenchmarkNative_ModExpGrid_1_256(b *testing.B)  { benchModExpGrid(b, 1, 256) }
func BenchmarkNative_ModExpGrid_1_1024(b *testing.B) { benchModExpGrid(b, 1, 1024) }
func BenchmarkNative_ModExpGrid_3_256(b *testing.B)  { benchModExpGrid(b, 3, 256) }
func BenchmarkNative_ModExpGrid_3_1024(b *testing.B) { benchModExpGrid(b, 3, 1024) }
func BenchmarkNative_ModExpGrid_8_32(b *testing.B)   { benchModExpGrid(b, 8, 32) }
func BenchmarkNative_ModExpGrid_8_256(b *testing.B)  { benchModExpGrid(b, 8, 256) }
func BenchmarkNative_ModExpGrid_8_1024(b *testing.B) { benchModExpGrid(b, 8, 1024) }
func BenchmarkNative_ModExpGrid_9_32(b *testing.B)   { benchModExpGrid(b, 9, 32) }
func BenchmarkNative_ModExpGrid_9_256(b *testing.B)  { benchModExpGrid(b, 9, 256) }
func BenchmarkNative_ModExpGrid_9_1024(b *testing.B) { benchModExpGrid(b, 9, 1024) }
func BenchmarkNative_ModExpGrid_16_256(b *testing.B) { benchModExpGrid(b, 16, 256) }
func BenchmarkNative_ModExpGrid_24_256(b *testing.B) { benchModExpGrid(b, 24, 256) }

// ----- crypto/bn254 EIP-196/197 precompile natives -----

// g1Doubling input (G + G = 2G): 128 bytes, fixed size.
var bn254G1AddInput = mustHex("0000000000000000000000000000000000000000000000000000000000000001" +
	"0000000000000000000000000000000000000000000000000000000000000002" +
	"0000000000000000000000000000000000000000000000000000000000000001" +
	"0000000000000000000000000000000000000000000000000000000000000002")

// G1Mul input: G with scalar 2. Always 96 bytes.
var bn254G1MulInput = mustHex("0000000000000000000000000000000000000000000000000000000000000001" +
	"0000000000000000000000000000000000000000000000000000000000000002" +
	"0000000000000000000000000000000000000000000000000000000000000002")

// Single pairing chunk taken from Ethereum's bn256Pairing_chfast1 test
// vector (g1 paired with g2). 192 bytes, valid + in-subgroup.
const bn254PairingChunkHex = "1c76476f4def4bb94541d57ebba1193381ffa7aa76ada664dd31c16024c43f59" +
	"3034dd2920f673e204fee2811c678745fc819b55d3e9d294e45c9b03a76aef41" +
	"209dd15ebff5d46c4bd888e51a93cf99a7329636c63514396b4a452003a35bf7" +
	"04bf11ca01483bfa8b34b43561848d28905960114c8ac04049af4b6315a41678" +
	"2bb8324af6cfc93537a2ad1a445cfd0ca2a71acd7ac41fadbf933c2a51be344d" +
	"120a2a4cf30c1bf9845f20c6fe39e07ea2cce61f0c9bb048165fe5e4de877550"

func benchBN254G1Add(b *testing.B) {
	b.Helper()
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, bn254G1AddInput)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "crypto/bn254", "g1Add"), nReturns: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func benchBN254G1Mul(b *testing.B) {
	b.Helper()
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, bn254G1MulInput)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "crypto/bn254", "g1Mul"), nReturns: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

// benchBN254PairingCheck benches the precompile with `pairs` (G1, G2) chunks
// concatenated. The cost scales linearly in the number of pairs — we expose a
// slope on input length (== 192·pairs bytes).
func benchBN254PairingCheck(b *testing.B, pairs int) {
	b.Helper()
	chunk := mustHex(bn254PairingChunkHex)
	input := make([]byte, 0, len(chunk)*pairs)
	for range pairs {
		input = append(input, chunk...)
	}
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, input)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "crypto/bn254", "pairingCheck"), nReturns: 1}
	b.ResetTimer()
	b.SetBytes(int64(len(input)))
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_BN254_G1Add(b *testing.B)          { benchBN254G1Add(b) }
func BenchmarkNative_BN254_G1Mul(b *testing.B)          { benchBN254G1Mul(b) }
func BenchmarkNative_BN254_PairingCheck_1(b *testing.B) { benchBN254PairingCheck(b, 1) }
func BenchmarkNative_BN254_PairingCheck_2(b *testing.B) { benchBN254PairingCheck(b, 2) }
func BenchmarkNative_BN254_PairingCheck_4(b *testing.B) { benchBN254PairingCheck(b, 4) }

// ----- crypto/merkle.* -----

func benchMerkleLeafHash(b *testing.B, n int) {
	b.Helper()
	leaf := make([]byte, n)
	for i := range leaf {
		leaf[i] = byte(i)
	}
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, leaf)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "crypto/merkle", "leafHash"), nReturns: 1}
	b.ResetTimer()
	b.SetBytes(int64(n))
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Merkle_LeafHash_0(b *testing.B)    { benchMerkleLeafHash(b, 0) }
func BenchmarkNative_Merkle_LeafHash_64(b *testing.B)   { benchMerkleLeafHash(b, 64) }
func BenchmarkNative_Merkle_LeafHash_256(b *testing.B)  { benchMerkleLeafHash(b, 256) }
func BenchmarkNative_Merkle_LeafHash_1024(b *testing.B) { benchMerkleLeafHash(b, 1024) }
func BenchmarkNative_Merkle_LeafHash_4096(b *testing.B) { benchMerkleLeafHash(b, 4096) }

// benchMerkleInnerHash benches innerHash with both operands at n bytes, so a
// single-slope fit over n yields the cost per byte of one operand — which is
// what the table charges on each of the two. Merkle tree use only ever passes
// 32-byte hashes, but nothing enforces that, so the sweep has to reach the
// sizes an caller can actually pass (see the innerHash comment in
// gnovm/stdlibs/native_gas.go).
func benchMerkleInnerHash(b *testing.B, n int) {
	b.Helper()
	left := make([]byte, n)
	right := make([]byte, n)
	for i := range left {
		left[i] = byte(i)
		right[i] = byte(i + 1)
	}
	m := newDispatchMachine(2)
	setBlockValueFromGo(m, 0, left)
	setBlockValueFromGo(m, 1, right)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "crypto/merkle", "innerHash"), nReturns: 1}
	b.ResetTimer()
	b.SetBytes(int64(2 * n))
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Merkle_InnerHash_32(b *testing.B)   { benchMerkleInnerHash(b, 32) }
func BenchmarkNative_Merkle_InnerHash_64(b *testing.B)   { benchMerkleInnerHash(b, 64) }
func BenchmarkNative_Merkle_InnerHash_256(b *testing.B)  { benchMerkleInnerHash(b, 256) }
func BenchmarkNative_Merkle_InnerHash_1024(b *testing.B) { benchMerkleInnerHash(b, 1024) }
func BenchmarkNative_Merkle_InnerHash_4096(b *testing.B) { benchMerkleInnerHash(b, 4096) }

// encodeMerkleItems builds the [4-byte BE count][4-byte BE len][data]…
// wire format consumed by hashFromByteSlices.
func encodeMerkleItems(items [][]byte) []byte {
	out := make([]byte, 4)
	out[0] = byte(len(items) >> 24)
	out[1] = byte(len(items) >> 16)
	out[2] = byte(len(items) >> 8)
	out[3] = byte(len(items))
	for _, it := range items {
		var hdr [4]byte
		hdr[0] = byte(len(it) >> 24)
		hdr[1] = byte(len(it) >> 16)
		hdr[2] = byte(len(it) >> 8)
		hdr[3] = byte(len(it))
		out = append(out, hdr[:]...)
		out = append(out, it...)
	}
	return out
}

func benchMerkleHashFromByteSlices(b *testing.B, nItems int) {
	b.Helper()
	items := make([][]byte, nItems)
	for i := range items {
		items[i] = []byte("item-" + string(rune('a'+(i%26))))
	}
	encoded := encodeMerkleItems(items)
	m := newDispatchMachine(1)
	setBlockValueFromGo(m, 0, encoded)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "crypto/merkle", "hashFromByteSlices"), nReturns: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Merkle_HashFromByteSlices_1(b *testing.B)  { benchMerkleHashFromByteSlices(b, 1) }
func BenchmarkNative_Merkle_HashFromByteSlices_8(b *testing.B)  { benchMerkleHashFromByteSlices(b, 8) }
func BenchmarkNative_Merkle_HashFromByteSlices_64(b *testing.B) { benchMerkleHashFromByteSlices(b, 64) }
func BenchmarkNative_Merkle_HashFromByteSlices_512(b *testing.B) {
	benchMerkleHashFromByteSlices(b, 512)
}

func benchMerkleVerifySimpleProof(b *testing.B, total int) {
	b.Helper()
	// 32-byte zero hash works for the Verify call path; the bench measures
	// dispatch + per-aunt work, not cryptographic validity.
	root := make([]byte, 32)
	leaf := []byte("leaf-payload")
	// log2(total) aunts, each 32 bytes.
	auntLen := 0
	for n := total; n > 1; n >>= 1 {
		auntLen++
	}
	aunts := make([]byte, auntLen*32)
	m := newDispatchMachine(5)
	setBlockValueFromGo(m, 0, root)
	setBlockValueFromGo(m, 1, leaf)
	setBlockValueFromGo(m, 2, 0)
	setBlockValueFromGo(m, 3, total)
	setBlockValueFromGo(m, 4, aunts)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "crypto/merkle", "verifySimpleProof"), nReturns: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Merkle_VerifySimpleProof_8(b *testing.B)  { benchMerkleVerifySimpleProof(b, 8) }
func BenchmarkNative_Merkle_VerifySimpleProof_64(b *testing.B) { benchMerkleVerifySimpleProof(b, 64) }
func BenchmarkNative_Merkle_VerifySimpleProof_1024(b *testing.B) {
	benchMerkleVerifySimpleProof(b, 1024)
}

// ----- crypto/cometbls.verifyZKP(chainID, tvh, hdr, zkp) -----
//
// Uses the same valid Groth16 vector that cometbls_test.go covers, so a
// happy-path proof is verified end-to-end on every call.

func BenchmarkNative_CometBLS_VerifyZKP(b *testing.B) {
	chainID := "union-devnet-1337"
	tvh := mustHex("20DDFE7A0F75C65D876316091ECCD494A54A2BB324C872015F73E528D53CB9C4")
	appHash := mustHex("EE7E3E58F98AC95D63CE93B270981DF3EE54CA367F8D521ED1F444717595CD36")
	header := make([]byte, 116)
	// height: 3405691582 (BE u64)
	putBE64(header[0:8], 3405691582)
	// timeSeconds: 1732205251
	putBE64(header[8:16], 1732205251)
	// timeNanos: 998131342 (BE u32)
	putBE32(header[16:20], 998131342)
	copy(header[20:52], tvh) // validators_hash
	copy(header[52:84], tvh) // next_validators_hash
	copy(header[84:116], appHash)

	proof := mustHex("03CF56142A1E03D2445A82100FEAF70C1CD95A731ED85792AFFF5792EC0BDD2108991BB56F9043A269F88903DE616A9AB99A3C5AB778E566744B060456C5616C" +
		"06BCE7F1930421768C2CBD79F88D08EC3A52D7C9A867064E973064385E9C945E02951190DD7CE1662546733DD540188C96E608CA750FEF36B39E2577833634C7" +
		"0AE6F1A6D00DC6C21446AAF285EF35D944E8782B131300574F9A889C7E708A2325E9A78013BBE869D38B19C602DAF69644C77D177E99ED76398BCEE13C61FDBF" +
		"2E178A5BA028A36033E54D1D9A0071E82E04079A5305347EBAC6D66F6EBFA48B1DA1BF9DC5A51EFA292E1DC7B85D26F18422EB386C48CA75434039764448BB96" +
		"268DDC2CF683DDCA4BD83DF21C5631CF784375EEBE77EABC2DE77886BF1D48392C9C52E063B4A7131EAB9ABBA12A9F26888BC37366D41AC7D4BAC0BF6755ACB0" +
		"09BF9F36F380B6D0EEAABF066503A1B6E01DCC965D968D7694E01B1755E6BDD21C7A80B41682748F9B7151714BE34AA79AAD48BBB2A84525F6CDF812658C6E4F")

	m := newDispatchMachine(4)
	setBlockValueFromGo(m, 0, chainID)
	setBlockValueFromGo(m, 1, tvh)
	setBlockValueFromGo(m, 2, header)
	setBlockValueFromGo(m, 3, proof)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "crypto/cometbls", "verifyZKP"), nReturns: 1}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

// --- helpers ---

func mustHex(s string) []byte {
	out, err := hex.DecodeString(s)
	if err != nil {
		panic("invalid hex: " + err.Error())
	}
	return out
}

func putBE64(dst []byte, v uint64) {
	dst[0] = byte(v >> 56)
	dst[1] = byte(v >> 48)
	dst[2] = byte(v >> 40)
	dst[3] = byte(v >> 32)
	dst[4] = byte(v >> 24)
	dst[5] = byte(v >> 16)
	dst[6] = byte(v >> 8)
	dst[7] = byte(v)
}

func putBE32(dst []byte, v uint32) {
	dst[0] = byte(v >> 24)
	dst[1] = byte(v >> 16)
	dst[2] = byte(v >> 8)
	dst[3] = byte(v)
}

// Silence unused gno import on platforms that strip it.
var _ = gno.Name("")
