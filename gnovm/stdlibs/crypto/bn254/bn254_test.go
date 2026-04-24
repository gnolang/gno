package bn254

import (
	"bytes"
	"encoding/hex"
	"math/big"
	"testing"

	bn254 "github.com/consensys/gnark-crypto/ecc/bn254"
)

// g1Generator returns the EIP-196 canonical G1 generator as a 64-byte (x|y)
// buffer: x=1, y=2.
func g1Generator() []byte {
	out := make([]byte, 64)
	out[31] = 1
	out[63] = 2
	return out
}

// pack32 concatenates its arguments, each of which must already be 32 bytes.
func pack32(t *testing.T, parts ...[]byte) []byte {
	t.Helper()
	out := make([]byte, 0, 32*len(parts))
	for i, p := range parts {
		if len(p) != 32 {
			t.Fatalf("pack32: part %d has length %d, want 32", i, len(p))
		}
		out = append(out, p...)
	}
	return out
}

func fpModulusBytes() []byte {
	out := make([]byte, 32)
	fpModulus.FillBytes(out)
	return out
}

func TestG1AddGeneratorDouble(t *testing.T) {
	// (1, 2) + (1, 2) should equal 2G. Compute 2G via gnark-crypto so the
	// test is validated against an independent code path (the gnark-crypto
	// generator routine rather than our parse/marshal).
	var g, twoG bn254.G1Affine
	g.X.SetUint64(1)
	g.Y.SetUint64(2)
	twoG.Double(&g)

	input := make([]byte, 128)
	copy(input[0:64], g1Generator())
	copy(input[64:128], g1Generator())

	got := X_g1Add(input)
	if got == nil {
		t.Fatalf("X_g1Add returned nil for a valid addition")
	}
	if len(got) != 64 {
		t.Fatalf("expected 64-byte output, got %d bytes", len(got))
	}

	wantX := twoG.X.Bytes()
	wantY := twoG.Y.Bytes()
	if !bytes.Equal(got[0:32], wantX[:]) || !bytes.Equal(got[32:64], wantY[:]) {
		t.Fatalf("mismatch:\n got %x\n want %x%x", got, wantX, wantY)
	}
}

func TestG1AddWithIdentity(t *testing.T) {
	// (1, 2) + (0, 0) = (1, 2). The all-zero encoding is the point at
	// infinity in EIP-196.
	input := make([]byte, 128)
	copy(input[0:64], g1Generator())
	// second half stays zero.

	got := X_g1Add(input)
	if !bytes.Equal(got, g1Generator()) {
		t.Fatalf("P + 0 != P: got %x", got)
	}
}

func TestG1AddInvalidInputs(t *testing.T) {
	// Wrong length.
	if X_g1Add(make([]byte, 127)) != nil {
		t.Fatalf("expected nil for wrong-length input")
	}
	if X_g1Add(nil) != nil {
		t.Fatalf("expected nil for nil input")
	}

	// Coordinate exactly equal to the field modulus must be rejected.
	p := fpModulusBytes()
	input := make([]byte, 128)
	copy(input[0:32], p) // x1 = p, not reduced
	input[63] = 2        // y1 = 2, valid
	copy(input[64:128], g1Generator())
	if X_g1Add(input) != nil {
		t.Fatalf("expected nil for unreduced coordinate")
	}

	// Point not on curve: (1, 3) doesn't satisfy y^2 = x^3 + 3.
	input = make([]byte, 128)
	input[31] = 1
	input[63] = 3
	copy(input[64:128], g1Generator())
	if X_g1Add(input) != nil {
		t.Fatalf("expected nil for off-curve point")
	}
}

func TestG1MulKnownScalars(t *testing.T) {
	// 0 * G = identity (0, 0).
	input := make([]byte, 96)
	copy(input[0:64], g1Generator())
	// scalar (last 32 bytes) stays zero.
	got := X_g1Mul(input)
	want := make([]byte, 64)
	if !bytes.Equal(got, want) {
		t.Fatalf("0*G != 0: got %x", got)
	}

	// 1 * G = G.
	input = make([]byte, 96)
	copy(input[0:64], g1Generator())
	input[95] = 1
	got = X_g1Mul(input)
	if !bytes.Equal(got, g1Generator()) {
		t.Fatalf("1*G != G: got %x", got)
	}

	// 2 * G must match the G1Add result (lets us cross-validate without
	// hardcoded hex coordinates).
	addInput := make([]byte, 128)
	copy(addInput[0:64], g1Generator())
	copy(addInput[64:128], g1Generator())
	wantDouble := X_g1Add(addInput)

	input = make([]byte, 96)
	copy(input[0:64], g1Generator())
	input[95] = 2
	gotDouble := X_g1Mul(input)
	if !bytes.Equal(gotDouble, wantDouble) {
		t.Fatalf("2*G via mul != G+G via add:\n mul: %x\n add: %x", gotDouble, wantDouble)
	}
}

