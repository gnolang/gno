package calibrate

// Calibration benchmarks for crypto/secp256k1.recoverPubkey, through the
// native dispatcher like every other row (conversion of the two arrays, the
// recovery, and the [65]byte and bool results).
//
// The row is flat: RecoverPubkey hands the native two fixed-size arrays, so the
// conversion is the same for every call, and only the scalars change what a
// recovery costs. It is derived rather than fitted (see the row in
// gnovm/stdlibs/native_gas.go): _PerVerify times each valid input against
// crypto/secp256k1.verify in pairs, which holds on a loaded machine, and the
// row is the slowest input's ratio times the verify row at 64 bytes, times
// 1.2. secp256k1_recover_bench_m1pro.txt is the run behind it.
//
// The single benches show the spread. _Slow is a signature whose scalar
// u2 = s/r has a dense signed-digit form, so the variable-base multiplication
// does close to its most point additions; _Random holds 64 ordinary
// signatures. The _Refuse benches show what refusals cost: little for a bad
// recovery ID or scalar, a square root for an r with no curve point, and both
// multiplications for a signature that recovers the point at infinity.

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"slices"
	"testing"
	"time"

	dcrsecp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	dcrecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

// Signer 0 of the recovery corpus in gnovm/stdlibs/crypto/secp256k1/testdata,
// its high-S twin, the same r and digest with a dense s, and refusal inputs.
const (
	secpRecoverDigest = "4c358ead6f3ea252278a6c0e69069eecec31c517786eba631a79352dc8f61c56"
	secpRecoverR      = "a1842bb853abc7b49fe40f5af46b0f75aa6c962219e689f5a7ccea2b81d4f4e8"
	secpRecoverS      = "1091ef7e7b447cac51d460e19a242e5237810ddc1373a925bfd83f77263d0c2f"
	secpRecoverHighS  = "ef6e108184bb8353ae2b9f1e65dbd1ac832dcf0a9bd4f715fffa1f15a9f93512"
	secpRecoverSlowS  = "472aebbbdd5aa013207fa01a3a071b7b5890278fd3c1f76823ad13ba4f75fac8"
	secpOrder         = "fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141"
	secpOne           = "0000000000000000000000000000000000000000000000000000000000000001"
	secpGx            = "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"
)

func secpRecoverInput(b *testing.B, digestHex, r, s string, id byte) (digest [32]byte, sig [65]byte) {
	b.Helper()
	d, err := hex.DecodeString(digestHex)
	if err != nil {
		b.Fatal(err)
	}
	rs, err := hex.DecodeString(r + s)
	if err != nil {
		b.Fatal(err)
	}
	copy(sig[:64], rs)
	sig[64] = id
	return [32]byte(d), sig
}

func benchSecp256k1RecoverPubkey(b *testing.B, digest [32]byte, sig [65]byte) {
	b.Helper()
	m := newDispatchMachine(2)
	setBlockValueFromGo(m, 0, digest)
	setBlockValueFromGo(m, 1, sig)
	h := &dispatchHarness{m: m, wrapper: resolveWrapper(b, "crypto/secp256k1", "recoverPubkey"), nReturns: 2}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.call()
	}
}

func BenchmarkNative_Secp256k1_RecoverPubkey(b *testing.B) {
	digest, sig := secpRecoverInput(b, secpRecoverDigest, secpRecoverR, secpRecoverS, 0)
	benchSecp256k1RecoverPubkey(b, digest, sig)
}

func BenchmarkNative_Secp256k1_RecoverPubkey_HighS(b *testing.B) {
	digest, sig := secpRecoverInput(b, secpRecoverDigest, secpRecoverR, secpRecoverHighS, 1)
	benchSecp256k1RecoverPubkey(b, digest, sig)
}

func BenchmarkNative_Secp256k1_RecoverPubkey_Slow(b *testing.B) {
	digest, sig := secpRecoverInput(b, secpRecoverDigest, secpRecoverR, secpRecoverSlowS, 0)
	benchSecp256k1RecoverPubkey(b, digest, sig)
}

// secpRecoverRandomInputs returns 64 signatures by deterministic keys over
// deterministic digests, with RFC 6979 nonces.
func secpRecoverRandomInputs() (digests [][32]byte, sigs [][65]byte) {
	for i := range 64 {
		seed := sha256.Sum256(fmt.Appendf(nil, "recoverPubkey bench key %d", i))
		digest := sha256.Sum256(fmt.Appendf(nil, "recoverPubkey bench digest %d", i))
		compact := dcrecdsa.SignCompact(dcrsecp.PrivKeyFromBytes(seed[:]), digest[:], false)
		var sig [65]byte
		copy(sig[:64], compact[1:]) // compact is 27+id || r || s
		sig[64] = compact[0] - 27
		digests, sigs = append(digests, digest), append(sigs, sig)
	}
	return digests, sigs
}

func BenchmarkNative_Secp256k1_RecoverPubkey_Random(b *testing.B) {
	digests, sigs := secpRecoverRandomInputs()
	for i := range digests {
		b.Run(fmt.Sprintf("%02d", i), func(b *testing.B) { benchSecp256k1RecoverPubkey(b, digests[i], sigs[i]) })
	}
}

