package amino_test

import (
	"fmt"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	proto "google.golang.org/protobuf/proto"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
)

// fuzzPerSubtest is the time each property-fuzz subtest should loop when
// AMINO_FUZZ_BUDGET is set. Computed as:
//
//	FUZZ_BUDGET × GOMAXPROCS / total_subtests
//
// With t.Parallel(), up to GOMAXPROCS subtests run concurrently. Each runs
// for fuzzPerSubtest, then returns — freeing its slot for the next batch.
// Total wallclock ≈ FUZZ_BUDGET, and every type gets equal coverage.
//
// When AMINO_FUZZ_BUDGET is unset (normal `go test`), this is 0 and the
// helpers fall back to a fast fixed iteration count (10k).
var fuzzPerSubtest = func() time.Duration {
	s := os.Getenv("AMINO_FUZZ_BUDGET")
	if s == "" {
		return 0
	}
	total, err := time.ParseDuration(s)
	if err != nil {
		return 0
	}
	if total <= 0 {
		// AMINO_FUZZ_BUDGET=0 means "run forever" (Ctrl-C to stop).
		// Use a 1h rotation period so all subtests cycle through.
		total = 1 * time.Hour
	}
	procs := runtime.GOMAXPROCS(0)
	// Total property-fuzz subtests across all three test functions:
	//   StructTypes×2 (binary+json) + AminoTagTypes×2 + DefTypes×2 + DeepCopy(Struct+AminoTag+Def)
	numSubtests := (len(tests.StructTypes)+len(tests.AminoTagTypes)+len(tests.DefTypes))*2 +
		len(tests.StructTypes) + len(tests.AminoTagTypes) + len(tests.DefTypes)
	if numSubtests == 0 {
		return total
	}
	return total * time.Duration(procs) / time.Duration(numSubtests)
}()

// shouldContinue returns true while the fuzz loop should keep iterating.
// In budgeted mode, each call gets its own deadline (now+fuzzPerSubtest);
// otherwise caps at iters.
func shouldContinue(i int, iters int, deadline time.Time) bool {
	if !deadline.IsZero() {
		return time.Now().Before(deadline)
	}
	return i < iters
}

// -------------------------------------
// Non-interface Google fuzz tests

func TestCodecStruct(t *testing.T) {
	t.Parallel()

	for _, ptr := range tests.StructTypes {
		t.Logf("case %v", reflect.TypeOf(ptr))
		rt := getTypeFromPointer(ptr)
		name := rt.Name()
		t.Run(name+":binary", func(t *testing.T) {
			t.Parallel()
			_testCodec(t, rt, "binary")
		})
		t.Run(name+":json", func(t *testing.T) {
			t.Parallel()
			_testCodec(t, rt, "json")
		})
	}
}

// TestCodecAminoTags tests types with amino-specific encoding tags
// (write_empty, nil_elements) that have no proto3 equivalent.
// Same checks as TestCodecStruct but skips proto.Marshal byte comparison.
//
// FuzzNilElements is structurally lossy: with amino:"nil_elements", both
// a nil *Struct and a non-nil zero-valued *Struct encode to 0x00 (proto3
// can't distinguish "nil" from "empty message"), and 0x00 always decodes
// to nil. So orig-containing-&Struct{} round-trips to decoded-containing-nil,
// and strict equality cannot hold. For that type we run byte-stability
// and cross-encoder/decoder checks instead, which still catch regressions.
func TestCodecAminoTags(t *testing.T) {
	t.Parallel()

	for _, ptr := range tests.AminoTagTypes {
		t.Logf("case %v", reflect.TypeOf(ptr))
		rt := getTypeFromPointer(ptr)
		name := rt.Name()
		lossyDecode := name == "FuzzNilElements"
		t.Run(name+":binary", func(t *testing.T) {
			t.Parallel()
			_testCodecAminoTags(t, rt, "binary", lossyDecode)
		})
		t.Run(name+":json", func(t *testing.T) {
			t.Parallel()
			_testCodecAminoTags(t, rt, "json", lossyDecode)
		})
	}
}

func TestCodecDef(t *testing.T) {
	t.Parallel()

	for _, ptr := range tests.DefTypes {
		t.Logf("case %v", reflect.TypeOf(ptr))
		rt := getTypeFromPointer(ptr)
		name := rt.Name()
		t.Run(name+":binary", func(t *testing.T) {
			t.Parallel()
			_testCodec(t, rt, "binary")
		})
		t.Run(name+":json", func(t *testing.T) {
			t.Parallel()
			_testCodec(t, rt, "json")
		})
	}
}

func TestDeepCopyStruct(t *testing.T) {
	t.Parallel()

	all := append([]any{}, tests.StructTypes...)
	all = append(all, tests.AminoTagTypes...)
	for _, ptr := range all {
		t.Logf("case %v", reflect.TypeOf(ptr))
		rt := getTypeFromPointer(ptr)
		name := rt.Name()
		t.Run(name+":deepcopy", func(t *testing.T) {
			t.Parallel()
			_testDeepCopy(t, rt)
		})
	}
}