func TestG1MulInvalidInputs(t *testing.T) {
	if X_g1Mul(make([]byte, 95)) != nil {
		t.Fatalf("expected nil for wrong-length input")
	}
	if X_g1Mul(nil) != nil {
		t.Fatalf("expected nil for nil input")
	}

	// Off-curve point with any scalar must be rejected.
	input := make([]byte, 96)
	input[31] = 1
	input[63] = 3
	input[95] = 5
	if X_g1Mul(input) != nil {
		t.Fatalf("expected nil for off-curve point")
	}
}

func TestPairingCheckEmptyInputIsIdentity(t *testing.T) {
	// EIP-197: the product of zero pairings is 1 in GT. The precompile
	// returns a 32-byte 0x...01.
	got := X_pairingCheck(nil)
	if len(got) != 32 {
		t.Fatalf("expected 32-byte output, got %d", len(got))
	}
	want := make([]byte, 32)
	want[31] = 1
	if !bytes.Equal(got, want) {
		t.Fatalf("empty input should report 1: got %x", got)
	}
}

// TestPairingCheckOppositePairs verifies e(P, Q) * e(-P, Q) = 1 by building a
// 2-pair input. The first pair uses G1's generator and a generic G2 element
// produced by gnark-crypto; the second pair negates the G1 coordinate.
func TestPairingCheckOppositePairs(t *testing.T) {
	_, _, _, g2Gen := bn254.Generators()

	// G1 generator and its negation in EIP-196 encoding (x|y, 32 BE each).
	var pMinus bn254.G1Affine
	pMinus.X.SetUint64(1)
	pMinus.Y.SetUint64(2)
	pMinus.Y.Neg(&pMinus.Y)
	pMinusX := pMinus.X.Bytes()
	pMinusY := pMinus.Y.Bytes()

	// G2 generator in EIP-197 encoding (x_imag|x_real|y_imag|y_real).
	g2Marshal := g2Gen.Marshal() // 128 bytes, imag-first.

	input := make([]byte, 0, 2*192)
	// Pair 1: e(G1, G2Gen)
	input = append(input, g1Generator()...)
	input = append(input, g2Marshal...)
	// Pair 2: e(-G1, G2Gen)
	input = append(input, pMinusX[:]...)
	input = append(input, pMinusY[:]...)
	input = append(input, g2Marshal...)

	got := X_pairingCheck(input)
	want := make([]byte, 32)
	want[31] = 1
	if !bytes.Equal(got, want) {
		t.Fatalf("e(P, Q) * e(-P, Q) should equal 1: got %x", got)
	}
}

func TestPairingCheckMismatchIsZero(t *testing.T) {
	// e(G1, G2) ≠ 1 for non-degenerate generators. A single pair must
	// therefore NOT pass the check — the precompile returns 0x...00.
	_, _, _, g2Gen := bn254.Generators()
	g2Marshal := g2Gen.Marshal()

	input := make([]byte, 0, 192)
	input = append(input, g1Generator()...)
	input = append(input, g2Marshal...)

	got := X_pairingCheck(input)
	want := make([]byte, 32) // all zeros.
	if !bytes.Equal(got, want) {
		t.Fatalf("single pair should report 0: got %x", got)
	}
}

func TestPairingCheckInvalidLengths(t *testing.T) {
	// Not a multiple of 192.
	if X_pairingCheck(make([]byte, 191)) != nil {
		t.Fatalf("expected nil for non-192-multiple input")
	}
	if X_pairingCheck(make([]byte, 193)) != nil {
		t.Fatalf("expected nil for non-192-multiple input")
	}
}

func TestPairingCheckRejectsUnreducedCoordinates(t *testing.T) {
	_, _, _, g2Gen := bn254.Generators()
	g2Marshal := g2Gen.Marshal()

	// Overwrite G1.x with the field modulus to force rejection.
	p := fpModulusBytes()
	input := make([]byte, 0, 192)
	input = append(input, p...)                                   // x = modulus
	input = append(input, []byte{0, 0, 0, 0, 0, 0, 0, 0,          // y = 2 (padded)
		0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 2}...)
	input = append(input, g2Marshal...)

	if X_pairingCheck(input) != nil {
		t.Fatalf("expected nil for unreduced G1.x coordinate")
	}
}

// TestG1AddRejectsOutOfBoundsScalar is a negative companion to the add tests:
// coordinate equal to p-1 is valid, coordinate equal to p is not.
func TestG1AddBoundaryCoordinate(t *testing.T) {
	// pMinusOne as 32-byte BE.
	pMinusOne := new(big.Int).Sub(fpModulus, big.NewInt(1))
	buf := make([]byte, 32)
	pMinusOne.FillBytes(buf)

	// Construct a random valid point ((0, 0) = infinity) + something with
	// x = p-1, y = 0. (p-1, 0) is NOT on y^2 = x^3 + 3 for any interesting
	// curve, so we expect rejection.
	input := make([]byte, 128)
	copy(input[0:32], buf)
	// y1 = 0.
	// Second half: identity, valid.

	// The coordinate is reduced (< p), but the point is off-curve, so we
	// should still get nil.
	if X_g1Add(input) != nil {
		t.Fatalf("expected nil for off-curve point with reduced coordinate")
	}
	_ = hex.EncodeToString(buf)
}
