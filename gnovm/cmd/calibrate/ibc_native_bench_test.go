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
// Modular exponentiation runs one modular squaring per exponent bit, and
// each squaring is quadratic in the modulus, so cost tracks
// len(exp)·len(mod)². Sampling only the diagonal len(exp) == len(mod)
// makes the two lengths indistinguishable to a fitter, and pricing the
// modulus alone leaves the exponent free: a small modulus with a huge
// exponent is cheap to buy and expensive to run.
//
// The grid below walks the diagonal, then holds the modulus at 256 bytes
// while sweeping the exponent and holds the exponent at 256 bytes while
// sweeping the modulus, so each length varies with the other pinned. Bench
// name encodes both as _E<exp>_M<mod>.
//
// The base is sized to the modulus throughout. It is not a fitted
// parameter: NativeGasInfo carries at most two pre-call slopes, and those
// go to the two lengths that dominate.
func benchModExp(b *testing.B, expLen, modLen int) {
	b.Helper()
	base := make([]byte, modLen)
	mod := make([]byte, modLen)
	for i := range modLen {
		base[i] = byte(i + 1)
		mod[i] = 0xFF // large odd modulus
	}
	if modLen > 0 {
		mod[modLen-1] = 0xFD
	}
	exp := make([]byte, expLen)
	for i := range expLen {
		exp[i] = byte(i + 3)
	}
	m := newDispatchMachine(3)
	setBlockValueFromGo(m, 0, base)
	setBlockValueFromGo(m, 1, exp)
	setBlockValueFromGo(m, 2, mod)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "crypto/modexp", "modExp"), nReturns: 1}
	b.ResetTimer()
	b.SetBytes(int64(expLen + modLen))
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

// Diagonal: len(exp) == len(mod).
func BenchmarkNative_ModExp_E32_M32(b *testing.B)   { benchModExp(b, 32, 32) }
func BenchmarkNative_ModExp_E64_M64(b *testing.B)   { benchModExp(b, 64, 64) }
func BenchmarkNative_ModExp_E128_M128(b *testing.B) { benchModExp(b, 128, 128) }
func BenchmarkNative_ModExp_E256_M256(b *testing.B) { benchModExp(b, 256, 256) }
func BenchmarkNative_ModExp_E512_M512(b *testing.B) { benchModExp(b, 512, 512) }

// Exponent sweep, modulus pinned at 256 bytes.
func BenchmarkNative_ModExp_E32_M256(b *testing.B)   { benchModExp(b, 32, 256) }
func BenchmarkNative_ModExp_E1024_M256(b *testing.B) { benchModExp(b, 1024, 256) }

// Modulus sweep, exponent pinned at 256 bytes.
func BenchmarkNative_ModExp_E256_M32(b *testing.B)  { benchModExp(b, 256, 32) }
func BenchmarkNative_ModExp_E256_M512(b *testing.B) { benchModExp(b, 256, 512) }

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

// innerHash accepts []byte of any length on both sides, so both must be
// sized. Benching only the 32+32B digest case yields a single point, which
// the fitter can only express as SizeFlat, and a flat charge lets a realm
// hash megabytes for the price of 64 bytes.
//
// Sizing left and right together is still not enough: on the diagonal
// len(left) == len(right) the two columns of the design matrix are
// identical, so a two-slope fit is rank-deficient and cannot attribute
// cost to either parameter. The off-diagonal pairs below break that tie.
// Bench name encodes both lengths as _L<left>_R<right> so the fitter reads
// them as two independent arguments.
func benchMerkleInnerHash(b *testing.B, leftLen, rightLen int) {
	b.Helper()
	left := make([]byte, leftLen)
	for i := range leftLen {
		left[i] = byte(i + 1)
	}
	right := make([]byte, rightLen)
	for i := range rightLen {
		right[i] = byte(i + 2)
	}
	m := newDispatchMachine(2)
	setBlockValueFromGo(m, 0, left)
	setBlockValueFromGo(m, 1, right)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "crypto/merkle", "innerHash"), nReturns: 1}
	b.ResetTimer()
	b.SetBytes(int64(leftLen + rightLen))
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Merkle_InnerHash_L32_R32(b *testing.B) { benchMerkleInnerHash(b, 32, 32) }
func BenchmarkNative_Merkle_InnerHash_L256_R256(b *testing.B) {
	benchMerkleInnerHash(b, 256, 256)
}

func BenchmarkNative_Merkle_InnerHash_L1024_R1024(b *testing.B) {
	benchMerkleInnerHash(b, 1024, 1024)
}

func BenchmarkNative_Merkle_InnerHash_L4096_R4096(b *testing.B) {
	benchMerkleInnerHash(b, 4096, 4096)
}
func BenchmarkNative_Merkle_InnerHash_L32_R4096(b *testing.B) { benchMerkleInnerHash(b, 32, 4096) }
func BenchmarkNative_Merkle_InnerHash_L4096_R32(b *testing.B) { benchMerkleInnerHash(b, 4096, 32) }
func BenchmarkNative_Merkle_InnerHash_L32_R1024(b *testing.B) { benchMerkleInnerHash(b, 32, 1024) }
func BenchmarkNative_Merkle_InnerHash_L1024_R32(b *testing.B) { benchMerkleInnerHash(b, 1024, 32) }

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
