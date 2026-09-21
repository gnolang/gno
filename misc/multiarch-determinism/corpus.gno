// Package main drives a fixed corpus of gno stdlib operations and emits
// one canonical "<op> <args_hex...> <output_hex>" line per case to
// stdout. The output is intended to be byte-identical when this program
// is executed by a `gno` binary cross-compiled for and run on different
// CPU architectures — any divergence indicates non-determinism in a
// gno stdlib that gnovm exposes to user code.
//
// Run via `gno run --root-dir <repo-root> corpus.gno`.
package main

import (
	"crypto/bech32"
	"crypto/bn254"
	"crypto/chacha20"
	"crypto/ed25519"
	"crypto/keccak256"
	"crypto/merkle"
	"crypto/modexp"
	"crypto/sha256"
	"crypto/subtle"
	"hash/adler32"
	"strconv"
)

const hexChars = "0123456789abcdef"

func toHex(b []byte) string {
	out := make([]byte, len(b)*2)
	for i, c := range b {
		out[i*2] = hexChars[c>>4]
		out[i*2+1] = hexChars[c&0x0f]
	}
	return string(out)
}

func fromHex(s string) []byte {
	if len(s)%2 != 0 {
		panic("odd hex length")
	}
	out := make([]byte, len(s)/2)
	for i := 0; i < len(out); i++ {
		hi := hexVal(s[i*2])
		lo := hexVal(s[i*2+1])
		out[i] = hi<<4 | lo
	}
	return out
}

func hexVal(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	}
	panic("bad hex char")
}

func emit(parts ...string) {
	// Deterministic single-line output. println goes to stderr in gno;
	// use a build-it-yourself string and print via println-equivalent.
	// We use fmt-free formatting to avoid pulling fmt's reflection.
	line := ""
	for i, p := range parts {
		if i > 0 {
			line += " "
		}
		line += p
	}
	println(line) //nolint
}

// pattern builders for deterministic inputs.

func zeros(n int) []byte { return make([]byte, n) }

func ones(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = 0xff
	}
	return b
}

func seq(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i)
	}
	return b
}

func runSHA256() {
	inputs := [][]byte{
		nil,
		{0x00},
		seq(31), seq(32), seq(33),
		seq(63), seq(64), seq(65),
		seq(127), seq(128), seq(129),
		zeros(256), ones(256), seq(256),
		seq(1000), seq(8192),
	}
	for _, in := range inputs {
		h := sha256.Sum256(in)
		emit("sha256", toHex(in), toHex(h[:]))
	}
}

func runEd25519Verify() {
	type tc struct{ pub, msg, sig string }
	cases := []tc{
		{"3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29", "", "8f895b3cafe2c9506039d0e2a66382568004674fe8d237785092e40d6aaf483e4fc60168705f31f101596138ce21aa357c0d32a064f423dc3ee4aa3abf53f803"},
		{"3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29", "00", "1b9da904f12708363c88d4b96b33b474b2a5a863e290be2be5d4cacef8a5cbac1c0132c3e20e477d7affd1491be6577e8b83af012773beaf51c6df3f4af95b0e"},
		{"3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29", "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", "cd7641d4fdd9fecf9b34f7e8b04e909687da548d1150bf01db964da554463118b132030f01f726723214952260cc6ab3410332bfab755b12d0f786b7b2f14508"},
		{"76a1592044a6e4f511265bca73a604d90b0529d1df602be30a19a9257660d1f5", "", "ee00ac6e44a6a43bc3a701c137b467473e32a4bed47d3293fe6186fe9868c7d254024f65ffa552487bdbd34aaccd1e82c1a20a191d183dbb64c3523cccd89c0f"},
		{"76a1592044a6e4f511265bca73a604d90b0529d1df602be30a19a9257660d1f5", "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", "5c1a933f54393586a71449c8117f053c81b631c0589537718e3c804ef6e314fbe0aee160e4ce46b711b05bfbdae55eb6537b1bc0f484b3cd6fda7efddcc04700"},
		{"03a107bff3ce10be1d70dd18e74bc09967e4d6309ba50d5f1ddc8664125531b8", "00", "d45bd436eca8cf5deec21c0da72226f03aa29cabccce352c6740fba7c61267c8057328c52ef100e5bc02356fe2d4c53a1424116ce6a1cf77b4480fe3a905f90e"},
		// invalid (all-zero signature)
		{"3b6a27bcceb6a42d62a3a8d02a6f0d73653215771de243a63ac048a18b59da29", "00", "00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"},
	}
	for _, c := range cases {
		ok := ed25519.Verify(fromHex(c.pub), fromHex(c.msg), fromHex(c.sig))
		emit("ed25519-verify", c.pub, c.msg, c.sig, strconv.FormatBool(ok))
	}
}

func runChaCha20() {
	// 32-byte key, 12-byte nonce (RFC 7539). Output is keystream xor input.
	key := seq(32)
	nonces := [][]byte{seq(12), zeros(12), ones(12)}
	srcs := [][]byte{nil, zeros(1), zeros(63), zeros(64), zeros(65), zeros(1024)}
	for _, n := range nonces {
		for _, src := range srcs {
			dst := make([]byte, len(src))
			chacha20.XORKeyStream(dst, src, n, key)
			emit("chacha20", toHex(key), toHex(n), toHex(src), toHex(dst))
		}
	}
}

