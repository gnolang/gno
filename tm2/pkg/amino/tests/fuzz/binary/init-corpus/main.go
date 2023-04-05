package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"

	amino "github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
)

func main() {
	corpusParentDir := flag.String("corpus-parent", ".", "the directory in which we should place the corpus directory")
	flag.Parse()

	corpusDir := filepath.Join(*corpusParentDir, "corpus")
	if err := os.MkdirAll(corpusDir, 0o755); err != nil {
		log.Fatalf("Cannot mkdirAll: %q err: %v", corpusDir, err)
	}

	// Let's seed the fuzzer by filling in the tests
	// manually.
	ps := tests.PrimitivesStruct{
		Int8:        0x7F,
		Int16:       0x7FFF,
		Int32:       0x1EADBEEF,
		Int32Fixed:  0x7FFFFFFF,
		Int64:       0x00,
		Int64Fixed:  0x00,
		Int:         0x7FFFFFFF,
		Byte:        0xCD,
		Uint8:       0xFF,
		Uint16:      0xFFFF,
		Uint32:      0xFFFFFFFF,
		Uint32Fixed: 0xFFFFFFFF,
		Uint64:      0x8000000000000000,
		Uint64Fixed: 0x8000000000000000,
		Uint:        0x80000000,
		Str:         "Tendermint!",
		Bytes:       []byte("DEEZMINTS"),
		Time:        time.Date(2018, 3, 2, 21, 10, 12, 1e5, time.UTC),
	}

	hour := 60 * time.Minute
	as := tests.ArraysStruct{
		Int8Ar:   [4]int8{0x7F, 0x6F, 0x5F, 0x4F},
		Int16Ar:  [4]int16{0x7FFF, 0x6FFF, 0x5FFF, 0x00},
		Int32Ar:  [4]int32{0x7FFFFFFF, 0x6FFFFFFF, 0x5FFFFFFF, 0x77777777},
		Int64Ar:  [4]int64{0x7FFFFFFFFFFFF, 0x6FFFFFFFFFFFF, 0x5FFFFFFFFFFFF, 0x80808000FFFFF},
		IntAr:    [4]int{0x7FFFFFFF, 0x6FFFFFFF, 0x5FFFFFFF, math.MaxInt32},
		ByteAr:   [4]byte{0xDE, 0xAD, 0xBE, 0xEF},
		Uint8Ar:  [4]uint8{0xFF, 0xFF, 0x00, 0x88},
		Uint16Ar: [4]uint16{0xFFFF, 0xFFFF, 0xFF00, 0x8800},
		Uint32Ar: [4]uint32{0x80808080, 0x110202FF, 0xAE21FF00, 0x10458800},
		Uint64Ar: [4]uint64{0x80808080FFFFFF77, 0x110202FFFFFFFF77, 0xAE21FF0051F23F77, 0x1045880011AABBCC},
		UintAr:   [4]uint{0x80808080, 0x110202FF, 0xAE21FF00, 0x10458800},
		StrAr:    [4]string{"Tendermint", "Fuzzing", "Blue", "410DDC670CF9BFD7"},
		TimeAr:   [4]time.Time{{}, time.Time{}.Add(1000 * hour * 24), time.Time{}.Add(20 * time.Minute)},
	}

	ss := tests.SlicesStruct{
		Int8Sl:   []int8{0x6F, 0x5F, 0x7F, 0x4F},
		Int16Sl:  []int16{0x6FFF, 0x5FFF, 0x7FFF, 0x00},
		Int32Sl:  []int32{0x6FFFFFFF, 0x5FFFFFFF, 0x7FFFFFFF, 0x7F000000},
		Int64Sl:  []int64{0x6FFFFFFFFFFFF, 0x5FFFFFFFFFFFF, 0x7FFFFFFFFFFFF, 0x80808000FFFFF},
		IntSl:    []int{0x6FFFFFFF, 0x7FFFFFFF, math.MaxInt32, 0x5FFFFFFF},
		ByteSl:   []byte{0xAD, 0xBE, 0xDE, 0xEF},
		Uint8Sl:  []uint8{0xFF, 0x00, 0x88, 0xFF},
		Uint16Sl: []uint16{0xFFFF, 0xFFFF, 0xFF00, 0x8800},
		Uint32Sl: []uint32{0x110202FF, 0xAE21FF00, 0x80808080, 0x10458800},
		Uint64Sl: []uint64{0x110202FFFFFFFF77, 0xAE21FF0051F23F77, 0x80808080FFFFFF77, 0x1045880011AABBCC},
		UintSl:   []uint{0x80808080, 0x110202FF, 0xAE21FF00, 0x10458800},
		StrSl:    []string{"Tendermint", "Fuzzing", "Blue", "410DDC670CF9BFD7"},
		TimeSl: []time.Time{
			(time.Time{}).Add(60 * 24 * time.Minute), (time.Time{}).Add(1000 * hour * 24), time.Time{}.Add(20 * time.Minute),
		},
	}

	bslice := []byte("VIVA LA VIDA!")
	pts1 := tests.PointersStruct{}
	pts2 := tests.PointersStruct{
		Int8Pt:   new(int8),
		Int16Pt:  &ss.Int16Sl[0],
		Int32Pt:  new(int32),
		Int64Pt:  &ss.Int64Sl[2],
		IntPt:    &as.IntAr[3],
		BytePt:   &ss.ByteSl[0],
		Uint8Pt:  new(uint8),
		Uint16Pt: &ss.Uint16Sl[2],
		Uint32Pt: &ss.Uint32Sl[1],
		Uint64Pt: &ss.Uint64Sl[0],
		UintPt:   &ss.UintSl[2],
		StrPt:    &as.StrAr[1],
		BytesPt:  &bslice,
		TimePt:   &ss.TimeSl[2],
	}

	seeds := []*tests.ComplexSt{
		{PrField: ps, ArField: tests.ArraysStruct{}, SlField: tests.SlicesStruct{}, PtField: tests.PointersStruct{}},
		{PrField: ps, ArField: as, SlField: tests.SlicesStruct{}, PtField: tests.PointersStruct{}},
		{PrField: ps, ArField: as, SlField: ss, PtField: tests.PointersStruct{}},
		{PrField: ps, ArField: as, SlField: ss, PtField: pts1},
		{PrField: ps, ArField: as, SlField: ss, PtField: pts2},

		{PrField: tests.PrimitivesStruct{}, ArField: as, SlField: tests.SlicesStruct{}, PtField: tests.PointersStruct{}},
		{PrField: ps, ArField: as, SlField: tests.SlicesStruct{}, PtField: tests.PointersStruct{}},
		{PrField: ps, ArField: as, SlField: ss, PtField: tests.PointersStruct{}},
		{PrField: ps, ArField: as, SlField: ss, PtField: pts1},
		{PrField: ps, ArField: as, SlField: ss, PtField: pts2},

		{PrField: tests.PrimitivesStruct{}, ArField: tests.ArraysStruct{}, SlField: ss, PtField: tests.PointersStruct{}},
		{PrField: ps, ArField: tests.ArraysStruct{}, SlField: ss, PtField: tests.PointersStruct{}},
		{PrField: ps, ArField: as, SlField: ss, PtField: tests.PointersStruct{}},
		{PrField: ps, ArField: as, SlField: ss, PtField: pts1},
		{PrField: ps, ArField: as, SlField: ss, PtField: pts2},

		{PrField: tests.PrimitivesStruct{}, ArField: tests.ArraysStruct{}, SlField: tests.SlicesStruct{}, PtField: pts2},
		{PrField: ps, ArField: tests.ArraysStruct{}, SlField: tests.SlicesStruct{}, PtField: pts2},
		{PrField: ps, ArField: as, SlField: tests.SlicesStruct{}, PtField: pts2},
		{PrField: ps, ArField: as, SlField: ss, PtField: pts2},

		{PrField: tests.PrimitivesStruct{}, ArField: tests.ArraysStruct{}, SlField: tests.SlicesStruct{}, PtField: pts1},
		{PrField: ps, ArField: tests.ArraysStruct{}, SlField: tests.SlicesStruct{}, PtField: pts1},
		{PrField: ps, ArField: as, SlField: tests.SlicesStruct{}, PtField: pts1},
		{PrField: ps, ArField: as, SlField: ss, PtField: pts1},
	}

	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)

	for i, seed := range seeds {
		blob, err := cdc.MarshalSized(seed)
		if err != nil {
			log.Fatalf("Failed to Marshal on seed: %d", i)
		}

		fullPath := filepath.Join(corpusDir, fmt.Sprintf("%d", i))
		f, err := os.Create(fullPath)
		if err != nil {
			log.Fatalf("Failed to create path: %q", fullPath)
		}
		_, err = f.Write(blob)
		if err != nil {
			log.Fatalf("failed to write to file")
		}
		_ = f.Close()
	}
}
