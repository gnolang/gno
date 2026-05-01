package genproto2

import (
	"bytes"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
)

// Tests for the amino reserved-field feature, exercised end-to-end through
// the genproto2 generator. The feature has two contracts:
//
//   1. Field declared as `_ struct{} `amino:"reserved"`` reserves its fnum.
//   2. The generator emits a per-typ3 skip stub at that fnum so old wire
//      bytes carrying the formerly-occupied slot are consumed and discarded
//      rather than misparsed or rejected.
//
// The tests below form a truth table over (declaration present?, generator
// emission correct?), all sharing one V2-fixture pair declared in tests/:
//
//   tests.FixtureV2Reserved   — A=1, _=2(reserved), C=3   (correct migration)
//   tests.FixtureV2Shifted    — A=1, C=2                  (bad migration: silent shift)
//
// Both correspond to the same V1 ancestor: A int32 / B int32 / C string at
// fnums 1/2/3. V1 wire bytes are produced by encodeFixtureV1Bytes (we don't
// declare a Go type for V1 since nothing decodes into it).
//
// | Test                                        | Reserved declared? | Generator emits skip? | Decoder used                       | Decode V1 bytes |
// |---------------------------------------------|--------------------|------------------------|------------------------------------|-----------------|
// | TestPos_ReservedDeclared_DecodesOldBytes    | yes                | yes                    | real, generated (pb3_gen.go)       | succeeds        |
// | TestNeg1_NoReservedDeclared_FieldShifts     | no                 | n/a                    | real, generated (pb3_gen.go)       | typ3 mismatch   |
// | TestNeg2_ReservedDeclared_GeneratorRegressed| yes                | no (simulated)         | hand-written buggy mimic           | unknown fnum    |
//
// The positive and Neg #1 tests call the live generated UnmarshalBinary2
// methods directly, so they can never silently drift from the generator's
// emission shape. Neg #2 uses a hand-written decoder because the
// regressed-generator failure mode has no production analogue — implementing
// it via the real generator would either pollute gen_unmarshal.go with a
// test-only flag or stand up a `go run` subprocess per test (heavy,
// CI-fragile). Instead, TestBuggyMimic_MatchesRealGenMinusCase2 keeps the
// buggy mimic honest: it extracts the real case-2 reserved-skip block from
// pb3_gen.go at test time and asserts it equals the embedded snapshot,
// confirming the buggy mimic differs from the real generated code by exactly
// that block.

// encodeFixtureV1Bytes constructs wire bytes a non-regressed generator
// would emit for the V1 ancestor (A int32 / B int32 / C string at fnums
// 1/2/3) with the given values. Built directly with amino's primitive
// encoders to keep this test independent of codec dispatch and avoid
// pulling V1 into any package-level type registry.
func encodeFixtureV1Bytes(t *testing.T, a, b int32, c string) []byte {
	t.Helper()
	var buf bytes.Buffer
	// fnum 1 (A), Typ3Varint
	if err := amino.EncodeFieldNumberAndTyp3(&buf, 1, amino.Typ3Varint); err != nil {
		t.Fatalf("encode field 1 key: %v", err)
	}
	if err := amino.EncodeVarint(&buf, int64(a)); err != nil {
		t.Fatalf("encode A: %v", err)
	}
	// fnum 2 (B), Typ3Varint
	if err := amino.EncodeFieldNumberAndTyp3(&buf, 2, amino.Typ3Varint); err != nil {
		t.Fatalf("encode field 2 key: %v", err)
	}
	if err := amino.EncodeVarint(&buf, int64(b)); err != nil {
		t.Fatalf("encode B: %v", err)
	}
	// fnum 3 (C), Typ3ByteLength
	if err := amino.EncodeFieldNumberAndTyp3(&buf, 3, amino.Typ3ByteLength); err != nil {
		t.Fatalf("encode field 3 key: %v", err)
	}
	if err := amino.EncodeString(&buf, c); err != nil {
		t.Fatalf("encode C: %v", err)
	}
	return buf.Bytes()
}