// BenchmarkNative_Secp256k1_RecoverPubkey_PerVerify reports how long each
// valid recovery (the 64 random inputs, signer 0, its high-S twin and the
// dense-s input) takes per crypto/secp256k1.verify of a 64-byte message:
// every round runs a batch of verifications, then a batch of recoveries of
// one input, for every input. Run it with -benchtime Nx: N rounds. On a loaded
// machine single timings swing several-fold, while the ratio of two batches
// run back to back holds; projected onto the verify row it prices a recovery
// on that row's machine. It reports the median ratio of the slowest input
// ("worst"), the median over inputs ("typical"), and the dense-s input's.
func BenchmarkNative_Secp256k1_RecoverPubkey_PerVerify(b *testing.B) {
	const batch = 8
	digests, sigs := secpRecoverRandomInputs()
	for _, c := range []struct {
		s  string
		id byte
	}{{secpRecoverS, 0}, {secpRecoverHighS, 1}, {secpRecoverSlowS, 0}} {
		d, sig := secpRecoverInput(b, secpRecoverDigest, secpRecoverR, c.s, c.id)
		digests, sigs = append(digests, d), append(sigs, sig)
	}
	recovers := make([]*dispatchHarness, len(digests))
	for i := range digests {
		m := newDispatchMachine(2)
		setBlockValueFromGo(m, 0, digests[i])
		setBlockValueFromGo(m, 1, sigs[i])
		recovers[i] = &dispatchHarness{m: m, wrapper: resolveWrapper(b, "crypto/secp256k1", "recoverPubkey"), nReturns: 2}
	}

	// A fixed key and message, so every run verifies the same signature.
	seed := sha256.Sum256([]byte("verify bench key"))
	priv := dcrsecp.PrivKeyFromBytes(seed[:])
	msg := make([]byte, 64)
	msgHash := sha256.Sum256(msg)
	compact := dcrecdsa.SignCompact(priv, msgHash[:], true)
	vm := newDispatchMachine(3)
	setBlockValueFromGo(vm, 0, priv.PubKey().SerializeCompressed())
	setBlockValueFromGo(vm, 1, msg)
	setBlockValueFromGo(vm, 2, compact[1:]) // r || s
	verify := &dispatchHarness{m: vm, wrapper: resolveWrapper(b, "crypto/secp256k1", "verify"), nReturns: 1}

	// A refusal returns early and would skew the ratio: check every input.
	if verify.wrapper(vm); !vm.PopValue().GetBool() {
		b.Fatal("verify refuses the bench signature")
	}
	for i, h := range recovers {
		if h.wrapper(h.m); !h.m.PopValue().GetBool() {
			b.Fatalf("recovery %d refused", i)
		}
		h.m.PopValue()
	}

	ratios := make([][]float64, len(recovers))
	b.ResetTimer()
	for range b.N {
		for i, h := range recovers {
			start := time.Now()
			for range batch {
				verify.call()
			}
			verifyTime := time.Since(start)
			start = time.Now()
			for range batch {
				h.call()
			}
			ratios[i] = append(ratios[i], float64(time.Since(start))/float64(verifyTime))
		}
	}
	b.StopTimer()

	medians := make([]float64, len(ratios))
	for i, r := range ratios {
		slices.Sort(r)
		medians[i] = r[len(r)/2]
	}
	denseS := medians[len(medians)-1] // the last input added above
	worst := slices.Max(medians)
	slices.Sort(medians)
	b.ReportMetric(worst, "worst")
	b.ReportMetric(medians[len(medians)/2], "typical")
	b.ReportMetric(denseS, "dense-s")
}

func BenchmarkNative_Secp256k1_RecoverPubkey_RefuseID(b *testing.B) {
	digest, sig := secpRecoverInput(b, secpRecoverDigest, secpRecoverR, secpRecoverS, 2)
	benchSecp256k1RecoverPubkey(b, digest, sig)
}

func BenchmarkNative_Secp256k1_RecoverPubkey_RefuseScalar(b *testing.B) {
	digest, sig := secpRecoverInput(b, secpRecoverDigest, secpRecoverR, secpOrder, 0)
	benchSecp256k1RecoverPubkey(b, digest, sig)
}

// r = 5 is not the x-coordinate of a curve point: the square root is computed,
// then the signature refused.
func BenchmarkNative_Secp256k1_RecoverPubkey_RefusePoint(b *testing.B) {
	digest, sig := secpRecoverInput(b, secpRecoverDigest, "0000000000000000000000000000000000000000000000000000000000000005", secpRecoverS, 0)
	benchSecp256k1RecoverPubkey(b, digest, sig)
}

// Digest 1, r = Gx, s = 1, ID 0: the recovered point is G - G, the point at
// infinity, found only after both multiplications.
func BenchmarkNative_Secp256k1_RecoverPubkey_RefuseInfinity(b *testing.B) {
	digest, sig := secpRecoverInput(b, secpOne, secpGx, secpOne, 0)
	benchSecp256k1RecoverPubkey(b, digest, sig)
}