func TestDeepCopyDef(t *testing.T) {
	t.Parallel()

	for _, ptr := range tests.DefTypes {
		t.Logf("case %v", reflect.TypeOf(ptr))
		rt := getTypeFromPointer(ptr)
		name := rt.Name()
		t.Run(name+":deepcopy", func(t *testing.T) {
			t.Parallel()
			_testDeepCopy(t, rt)
		})
	}
}

func _testCodec(t *testing.T, rt reflect.Type, codecType string) {
	t.Helper()

	err := error(nil)
	bz := []byte{}
	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)
	f := fuzz.New()
	rv := reflect.New(rt)
	rv2 := reflect.New(rt)
	ptr := rv.Interface()
	ptr2 := rv2.Interface()
	rnd := rand.New(rand.NewSource(10))
	f.RandSource(rnd)
	f.Funcs(fuzzFuncs...)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic'd:\nreason: %v\n%s\nerr: %v\nbz: %X\nrv: %#v\nrv2: %#v\nptr: %v\nptr2: %v\n",
				r, debug.Stack(), err, bz, rv, rv2, spw(ptr), spw(ptr2),
			)
		}
	}()

	var deadline time.Time
	if fuzzPerSubtest > 0 {
		deadline = time.Now().Add(fuzzPerSubtest)
	}
	for i := 0; shouldContinue(i, 10_000, deadline); i++ {
		f.Fuzz(ptr)

		// Reset, which makes debugging decoding easier.
		rv2 = reflect.New(rt)
		ptr2 = rv2.Interface()

		// Encode to bz. Use MarshalReflect explicitly so the encoder-parity
		// check against MarshalBinary2 below (bz vs bz2) is a real cross-codec
		// test rather than comparing genproto2 to itself via dispatch.
		switch codecType {
		case "binary":
			bz, err = cdc.MarshalReflect(ptr)
		case "json":
			bz, err = cdc.JSONMarshal(ptr)
		default:
			panic("should not happen")
		}
		require.Nil(t, err,
			"failed to marshal %v to bytes: %v\n",
			spw(ptr), err)

		// Decode from bz. Use UnmarshalReflect explicitly for the binary
		// roundtrip so genproto2-registered types don't dispatch away from
		// the reflect path — the genproto2 path is exercised separately
		// below, and we want a real reflect-vs-genproto2 cross-check.
		switch codecType {
		case "binary":
			err = cdc.UnmarshalReflect(bz, ptr2)
		case "json":
			err = cdc.JSONUnmarshal(bz, ptr2)
		default:
			panic("should not happen")
		}
		require.NoError(t, err,
			"failed to unmarshal bytes %X (%s): %v\nptr: %v\n",
			bz, bz, err, spw(ptr))
		require.Equal(t, ptr, ptr2,
			"end to end failed.\nstart: %v\nend: %v\nbytes: %X\nstring(bytes): %s\n",
			spw(ptr), spw(ptr2), bz, bz)

		if codecType == "binary" {
			// Get pbo from rv. (go -> p3go)
			pbm, ok := rv.Interface().(amino.PBMessager)
			if !ok {
				// typedefs that are not structs, for example,
				// are not pbMessanger.
				continue
			}
			pbo, err := pbm.ToPBMessage(cdc)
			require.NoError(t, err)

			// Get back to go from pbo, and ensure equality. (go -> p3go -> go vs go)
			rv3 := reflect.New(rt)
			ptr3 := rv3.Interface()
			err = ptr3.(amino.PBMessager).FromPBMessage(cdc, pbo)
			require.NoError(t, err)
			require.Equal(t, ptr, ptr3,
				"end to end through pbo failed.\nstart(goo): %v\nend(goo): %v\nmid(pbo): %v\n",
				spw(ptr), spw(ptr3), spw(pbo))

			// Marshal pbo and check for equality of bz and b3. (go -> p3go -> bz vs go -> bz)
			bz3, err := proto.Marshal(pbo)
			require.NoError(t, err)
			require.Equal(t, bz, bz3,
				"pbo serialization check failed.\nbz(go): %X\nbz(pb-go): %X\nstart(goo): %v\nend(pbo): %v\n",
				bz, bz3, spw(ptr), spw(pbo))

			// Decode from bz and check for equality (go -> bz -> p3go -> go vs go)
			pbo2 := pbm.EmptyPBMessage(cdc)
			err = proto.Unmarshal(bz, pbo2)
			require.NoError(t, err)
			rv4 := reflect.New(rt)
			ptr4 := rv4.Interface()
			err = ptr4.(amino.PBMessager).FromPBMessage(cdc, pbo2)
			require.NoError(t, err)
			require.Equal(t, ptr, ptr4,
				"end to end through bytes and pbo failed.\nbz(go): %X\nstart(goo): %v\nend(goo): %v\nmid(pbo): %v\n",
				bz, spw(ptr), spw(ptr3), spw(pbo))
		}

		if codecType == "binary" {
			// Check genproto2 (go -> bz2 vs go -> bz)
			pbm2, ok := rv.Interface().(amino.PBMessager2)
			if !ok {
				continue
			}
			bz2, err := cdc.MarshalBinary2(pbm2)
			require.NoError(t, err,
				"MarshalBinary2 failed for %v: %v\n", spw(ptr), err)
			require.Equal(t, bz, bz2,
				"genproto2 bytes mismatch.\nbz(amino): %X\nbz(genproto2): %X\nstart(goo): %v\n",
				bz, bz2, spw(ptr))

			// Unmarshal with genproto2 and re-marshal to check roundtrip.
			rv5 := reflect.New(rt)
			ptr5 := rv5.Interface()
			err = ptr5.(amino.PBMessager2).UnmarshalBinary2(cdc, bz, 0)
			require.NoError(t, err,
				"UnmarshalBinary2 failed: %v\nbz: %X\n", err, bz)
			bz2rt, err := cdc.MarshalBinary2(ptr5.(amino.PBMessager2))
			require.NoError(t, err)
			require.Equal(t, bz, bz2rt,
				"genproto2 roundtrip bytes mismatch.\nbz(amino): %X\nbz(roundtrip): %X\nstart(goo): %v\nend(goo): %v\n",
				bz, bz2rt, spw(ptr), spw(ptr5))
		}
	}
}

