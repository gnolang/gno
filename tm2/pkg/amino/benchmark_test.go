package amino_test

import (
	"math/rand"
	"reflect"
	"testing"

	fuzz "github.com/google/gofuzz"
	"google.golang.org/protobuf/proto"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
)

// Pre-generate N random instances of each type and their encoded bytes.
const benchN = 100

type benchData struct {
	ptrs []any          // random struct pointers
	bzs  [][]byte       // amino-encoded bytes
	pbos []proto.Message // pbbindings proto messages (nil if not PBMessager)
}

func makeBenchData(cdc *amino.Codec, rt reflect.Type) benchData {
	f := fuzz.New()
	rnd := rand.New(rand.NewSource(42))
	f.RandSource(rnd)
	f.Funcs(fuzzFuncs...)

	bd := benchData{
		ptrs: make([]any, benchN),
		bzs:  make([][]byte, benchN),
		pbos: make([]proto.Message, benchN),
	}
	for i := range benchN {
		rv := reflect.New(rt)
		ptr := rv.Interface()
		f.Fuzz(ptr)
		bd.ptrs[i] = ptr
		bz, err := cdc.MarshalReflect(ptr)
		if err != nil {
			panic(err)
		}
		bd.bzs[i] = bz
		if pbm, ok := ptr.(amino.PBMessager); ok {
			pbo, err := pbm.ToPBMessage(cdc)
			if err != nil {
				panic(err)
			}
			bd.pbos[i] = pbo
		}
	}
	return bd
}

// BenchmarkEncode compares encode performance: genproto2 vs reflect vs pbbindings.
func BenchmarkEncode(b *testing.B) {
	cdc := amino.NewCodec()
	pbcdc := cdc.WithPBBindings()

	for _, ptr := range tests.StructTypes {
		rt := getTypeFromPointer(ptr)
		name := rt.Name()

		b.Run(name+"/genproto2", func(b *testing.B) {
			bd := makeBenchData(cdc, rt)
			for i := 0; i < b.N; i++ {
				p := bd.ptrs[i%benchN]
				if pbm2, ok := p.(amino.PBMessager2); ok {
					cdc.MarshalBinary2(pbm2)
				} else {
					cdc.MarshalReflect(p)
				}
			}
		})

		b.Run(name+"/reflect", func(b *testing.B) {
			bd := makeBenchData(cdc, rt)
			for i := 0; i < b.N; i++ {
				cdc.MarshalReflect(bd.ptrs[i%benchN])
			}
		})

		b.Run(name+"/pbbindings", func(b *testing.B) {
			bd := makeBenchData(cdc, rt)
			if _, ok := bd.ptrs[0].(amino.PBMessager); !ok {
				b.Skip("not PBMessager")
			}
			for i := 0; i < b.N; i++ {
				pbcdc.MarshalPBBindings(bd.ptrs[i%benchN].(amino.PBMessager))
			}
		})
	}
}

// BenchmarkDecode compares decode performance: genproto2 vs reflect vs pbbindings.
func BenchmarkDecode(b *testing.B) {
	cdc := amino.NewCodec()
	pbcdc := cdc.WithPBBindings()

	for _, ptr := range tests.StructTypes {
		rt := getTypeFromPointer(ptr)
		name := rt.Name()

		b.Run(name+"/genproto2", func(b *testing.B) {
			bd := makeBenchData(cdc, rt)
			for i := 0; i < b.N; i++ {
				rv := reflect.New(rt)
				p := rv.Interface()
				if pbm2, ok := p.(amino.PBMessager2); ok {
					pbm2.UnmarshalBinary2(cdc, bd.bzs[i%benchN])
				} else {
					cdc.UnmarshalReflect(bd.bzs[i%benchN], p)
				}
			}
		})

		b.Run(name+"/reflect", func(b *testing.B) {
			bd := makeBenchData(cdc, rt)
			for i := 0; i < b.N; i++ {
				rv := reflect.New(rt)
				cdc.UnmarshalReflect(bd.bzs[i%benchN], rv.Interface())
			}
		})

		b.Run(name+"/pbbindings", func(b *testing.B) {
			bd := makeBenchData(cdc, rt)
			if _, ok := bd.ptrs[0].(amino.PBMessager); !ok {
				b.Skip("not PBMessager")
			}
			for i := 0; i < b.N; i++ {
				rv := reflect.New(rt)
				pbm := rv.Interface().(amino.PBMessager)
				pbo := pbm.EmptyPBMessage(pbcdc)
				proto.Unmarshal(bd.bzs[i%benchN], pbo)
				pbm.FromPBMessage(pbcdc, pbo)
			}
		})
	}
}
