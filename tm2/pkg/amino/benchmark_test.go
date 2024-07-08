package amino_test

import (
	"math/rand"
	"reflect"
	"runtime/debug"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
)

func BenchmarkBinary(b *testing.B) {
	if testing.Short() {
		b.Skip("skipping testing in short mode")
	}

	cdc := amino.NewCodec()
	for _, ptr := range tests.StructTypes {
		b.Logf("case %v", reflect.TypeOf(ptr))
		rt := getTypeFromPointer(ptr)
		name := rt.Name()
		b.Run(name+":encode", func(b *testing.B) {
			_benchmarkBinary(b, cdc, rt, "binary", true)
		})
		b.Run(name+":decode", func(b *testing.B) {
			_benchmarkBinary(b, cdc, rt, "binary", false)
		})
	}
}

func BenchmarkBinaryPBBindings(b *testing.B) {
	b.Skip("fuzzing not benchmarking")

	cdc := amino.NewCodec().WithPBBindings()
	for _, ptr := range tests.StructTypes {
		b.Logf("case %v (pbbindings)", reflect.TypeOf(ptr))
		rt := getTypeFromPointer(ptr)
		name := rt.Name()
		b.Run(name+":encode:pbbindings", func(b *testing.B) {
			_benchmarkBinary(b, cdc, rt, "binary_pb", true)
		})

		// TODO: fix nil pointer error
		b.Run(name+":encode:pbbindings:translate_only", func(b *testing.B) {
			_benchmarkBinary(b, cdc, rt, "binary_pb_translate_only", true)
		})

		b.Run(name+":decode:pbbindings", func(b *testing.B) {
			_benchmarkBinary(b, cdc, rt, "binary_pb", false)
		})

		// TODO: fix nil pointer error
		b.Run(name+":decode:pbbindings:translate_only", func(b *testing.B) {
			_benchmarkBinary(b, cdc, rt, "binary_pb_translate_only", false)
		})
	}
}

func _benchmarkBinary(b *testing.B, cdc *amino.Codec, rt reflect.Type, codecType string, encode bool) {
	b.Helper()

	b.StopTimer()

	err := error(nil)
	bz := []byte{}
	f := fuzz.New()
	pbcdc := cdc.WithPBBindings()
	rv := reflect.New(rt)
	rv2 := reflect.New(rt)
	ptr := rv.Interface()
	ptr2 := rv2.Interface()
	rnd := rand.New(rand.NewSource(10))
	f.RandSource(rnd)
	f.Funcs(fuzzFuncs...)
	pbm := amino.PBMessager(nil)
	pbo := proto.Message(nil)

	defer func() {
		if r := recover(); r != nil {
			b.Fatalf("panic'd:\nreason: %v\n%s\nerr: %v\nbz: %X\nrv: %#v\nrv2: %#v\nptr: %v\nptr2: %v\n",
				r, debug.Stack(), err, bz, rv, rv2, spw(ptr), spw(ptr2),
			)
		}
	}()

	for i := 0; i < b.N; i++ {
		f.Fuzz(ptr)

		// Reset, which makes debugging decoding easier.
		rv2 = reflect.New(rt)
		ptr2 = rv2.Interface()

		// Encode to bz.
		if encode {
			b.StartTimer()
		}
		switch codecType {
		case "binary":
			bz, err = cdc.Marshal(ptr)
		case "json":
			bz, err = cdc.MarshalJSON(ptr)
		case "binary_pb":
			bz, err = pbcdc.Marshal(ptr)
		case "binary_pb_translate_only":
			pbm, _ = ptr.(amino.PBMessager)
			pbo, err = pbm.ToPBMessage(pbcdc)
		default:
			panic("should not happen")
		}
		if encode {
			b.StopTimer()
		}

		// Check for errors
		require.Nil(b, err,
			"failed to marshal %v to bytes: %v\n",
			spw(ptr), err)

		// Decode from bz.
		if !encode {
			b.StartTimer()
		}
		switch codecType {
		case "binary":
			err = cdc.Unmarshal(bz, ptr2)
		case "json":
			err = cdc.UnmarshalJSON(bz, ptr2)
		case "binary_pb":
			err = pbcdc.Unmarshal(bz, ptr2)
		case "binary_pb_translate_only":
			err = pbm.FromPBMessage(pbcdc, pbo)
		default:
			panic("should not happen")
		}
		if !encode {
			b.StopTimer()
		}

		if codecType != "binary_pb_translate_only" {
			// Decode for completeness and check for errors,
			// in case there were encoding/decoding issues.
			require.NoError(b, err,
				"failed to unmarshal bytes %X (%s): %v\nptr: %v\n",
				bz, bz, err, spw(ptr))
			require.Equal(b, ptr2, ptr,
				"end to end failed.\nstart: %v\nend: %v\nbytes: %X\nstring(bytes): %s\n",
				spw(ptr), spw(ptr2), bz, bz)
		}
	}
}