func _testDeepCopy(t *testing.T, rt reflect.Type) {
	t.Helper()

	err := error(nil)
	f := fuzz.New()
	rv := reflect.New(rt)
	ptr := rv.Interface()
	rnd := rand.New(rand.NewSource(10))
	f.RandSource(rnd)
	f.Funcs(fuzzFuncs...)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic'd:\nreason: %v\n%s\nerr: %v\nrv: %#v\nptr: %v\n",
				r, debug.Stack(), err, rv, spw(ptr),
			)
		}
	}()

	var deadline time.Time
	if fuzzPerSubtest > 0 {
		deadline = time.Now().Add(fuzzPerSubtest)
	}
	for i := 0; shouldContinue(i, 10_000, deadline); i++ {
		f.Fuzz(ptr)

		ptr2 := amino.DeepCopy(ptr)

		require.Equal(t, ptr, ptr2,
			"end to end failed.\nstart: %v\nend: %v\nbytes: %X\nstring(bytes): %s\n",
			spw(ptr), spw(ptr2))
	}
}

// _testCodecAminoTags mirrors _testCodec but skips the proto.Marshal byte
// comparison, which is invalid for types with amino-specific tags like
// write_empty (forces zero-value emission) or nil_elements (allows nil
// entries in repeated fields) — neither has a proto3 equivalent.
//
// If lossyDecode is true, the decoder may produce a struct that does not
// reflect.DeepEqual to the original (e.g. nil pointer → zero-value struct).
// In that case we use byte-stability + cross-encoder/decoder equivalence
// instead of strict struct equality.
func _testCodecAminoTags(t *testing.T, rt reflect.Type, codecType string, lossyDecode bool) {
	t.Helper()

	err := error(nil)
	bz := []byte{}
	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)
	f := fuzz.New()
	rv := reflect.New(rt)
	rv2 := reflect.New(rt)
	ptr := rv.Interface()
	ptr2 := rv2.Interface()
	rnd := rand.New(rand.NewSource(10))
	f.RandSource(rnd)
	f.Funcs(fuzzFuncs...)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("panic'd:\nreason: %v\n%s\nerr: %v\nbz: %X\nrv: %#v\nrv2: %#v\nptr: %v\nptr2: %v\n",
				r, debug.Stack(), err, bz, rv, rv2, spw(ptr), spw(ptr2),
			)
		}
	}()

	// Clamp the binary path to MarshalReflect/UnmarshalReflect so the
	// cross-encoder (bz vs bz2) and cross-decoder (ptr2 vs ptr5) checks
	// below are genuine reflect-vs-genproto2 comparisons instead of
	// comparing genproto2 to itself via dispatch.
	marshal := func(p any) ([]byte, error) {
		if codecType == "json" {
			return cdc.JSONMarshal(p)
		}
		return cdc.MarshalReflect(p)
	}
	unmarshal := func(b []byte, p any) error {
		if codecType == "json" {
			return cdc.JSONUnmarshal(b, p)
		}
		return cdc.UnmarshalReflect(b, p)
	}

	var deadline time.Time
	if fuzzPerSubtest > 0 {
		deadline = time.Now().Add(fuzzPerSubtest)
	}
	for i := 0; shouldContinue(i, 10_000, deadline); i++ {
		f.Fuzz(ptr)

		rv2 = reflect.New(rt)
		ptr2 = rv2.Interface()

		bz, err = marshal(ptr)
		require.Nil(t, err, "failed to marshal %v: %v\n", spw(ptr), err)

		err = unmarshal(bz, ptr2)
		require.NoError(t, err, "failed to unmarshal %X: %v\n", bz, err)

		if lossyDecode {
			// Byte stability: re-encode the (possibly lossy) decoded value
			// and require the bytes to match. Catches any new divergence.
			bzStable, err := marshal(ptr2)
			require.NoError(t, err, "re-marshal failed: %v\n", err)
			require.Equal(t, bz, bzStable,
				"amino byte-stability failed.\nbz:       %X\nbzStable: %X\nstart: %v\nmid: %v\n",
				bz, bzStable, spw(ptr), spw(ptr2))
		} else {
			require.Equal(t, ptr, ptr2,
				"amino roundtrip failed.\nstart: %v\nend: %v\nbytes: %X\n",
				spw(ptr), spw(ptr2), bz)
		}

		if codecType != "binary" {
			continue
		}

		// ToPBMessage + FromPBMessage. In strict mode we also check struct
		// equality; in lossy mode ToPBMessage/FromPBMessage may share the
		// same lossy decode path, so we only verify it doesn't error.
		pbm, ok := rv.Interface().(amino.PBMessager)
		if !ok {
			continue
		}
		pbo, err := pbm.ToPBMessage(cdc)
		require.NoError(t, err)
		rv3 := reflect.New(rt)
		ptr3 := rv3.Interface()
		err = ptr3.(amino.PBMessager).FromPBMessage(cdc, pbo)
		require.NoError(t, err)
		if !lossyDecode {
			require.Equal(t, ptr, ptr3,
				"ToPBMessage/FromPBMessage roundtrip failed.\nstart: %v\nend: %v\nmid(pbo): %v\n",
				spw(ptr), spw(ptr3), spw(pbo))
		}

		// genproto2 byte equality (cross-encoder) + roundtrip.
		pbm2, ok := rv.Interface().(amino.PBMessager2)
		if !ok {
			continue
		}
		bz2, err := cdc.MarshalBinary2(pbm2)
		require.NoError(t, err, "MarshalBinary2 failed: %v\n", err)
		require.Equal(t, bz, bz2,
			"genproto2 bytes mismatch.\nbz(amino): %X\nbz(genproto2): %X\nstart: %v\n",
			bz, bz2, spw(ptr))

		rv5 := reflect.New(rt)
		ptr5 := rv5.Interface()
		err = ptr5.(amino.PBMessager2).UnmarshalBinary2(cdc, bz, 0)
		require.NoError(t, err, "UnmarshalBinary2 failed: %v\nbz: %X\n", err, bz)
		bz2rt, err := cdc.MarshalBinary2(ptr5.(amino.PBMessager2))
		require.NoError(t, err)
		require.Equal(t, bz, bz2rt,
			"genproto2 roundtrip bytes mismatch.\nbz(amino): %X\nbz(roundtrip): %X\n",
			bz, bz2rt)

		if lossyDecode {
			// Cross-decoder: amino-decoded and genproto2-decoded should
			// produce the same Go value, even if both share the same bug.
			require.Equal(t, ptr2, ptr5,
				"cross-decoder struct mismatch.\namino: %v\ngenproto2: %v\nbytes: %X\n",
				spw(ptr2), spw(ptr5), bz)
		}
	}
}