func runBech32() {
	type tc struct {
		hrp string
		dat []byte
	}
	cases := []tc{
		{"gno", []byte{}},
		{"gno", seq(20)},
		{"gno", seq(32)},
		{"abc", seq(64)},
		{"x", ones(40)},
	}
	for _, c := range cases {
		// Bech32 data must be 5-bit groups; convert from 8-bit first.
		conv, err := bech32.ConvertBits(c.dat, 8, 5, true)
		if err != nil {
			emit("bech32-encode", c.hrp, toHex(c.dat), "convert-err:"+err.Error())
			continue
		}
		enc, err := bech32.Encode(c.hrp, conv)
		if err != nil {
			emit("bech32-encode", c.hrp, toHex(c.dat), "encode-err:"+err.Error())
			continue
		}
		emit("bech32-encode", c.hrp, toHex(c.dat), enc)
	}
}

func runSubtleXOR() {
	type tc struct{ x, y []byte }
	cases := []tc{
		{zeros(16), ones(16)},
		{seq(32), ones(32)},
		{seq(64), seq(63)}, // unequal lengths: result is min(len)
		{ones(128), zeros(128)},
	}
	for _, c := range cases {
		out := subtle.XORBytes(c.x, c.y)
		emit("subtle-xor", toHex(c.x), toHex(c.y), toHex(out))
	}
}

func runAdler32() {
	inputs := [][]byte{
		nil,
		seq(1), seq(31), seq(32), seq(33),
		seq(1024), seq(8192),
		zeros(64), ones(64),
	}
	for _, in := range inputs {
		sum := adler32.Checksum(in)
		emit("adler32", toHex(in), strconv.FormatUint(uint64(sum), 16))
	}
}

func runKeccak256() {
	inputs := [][]byte{
		nil,
		{0x00},
		seq(31), seq(32), seq(33),
		seq(63), seq(64), seq(65),
		seq(127), seq(128), seq(129),
		zeros(256), ones(256), seq(256),
		seq(1000), seq(8192),
	}
	for _, in := range inputs {
		h := keccak256.Sum256(in)
		emit("keccak256", toHex(in), toHex(h[:]))
	}
}

func runModExp() {
	type tc struct{ base, exp, mod string }
	cases := []tc{
		// 2^10 mod 1000 = 1024 mod 1000 = 24
		{"02", "0a", "03e8"},
		// 3^5 mod 7 = 243 mod 7 = 5
		{"03", "05", "07"},
		// Identity: anything ^ 0 mod m = 1
		{"deadbeef", "00", "ff"},
		// Larger: 5^17 mod 2^32
		{"05", "11", "0100000000"},
		// Modulus 1 → always 0
		{"deadbeefcafebabe", "01020304", "01"},
	}
	for _, c := range cases {
		out := modexp.ModExp(fromHex(c.base), fromHex(c.exp), fromHex(c.mod))
		emit("modexp", c.base, c.exp, c.mod, toHex(out))
	}
}

func runMerkle() {
	// LeafHash on a few inputs.
	leafInputs := [][]byte{nil, {0x00}, seq(32), seq(128)}
	for _, in := range leafInputs {
		out := merkle.LeafHash(in)
		emit("merkle-leaf", toHex(in), toHex(out))
	}
	// InnerHash on a couple of (left, right) pairs.
	pairs := [][2][]byte{
		{zeros(32), ones(32)},
		{seq(32), seq(32)},
	}
	for _, p := range pairs {
		out := merkle.InnerHash(p[0], p[1])
		emit("merkle-inner", toHex(p[0]), toHex(p[1]), toHex(out))
	}
}

func runBN254() {
	// G1Add — verbatim vectors from EIP-196 bn256Add test set.
	g1AddCases := []string{
		// cdetrio11: G + G = 2G
		"0000000000000000000000000000000000000000000000000000000000000001000000000000000000000000000000000000000000000000000000000000000200000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002",
		// chfast1: non-trivial point addition
		"18b18acfb4c2c30276db5411368e7185b311dd124691610c5d3b74034e093dc9063c909c4720840cb5134cb9f59fa749755796819658d32efc0d288198f3726607c2b7f58a84bd6145f00c9c2bc0bb1a187f20ff2c92963a88019e7c6a014eed06614e20c147e940f2d70da3f74c9a17df361706a4485c742bd6788478fa17d7",
	}
	for _, in := range g1AddCases {
		out := bn254.G1Add(fromHex(in))
		emit("bn254-g1add", in, toHex(out))
	}
	// G1Mul — G * 2 (scalar = 2).
	g1MulIn := "00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002" +
		"0000000000000000000000000000000000000000000000000000000000000002"
	emit("bn254-g1mul", g1MulIn, toHex(bn254.G1Mul(fromHex(g1MulIn))))
	// PairingCheck — empty input is the identity, expected true (32 bytes, value 1).
	pcOut := bn254.PairingCheck(nil)
	emit("bn254-pairing", "", toHex(pcOut))
}

func main() {
	// Calls are ordered; each runner emits cases in a fixed order. The
	// stable per-line format makes diffing across architectures trivial.
	runSHA256()
	runKeccak256()
	runEd25519Verify()
	runChaCha20()
	runBech32()
	runSubtleXOR()
	runAdler32()
	runModExp()
	runMerkle()
	runBN254()
}