// decodeFixtureV2ReservedBuggyGen mimics what gen_unmarshal.go would emit for
// tests.FixtureV2Reserved if the `for _, rnum := range info.Reserved` loop in
// writeStructUnmarshalBody were removed. The ONLY difference from the live
// generated tests.FixtureV2Reserved.UnmarshalBinary2 is the missing case-2
// skip stub: case 1 and case 3 are present, case 2 falls through to default.
//
// To verify "the ONLY difference" claim, see TestBuggyMimic_MatchesRealGenMinusCase2:
// it extracts the real case-2 block from pb3_gen.go at test time and
// asserts it equals reservedCase2Snapshot, the verbatim block this mimic
// elides.
func decodeFixtureV2ReservedBuggyGen(goo *tests.FixtureV2Reserved, bz []byte) error {
	*goo = tests.FixtureV2Reserved{}
	var lastFieldNum uint32
	for len(bz) > 0 {
		fnum, typ3, n, err := amino.DecodeFieldNumberAndTyp3(bz)
		if err != nil {
			return err
		}
		if fnum <= lastFieldNum {
			return fmt.Errorf("encountered fieldNum: %v, but we have already seen fnum: %v", fnum, lastFieldNum)
		}
		lastFieldNum = fnum
		bz = bz[n:]
		switch fnum {
		case 1:
			if typ3 != amino.Typ3Varint {
				return fmt.Errorf("field 1: expected typ3 %v, got %v", amino.Typ3Varint, typ3)
			}
			v, n, err := amino.DecodeVarint(bz)
			if err != nil {
				return err
			}
			bz = bz[n:]
			goo.A = int32(v)
		// BUG SIMULATION: case 2 (the reserved-skip stub the generator should
		// have emitted) is intentionally absent. See reservedCase2Snapshot
		// below for the verbatim block that's missing here.
		case 3:
			if typ3 != amino.Typ3ByteLength {
				return fmt.Errorf("field 3: expected typ3 %v, got %v", amino.Typ3ByteLength, typ3)
			}
			v, n, err := amino.DecodeString(bz)
			if err != nil {
				return err
			}
			bz = bz[n:]
			goo.C = v
		default:
			return fmt.Errorf("unknown field number %d for FixtureV2Reserved", fnum)
		}
	}
	return nil
}

// reservedCase2Snapshot is the verbatim case-2 reserved skip stub that
// gen_unmarshal.go's reserved-emission loop (lines ~322-344) writes into
// the generated FixtureV2Reserved.UnmarshalBinary2 in tests/pb3_gen.go.
// decodeFixtureV2ReservedBuggyGen above elides this block (and only this
// block) to simulate the regressed-generator failure mode.
//
// TestBuggyMimic_MatchesRealGenMinusCase2 keeps this snapshot honest by
// re-extracting the case-2 block from pb3_gen.go at test time and comparing.
const reservedCase2Snapshot = `		case 2:
			switch typ3 {
			case amino.Typ3Varint:
				_, n, err := amino.DecodeVarint(bz)
				if err != nil {
					return err
				}
				bz = bz[n:]
			case amino.Typ38Byte:
				_, n, err := amino.DecodeInt64(bz)
				if err != nil {
					return err
				}
				bz = bz[n:]
			case amino.Typ3ByteLength:
				_, n, err := amino.DecodeByteSlice(bz)
				if err != nil {
					return err
				}
				bz = bz[n:]
			case amino.Typ34Byte:
				_, n, err := amino.DecodeInt32(bz)
				if err != nil {
					return err
				}
				bz = bz[n:]
			default:
				return fmt.Errorf("invalid typ3 %v for reserved field 2", typ3)
			}
`