// ----------------------------------------
// Register/interface tests

func TestCodecMashalFailsOnUnregisteredConcrete(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()

	bz, err := cdc.Marshal(struct{ tests.Interface1 }{tests.Concrete1{}})
	assert.Error(t, err, "concrete type not registered")
	assert.Empty(t, bz)
}

func TestCodecMarshalPassesOnRegistered(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterTypeFrom(reflect.TypeOf(tests.Concrete1{}), tests.Package)

	bz, err := cdc.Marshal(struct{ tests.Interface1 }{tests.Concrete1{}})
	assert.NoError(t, err, "correctly registered")
	assert.Equal(t,
		//     0x0a --> field #1 Typ3ByteLength (anonymous struct)
		//           0x12 --> length prefix (18 bytes)
		//                 0x0a --> field #1 Typ3ByteLength (Any)
		//                       0x10 --> length prefix (12 bytes)
		//                             0x2f, ... 0x31 --> "/tests.Concrete1"
		[]byte{0x0a, 0x12, 0x0a, 0x10, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x63, 0x72, 0x65, 0x74, 0x65, 0x31},
		bz,
		"bytes did not match")
}

func TestCodecRegisterAndMarshalMultipleConcrete(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterTypeFrom(reflect.TypeOf(tests.Concrete1{}), tests.Package)
	cdc.RegisterTypeFrom(reflect.TypeOf(tests.Concrete2{}), tests.Package)

	{ // test tests.Concrete1, no conflict.
		bz, err := cdc.Marshal(struct{ tests.Interface1 }{tests.Concrete1{}})
		assert.NoError(t, err, "correctly registered")
		assert.Equal(t,
			//     0x0a --> field #1 Typ3ByteLength (anonymous struct)
			//           0x12 --> length prefix (18 bytes)
			//                 0x0a --> field #1 Typ3ByteLength (Any)
			//                       0x10 --> length prefix (12 bytes)
			//                             0x2f, ... 0x31 --> "/tests.Concrete1"
			[]byte{0x0a, 0x12, 0x0a, 0x10, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x63, 0x72, 0x65, 0x74, 0x65, 0x31},
			bz,
			"bytes did not match")
	}

	{ // test tests.Concrete2, no conflict
		bz, err := cdc.Marshal(struct{ tests.Interface1 }{tests.Concrete2{}})
		assert.NoError(t, err, "correctly registered")
		assert.Equal(t,
			//     0x0a --> field #1 Typ3ByteLength (anonymous struct)
			//           0x12 --> length prefix (18 bytes)
			//                 0x0a --> field #1 Typ3ByteLength (Any TypeURL)
			//                       0x10 --> length prefix (12 bytes)
			//                             0x2f, ... 0x31 --> "/tests.Concrete2"
			[]byte{0x0a, 0x12, 0x0a, 0x10, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x63, 0x72, 0x65, 0x74, 0x65, 0x32},
			bz,
			"bytes did not match")
	}
}

// Serialize and deserialize a registered typedef.
func TestCodecRoundtripNonNilRegisteredTypeDef(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterTypeFrom(reflect.TypeOf(tests.ConcreteTypeDef{}), tests.Package)

	c3 := tests.ConcreteTypeDef{}
	copy(c3[:], []byte("0123"))

	bz, err := cdc.Marshal(struct{ tests.Interface1 }{c3})
	assert.Nil(t, err)
	assert.Equal(t,
		//     0x0a --> field #1 Typ3ByteLength (anonymous struct)
		//           0x20 --> length prefix (32 bytes)
		//                 0x0a --> field #1 Typ3ByteLength (Any TypeURL)
		//                       0x16 --> length prefix (18 bytes)
		//                             0x2f, ... 0x31 --> "/tests.ConcreteTypeDef"
		[]byte{
			0x0a, 0x20, 0x0a, 0x16, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x63, 0x72, 0x65, 0x74, 0x65, 0x54, 0x79, 0x70, 0x65, 0x44, 0x65, 0x66,
			//   0x12 --> field #2 Typ3ByteLength (Any Value)
			//         0x06 --> length prefix (6 bytes)
			//               0x0a --> field #1, one and only, of implicit struct.
			//                     0x04 --> length prefix (4 bytes)
			/**/ 0x12, 0x06, 0x0a, 0x04, 0x30, 0x31, 0x32, 0x33,
		},
		bz,
		"ConcreteTypeDef incorrectly serialized")

	var i1 tests.Interface1
	err = cdc.Unmarshal(bz, &i1)
	assert.Error(t, err) // This fails, because the interface was wrapped in an anonymous struct.

	// try wrapping it in an Any struct
	// without changing the existing behavior.
	type anyType struct {
		TypeURL string
		Value   []byte
	}
	anyc3 := anyType{
		TypeURL: "/tests.ConcreteTypeDef",
		Value:   []byte{0x0a, 0x04, 0x30, 0x31, 0x32, 0x33}, // An implicit struct, the first field which is the length-prefixed 4 bytes.
	}

	// var i1c3 tests.Interface1 = c3
	// bz, err = cdc.Marshal(&i1c3)
	bz, err = cdc.Marshal(anyc3)
	assert.Nil(t, err)
	assert.Equal(t,
		//     0x0a --> field #1 Typ3ByteLength (Any TypeURL)
		//           0x16 --> length prefix (22 bytes)
		//                 0x2f, ... 0x33 --> "/tests.ConcreteTypeDef"
		[]byte{
			0x0a, 0x16, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x63, 0x72, 0x65, 0x74, 0x65, 0x54, 0x79, 0x70, 0x65, 0x44, 0x65, 0x66,
			//   0x12 --> field #2 Typ3ByteLength (Any Value)
			//         0x06 --> length prefix (6 bytes)
			//               0x0a --> field #1, one and only, of implicit struct.
			//                     0x04 --> length prefix (4 bytes)
			/**/ 0x12, 0x06, 0x0a, 0x04, 0x30, 0x31, 0x32, 0x33,
		},
		bz,
		"ConcreteTypeDef incorrectly serialized")

	// This time it should work.
	err = cdc.Unmarshal(bz, &i1)
	assert.NoError(t, err)
	assert.Equal(t, c3, i1)

	// The easiest way is this:
	bz2, err := cdc.MarshalAny(c3)
	assert.Nil(t, err)
	assert.Equal(t, bz, bz2)
}