// TestPos_ReservedDeclared_DecodesOldBytes is the positive bookend to the
// two negative tests below. With the reserved declaration in place AND the
// generator emitting the case-2 skip stub correctly, V1 wire bytes (which
// carry a Typ3Varint payload at fnum 2 — formerly B's int32) decode cleanly:
// the skip stub consumes B's payload and the loop continues to decode C at
// fnum 3.
//
// V1 layout:        A=1  B=2(int32)         C=3(string)
// V2Reserved:       A=1  _=2(reserved skip) C=3(string)
// Decode result:    A populated, B's bytes discarded, C populated.
//
// Calls the real generated UnmarshalBinary2 from pb3_gen.go.
func TestPos_ReservedDeclared_DecodesOldBytes(t *testing.T) {
	v1Bytes := encodeFixtureV1Bytes(t, 1, 99, "foo")

	cdc := amino.NewCodec()
	var v2 tests.FixtureV2Reserved
	if err := v2.UnmarshalBinary2(cdc, v1Bytes, 0); err != nil {
		t.Fatalf("decode of V1 bytes failed: %v\n"+
			"Expected the generated case-2 reserved skip stub to consume B's payload "+
			"(fnum 2, Typ3Varint=99) and continue past it to decode C at fnum 3. "+
			"If this fails, the generator's reserved-emission loop "+
			"(gen_unmarshal.go ~lines 322-344) has regressed.", err)
	}
	if v2.A != 1 {
		t.Errorf("A: got %d, want 1 (fnum 1 must decode before reserved fnum 2 is skipped)", v2.A)
	}
	if v2.C != "foo" {
		t.Errorf("C: got %q, want \"foo\" (fnum 3 must decode after reserved fnum 2 is skipped)", v2.C)
	}
}

// TestNeg1_NoReservedDeclared_FieldShifts proves the "silent shift"
// failure mode: deleting a field without an `amino:"reserved"` placeholder
// causes subsequent fields to slide to lower fnums. Old-encoder bytes
// carrying the now-deleted field's typ3 hit the shifted slot and fail
// with a typ3 mismatch.
//
// V1 layout:    A=1  B=2(int32)  C=3(string)
// V2Shifted:    A=1  C=2(string)             ← C silently moved
// V1 bytes at fnum 2 carry Varint (B was int32). V2Shifted's case 2
// expects ByteLength (C is string). Decode fails at the case-2 typ3 guard.
//
// Calls the real generated UnmarshalBinary2 from pb3_gen.go.
func TestNeg1_NoReservedDeclared_FieldShifts(t *testing.T) {
	v1Bytes := encodeFixtureV1Bytes(t, 1, 99, "foo")

	cdc := amino.NewCodec()
	var v2 tests.FixtureV2Shifted
	err := v2.UnmarshalBinary2(cdc, v1Bytes, 0)
	if err == nil {
		t.Fatalf("expected decode to fail with typ3 mismatch at field 2 (V1 wrote Varint, V2Shifted expects ByteLength); "+
			"got success with v2=%+v.\n"+
			"This means the silent-shift failure mode is not being triggered by the test fixtures, "+
			"which would invalidate the regression guard.", v2)
	}
	// Assert on the substring the generator emits for typ3 mismatches:
	// `field <N>: expected typ3 ...`. Not pinning the exact typ3 stringer
	// output keeps this resilient to harmless cosmetic changes.
	if !strings.Contains(err.Error(), "field 2: expected typ3") {
		t.Fatalf("decode failed for the wrong reason: %v\n"+
			"Expected a typ3 mismatch at field 2 (V1 wrote Varint, V2Shifted expects ByteLength). "+
			"Got an unrelated error, suggesting the generator's typ3-mismatch error format has changed.", err)
	}
}