// Exactly like TestCodecRoundtripNonNilRegisteredTypeDef but with struct
// around the value instead of a type def.
func TestCodecRoundtripNonNilRegisteredWrappedValue(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterTypeFrom(reflect.TypeOf(tests.ConcreteWrappedBytes{}), tests.Package)

	c3 := tests.ConcreteWrappedBytes{Value: []byte("0123")}

	bz, err := cdc.MarshalAny(c3)
	assert.Nil(t, err)
	assert.Equal(t,
		//     0x0a --> field #1 Typ3ByteLength (Any TypeURL)
		//           0x1b --> length prefix (27 bytes)
		//                 0x2f, ... 0x33 --> "/tests.ConcreteWrappedBytes"
		[]byte{
			0x0a, 0x1b, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x63, 0x72, 0x65, 0x74, 0x65, 0x57, 0x72, 0x61, 0x70, 0x70, 0x65, 0x64, 0x42, 0x79, 0x74, 0x65, 0x73,
			//   0x12 --> field #2 Typ3ByteLength (Any Value)
			//         0x06 --> length prefix (6 bytes)
			//               0x0a --> field #1, one and only, of implicit struct.
			//                     0x04 --> length prefix (4 bytes)
			/**/ 0x12, 0x06, 0x0a, 0x04, 0x30, 0x31, 0x32, 0x33,
		},
		bz,
		"ConcreteWrappedBytes incorrectly serialized")

	var i1 tests.Interface1
	err = cdc.Unmarshal(bz, &i1)
	assert.NoError(t, err)
	assert.Equal(t, c3, i1)
}

// MarshalAny(msg) and Marshal(&msg) are the same.
func TestCodecMarshalAny(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterTypeFrom(reflect.TypeOf(tests.ConcreteWrappedBytes{}), tests.Package)

	obj := tests.ConcreteWrappedBytes{Value: []byte("0123")}
	ifc := (any)(obj)

	bz1, err := cdc.MarshalAny(obj)
	assert.Nil(t, err)

	bz2, err := cdc.Marshal(&ifc)
	assert.Nil(t, err)

	assert.Equal(t, bz1, bz2, "Marshal(*interface) or MarshalAny(concrete) incorrectly serialized\nMarshalAny(concrete): %X\nMarshal(*interface):  %X", bz1, bz2)
}

// Like TestCodecRoundtripNonNilRegisteredTypeDef, but JSON.
func TestCodecJSONRoundtripNonNilRegisteredTypeDef(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterTypeFrom(reflect.TypeOf(tests.ConcreteTypeDef{}), tests.Package)

	var c3 tests.ConcreteTypeDef
	copy(c3[:], []byte("0123"))

	bz, err := cdc.MarshalJSONAny(c3)
	assert.Nil(t, err)
	assert.Equal(t,
		`{"@type":"/tests.ConcreteTypeDef","value":"MDEyMw=="}`, string(bz),
		"ConcreteTypeDef incorrectly serialized")

	var i1 tests.Interface1
	err = cdc.JSONUnmarshal(bz, &i1)
	assert.Nil(t, err)
	assert.Equal(t, c3, i1)
}

// Like TestCodecRoundtripNonNilRegisteredTypeDef, but serialize the concrete value directly.
func TestCodecRoundtripMarshalOnConcreteNonNilRegisteredTypeDef(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterTypeFrom(reflect.TypeOf(tests.ConcreteTypeDef{}), tests.Package)

	var c3 tests.ConcreteTypeDef
	copy(c3[:], []byte("0123"))

	bz, err := cdc.MarshalAny(c3)
	assert.Nil(t, err)
	assert.Equal(t,
		//     0x0a --> field #1 Typ3ByteLength (Any TypeURL)
		//           0x16 --> length prefix (18 bytes)
		//                 0x2f, ... 0x31 --> "/tests.ConcreteTypeDef"
		[]byte{
			0x0a, 0x16, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x73, 0x2e, 0x43, 0x6f, 0x6e, 0x63, 0x72, 0x65, 0x74, 0x65, 0x54, 0x79, 0x70, 0x65, 0x44, 0x65, 0x66,
			//   0x12 --> field #2 Typ3ByteLength (Any Value)
			//         0x06 --> length prefix (6 bytes)
			//               0x0a --> field #1, one and only, of implicit struct.
			//                     0x04 --> length prefix (4 bytes)
			/**/ 0x12, 0x06, 0x0a, 0x04, 0x30, 0x31, 0x32, 0x33,
		},
		bz,
		"ConcreteTypeDef incorrectly serialized")

	var i1 tests.Interface1
	err = cdc.Unmarshal(bz, &i1)
	assert.NoError(t, err)
	assert.Equal(t, c3, i1)
}

// Like TestCodecRoundtripNonNilRegisteredTypeDef but read into concrete var.
func TestCodecRoundtripUnmarshalOnConcreteNonNilRegisteredTypeDef(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterTypeFrom(reflect.TypeOf(tests.ConcreteTypeDef{}), tests.Package)

	var c3a tests.ConcreteTypeDef
	copy(c3a[:], []byte("0123"))

	bz, err := cdc.Marshal(c3a)
	assert.Nil(t, err)
	assert.Equal(t,
		[]byte{0xa, 0x4, 0x30, 0x31, 0x32, 0x33}, bz,
		"ConcreteTypeDef incorrectly serialized")

	var c3b tests.ConcreteTypeDef
	err = cdc.Unmarshal(bz, &c3b)
	assert.Nil(t, err)
	assert.Equal(t, c3a, c3b)
}