// TestNeg2_ReservedDeclared_GeneratorRegressed proves the "buggy generator"
// failure mode: even when the source code correctly declares
// `_ struct{} `amino:"reserved"“, if the generator regresses and stops
// emitting the case-N skip stub, old-encoder bytes carrying the reserved
// fnum hit `default:` and fail with "unknown field number N for <Type>".
//
// V2Reserved layout:        A=1  _=2(reserved)  C=3
// Buggy gen output:         case 1, case 3, default       ← case 2 missing
// V1 bytes have fnum 2 (B's Varint payload). Without case 2, this falls
// through to default → "unknown field number 2 for FixtureV2Reserved".
//
// Calls the hand-written decodeFixtureV2ReservedBuggyGen. The buggy mimic
// stays faithful to "real generated code minus the case-2 block" because
// TestBuggyMimic_MatchesRealGenMinusCase2 verifies the case-2 snapshot
// matches what the live generator emits.
func TestNeg2_ReservedDeclared_GeneratorRegressed(t *testing.T) {
	v1Bytes := encodeFixtureV1Bytes(t, 1, 99, "foo")

	var v2 tests.FixtureV2Reserved
	err := decodeFixtureV2ReservedBuggyGen(&v2, v1Bytes)
	if err == nil {
		t.Fatalf("expected decode to fail with 'unknown field number 2' (buggy generator omitted reserved skip stub); "+
			"got success with v2=%+v.\n"+
			"This means the buggy-generator failure mode is not being triggered, "+
			"which would invalidate the regression guard.", v2)
	}
	const wantSubstr = "unknown field number 2 for FixtureV2Reserved"
	if !strings.Contains(err.Error(), wantSubstr) {
		t.Fatalf("decode failed for the wrong reason: %v\n"+
			"Expected error containing %q. Got an unrelated error, suggesting "+
			"the hand-written buggy decoder no longer mirrors the live generator's "+
			"emission shape (default-arm message in gen_unmarshal.go ~line 347).",
			err, wantSubstr)
	}
}

// TestBuggyMimic_MatchesRealGenMinusCase2 keeps decodeFixtureV2ReservedBuggyGen
// honest: extract the case-2 reserved-skip block from the live generated
// FixtureV2Reserved.UnmarshalBinary2 in tests/pb3_gen.go and assert it equals
// reservedCase2Snapshot. If the generator changes its reserved-emission
// shape, this test fails loudly and points the maintainer at the snapshot
// to update — guaranteeing the buggy mimic stays exactly "real gen minus
// the case-2 block" rather than silently drifting away.
func TestBuggyMimic_MatchesRealGenMinusCase2(t *testing.T) {
	const pbgenPath = "../tests/pb3_gen.go"

	src, err := os.ReadFile(pbgenPath)
	if err != nil {
		t.Fatalf("read %s: %v", pbgenPath, err)
	}

	// Locate FixtureV2Reserved.UnmarshalBinary2 — the function whose case-2
	// block the buggy mimic elides.
	funcStart := bytes.Index(src, []byte("func (goo *FixtureV2Reserved) UnmarshalBinary2"))
	if funcStart < 0 {
		t.Fatalf("could not locate FixtureV2Reserved.UnmarshalBinary2 in %s; "+
			"either the type was renamed or pb3_gen.go is stale (run `make -C misc/genproto2`)", pbgenPath)
	}
	funcSrc := src[funcStart:]

	// Find the case-2 block. Match from `\t\tcase 2:\n` lazily up to the next
	// `\t\tcase ` or `\t\tdefault:` at the same indent. The trailing line is
	// not part of the block — the lookahead anchors the boundary.
	re := regexp.MustCompile(`(?s)(\t\tcase 2:\n.*?\n)(\t\t(?:case |default:))`)
	m := re.FindSubmatch(funcSrc)
	if m == nil {
		t.Fatalf("could not extract case-2 block from FixtureV2Reserved.UnmarshalBinary2 in %s; "+
			"either the generator changed indentation/layout or the reserved-emission loop was removed", pbgenPath)
	}
	got := string(m[1])

	if got != reservedCase2Snapshot {
		t.Fatalf("case-2 reserved skip stub in tests/pb3_gen.go has drifted from reservedCase2Snapshot.\n"+
			"This means gen_unmarshal.go's reserved-emission loop (~lines 322-344) emits a different shape now.\n"+
			"Update reservedCase2Snapshot to match the new emission, then verify\n"+
			"decodeFixtureV2ReservedBuggyGen still represents 'real gen minus the case-2 block'.\n\n"+
			"=== got (real generated) ===\n%s\n"+
			"=== want (snapshot) ===\n%s",
			got, reservedCase2Snapshot)
	}
}