func TestCodecBinaryStructFieldNilInterface(t *testing.T) {
	t.Parallel()

	cdc := amino.NewCodec()
	cdc.RegisterTypeFrom(reflect.TypeOf(tests.InterfaceFieldsStruct{}), tests.Package)

	i1 := &tests.InterfaceFieldsStruct{F1: new(tests.InterfaceFieldsStruct), F2: nil}
	bz, err := cdc.MarshalSized(i1)
	assert.NoError(t, err)

	i2 := new(tests.InterfaceFieldsStruct)
	err = cdc.UnmarshalSized(bz, i2)

	assert.NoError(t, err)
	require.Equal(t, i1, i2, "i1 and i2 should be the same after decoding")
}

// ----------------------------------------
// Misc.

func spw(o any) string {
	return spew.Sprintf("%#v", o)
}

var fuzzFuncs = []any{
	func(ptr **int8, c fuzz.Continue) {
		var i int8
		c.Fuzz(&i)
		*ptr = &i
	},
	func(ptr **int16, c fuzz.Continue) {
		var i int16
		c.Fuzz(&i)
		*ptr = &i
	},
	func(ptr **int32, c fuzz.Continue) {
		var i int32
		c.Fuzz(&i)
		*ptr = &i
	},
	func(ptr **int64, c fuzz.Continue) {
		var i int64
		c.Fuzz(&i)
		*ptr = &i
	},
	func(ptr **int, c fuzz.Continue) {
		var i int
		c.Fuzz(&i)
		*ptr = &i
	},
	func(ptr **uint8, c fuzz.Continue) {
		var ui uint8
		c.Fuzz(&ui)
		*ptr = &ui
	},
	/* go-amino 1.2 removed nested pointer support
	func(ptr ***uint8, c fuzz.Continue) {
		var ui uint8
		c.Fuzz(&ui)
		*ptr = new(*uint8)
		**ptr = new(uint8)
		***ptr = ui
	},
	func(ptr ****uint8, c fuzz.Continue) {
		var ui uint8
		c.Fuzz(&ui)
		*ptr = new(**uint8)
		**ptr = new(*uint8)
		***ptr = new(uint8)
		****ptr = ui
	},
	*/
	func(ptr **uint16, c fuzz.Continue) {
		var ui uint16
		c.Fuzz(&ui)
		*ptr = &ui
	},
	func(ptr **uint32, c fuzz.Continue) {
		var ui uint32
		c.Fuzz(&ui)
		*ptr = &ui
	},
	func(ptr **uint64, c fuzz.Continue) {
		var ui uint64
		c.Fuzz(&ui)
		*ptr = &ui
	},
	func(ptr **uint, c fuzz.Continue) {
		var ui uint
		c.Fuzz(&ui)
		*ptr = &ui
	},
	func(ptr **string, c fuzz.Continue) {
		// Prefer nil instead of zero, for deep equality.
		// (go-amino decoder will always prefer nil).
		s := randString(c)
		if len(s) == 0 {
			*ptr = nil
		} else {
			*ptr = &s
		}
	},
	func(bz **[]byte, c fuzz.Continue) {
		// Prefer nil instead of zero, for deep equality.
		// (go-amino decoder will always prefer nil).
		var by []byte
		c.Fuzz(&by)
		*bz = &by
	},
	func(tyme *time.Time, c fuzz.Continue) {
		// Set time.Unix(_,_) to wipe .wal
		switch c.Intn(4) {
		case 0:
			ns := c.Int63n(10)
			*tyme = time.Unix(0, ns)
		case 1:
			ns := c.Int63n(1e10)
			*tyme = time.Unix(0, ns)
		case 2:
			const maxSeconds = 4611686018 // (1<<63 - 1) / 1e9
			s := c.Int63n(maxSeconds)
			ns := c.Int63n(1e10)
			*tyme = time.Unix(s, ns)
		case 3:
			s := c.Int63n(10)
			ns := c.Int63n(1e10)
			*tyme = time.Unix(s, ns)
		}
		// Strip timezone and monotonic for deep equality.
		// Also set to UTC.
		*tyme = tyme.Truncate(0).UTC()
	},
	func(ptr **time.Duration, c fuzz.Continue) {
		// Zero should decode to ptr to zero duration,
		// rather than a nil duration pointer.
		switch c.Intn(4) {
		case 0:
			ns := c.Int63n(20) - 10
			dur := time.Duration(ns)
			*ptr = &dur
		case 1:
			ns := c.Int63n(2e10) - 1e10
			dur := time.Duration(ns)
			*ptr = &dur
		case 2: // NOTE: not max p3 duration
			ns := 1<<63 - 1
			dur := time.Duration(ns)
			*ptr = &dur
		case 3: // NOTE: not min p3 duration
			ns := -1<<63 + 1
			dur := time.Duration(ns)
			*ptr = &dur
		}
	},
	func(esz *[]*tests.EmptyStruct, c fuzz.Continue) {
		n := c.Intn(4)
		switch n {
		case 0:
			// Prefer nil over empty slice.
			*esz = nil
		default:
			// Empty slice elements should be non-nil,
			// since we don't set amino:"nil_elements".
			*esz = make([]*tests.EmptyStruct, n)
			for i := range n {
				(*esz)[i] = &tests.EmptyStruct{}
			}
		}
	},
	func(sl *[]*tests.StructWithStringRepr, c fuzz.Continue) {
		n := c.Intn(4)
		switch n {
		case 0:
			*sl = nil
		default:
			// Slice elements must be non-nil, since StructPtrSliceWithStringRepr
			// does not set amino:"nil_elements" (the fix #8 contract).
			*sl = make([]*tests.StructWithStringRepr, n)
			for i := range n {
				var elem tests.StructWithStringRepr
				c.Fuzz(&elem)
				(*sl)[i] = &elem
			}
		}
	},
	func(sl *[]*tests.FuzzFieldInfo, c fuzz.Continue) {
		n := c.Intn(4)
		switch n {
		case 0:
			*sl = nil
		default:
			// Include nil elements to exercise amino:"nil_elements".
			*sl = make([]*tests.FuzzFieldInfo, n)
			for i := range n {
				if c.Intn(3) == 0 {
					(*sl)[i] = nil // nil element
				} else {
					var fi tests.FuzzFieldInfo
					c.Fuzz(&fi)
					(*sl)[i] = &fi
				}
			}
		}
	},
	func(sl *[]*tests.GnoVMPos, c fuzz.Continue) {
		n := c.Intn(4)
		switch n {
		case 0:
			*sl = nil
		default:
			// Include nil elements to exercise amino:"nil_elements".
			*sl = make([]*tests.GnoVMPos, n)
			for i := range n {
				if c.Intn(3) == 0 {
					(*sl)[i] = nil // nil element
				} else {
					var p tests.GnoVMPos
					c.Fuzz(&p)
					(*sl)[i] = &p
				}
			}
		}
	},
	func(f *float64, c fuzz.Continue) {
		// Exercise amino:"unsafe" float encoding.
		// Avoid NaN/Inf which don't roundtrip via proto encoding.
		switch c.Intn(4) {
		case 0:
			*f = 0
		case 1:
			*f = float64(c.Int63n(1000)) / 100.0
		case 2:
			*f = -float64(c.Int63n(1000)) / 100.0
		case 3:
			*f = float64(c.Int63())
		}
	},
	func(f *float32, c fuzz.Continue) {
		switch c.Intn(4) {
		case 0:
			*f = 0
		case 1:
			*f = float32(c.Intn(1000)) / 100.0
		case 2:
			*f = -float32(c.Intn(1000)) / 100.0
		case 3:
			*f = float32(c.Int31())
		}
	},
	func(iface *tests.Interface1, c fuzz.Continue) {
		// Randomly pick a concrete type for the interface.
		switch c.Intn(4) {
		case 0:
			*iface = nil
			return
		case 1:
			*iface = tests.Concrete1{}
		case 2:
			*iface = tests.Concrete2{}
		case 3:
			n := c.Intn(20)
			if n == 0 {
				*iface = tests.ConcreteWrappedBytes{Value: nil}
			} else {
				bz := make([]byte, n)
				for i := range bz {
					bz[i] = byte(c.Intn(256))
				}
				*iface = tests.ConcreteWrappedBytes{Value: bz}
			}
		}
	},
	func(sl *[]tests.Interface1, c fuzz.Continue) {
		n := c.Intn(4)
		if n == 0 {
			*sl = nil
			return
		}
		*sl = make([]tests.Interface1, n)
		for i := range n {
			c.Fuzz(&(*sl)[i])
		}
	},
}

func getTypeFromPointer(ptr any) reflect.Type {
	rt := reflect.TypeOf(ptr)
	if rt.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("expected pointer, got %v", rt))
	}
	return rt.Elem()
}

// ----------------------------------------
// From https://github.com/google/gofuzz/blob/master/fuzz.go
// (Apache2.0 License)

type charRange struct {
	first, last rune
}

// choose returns a random unicode character from the given range, using the
// given randomness source.
func (r *charRange) choose(rand fuzz.Continue) rune {
	count := int64(r.last - r.first)
	return r.first + rune(rand.Int63n(count))
}

var unicodeRanges = []charRange{
	{' ', '~'},           // ASCII characters
	{'\u00a0', '\u02af'}, // Multi-byte encoded characters
	{'\u4e00', '\u9fff'}, // Common CJK (even longer encodings)
}

// randString makes a random string up to 20 characters long. The returned string
// may include a variety of (valid) UTF-8 encodings.
func randString(r fuzz.Continue) string {
	n := r.Intn(19) + 1
	runes := make([]rune, n)
	for i := range runes {
		runes[i] = unicodeRanges[r.Intn(len(unicodeRanges))].choose(r)
	}
	return string(runes)
}
