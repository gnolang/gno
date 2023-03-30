package main

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	amino "github.com/gnolang/gno/tm2/pkg/amino"
)

func main() {
	// Print help.
	if len(os.Args) == 1 {
		fmt.Println(`Usage: aminoscan <STRUCT HEXBYTES> or --help`)
		return
	}

	// Parse flags...
	var colorize bool
	flgs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	flgs.BoolVar(&colorize, "color", false, "Just print the colored bytes and exit.")
	err := flgs.Parse(os.Args[1:])
	if errors.Is(err, flag.ErrHelp) {
		fmt.Println(`Usage: aminoscan <STRUCT HEXBYTES> or --help
		
		You can also use aminoscan to print "colored" bytes.  This view will
		try to display bytes in ascii in a different color if it happens to be
		a printable character.

		> aminoscan --color <HEXBYTES>`)
		return
	} else if err != nil {
		fmt.Println(err)
		return
	}

	// If we just want to show colored bytes...
	if colorize {
		if flgs.Arg(0) == "" {
			fmt.Println(`Usage: aminoscan --color <HEXBYTES>`)
			return
		}
		bz := hexDecode(flgs.Arg(0))
		fmt.Println(ColoredBytes(bz, Green, Blue))
		return
	}

	// Parse struct Amino bytes.
	bz := hexDecode(os.Args[1]) // Read input hex bytes.
	fmt.Println(Yellow("## Root Struct (assumed)"))
	s, n, err := scanStruct(bz, "", true) // Assume that it's  struct.
	s += Red(fmt.Sprintf("%X", bz[n:]))   // Bytes remaining are red.
	fmt.Println(Yellow("## Root Struct END"))
	fmt.Println(s, n, err) // Print color-encoded bytes s.
}

func scanAny(typ amino.Typ3, bz []byte, indent string) (s string, n int, err error) {
	switch typ {
	case amino.Typ3Varint:
		s, n, err = scanVarint(bz, indent)
	case amino.Typ38Byte:
		s, n, err = scan8Byte(bz, indent)
	case amino.Typ3ByteLength:
		s, n, err = scanByteLength(bz, indent)
	case amino.Typ34Byte:
		s, n, err = scan4Byte(bz, indent)
	default:
		panic("should not happen")
	}
	return
}

func scanVarint(bz []byte, indent string) (s string, n int, err error) {
	if len(bz) == 0 {
		err = fmt.Errorf("EOF while reading (U)Varint")
	}
	// First try Varint.
	okI64 := true
	i64, n := binary.Varint(bz)
	if n <= 0 {
		n = 0
		okI64 = false
	}
	// Then try Uvarint.
	okU64 := true
	u64, _n := binary.Uvarint(bz)
	if n != _n {
		n = 0
		okU64 = false
	}
	// If neither work, return error.
	if !okI64 && !okU64 {
		err = fmt.Errorf("invalid (u)varint")
		return
	}
	// s is the same either way.
	s = Cyan(fmt.Sprintf("%X", bz[:n]))
	fmt.Printf("%s%s (", indent, s)
	if okI64 {
		fmt.Printf("i64:%v ", i64)
	}
	if okU64 {
		fmt.Printf("u64:%v", u64)
	}
	fmt.Print(")\n")
	return s, n, err
}

func scan8Byte(bz []byte, indent string) (s string, n int, err error) {
	if len(bz) < 8 {
		err = errors.New("while reading 8byte field, EOF was encountered")
		return
	}
	n = 8
	s = Blue(fmt.Sprintf("%X", bz[:8]))
	fmt.Printf("%s%s\n", indent, s)
	return
}

func scanByteLength(bz []byte, indent string) (s string, n int, err error) {
	// Read the length.
	l64, _n := binary.Uvarint(bz)
	if n < 0 {
		n = 0
		err = errors.New("error decoding uvarint")
		return
	}
	length := int(l64)
	if length >= len(bz) {
		err = errors.New("while reading 8byte field, EOF was encountered")
		return
	}
	lengthStrLong := fmt.Sprintf("%X (%v bytes) ", bz[:_n], length)
	lengthStrLongLen := len(lengthStrLong) // for indenting later
	lengthStrShort := fmt.Sprintf("%X", bz[:_n])
	s = Cyan(lengthStrShort)
	slide(&bz, &n, _n)
	// Read the remaining bytes.
	contents := bz[:length]
	contentsStr := fmt.Sprintf("%X", contents)
	s += Green(contentsStr)
	slide(&bz, &n, length)
	fmt.Printf("%s%s%s\n", indent, Cyan(lengthStrLong), Green(contentsStr))
	// If ascii string, also show the string in quotes.
	if amino.IsASCIIText(string(contents)) {
		fmt.Printf("%s%s\n", indent+strings.Repeat(" ", lengthStrLongLen), Green("("+strconv.Quote(string(contents))+" in ASCII)"))
	}
	return
}

func scanStruct(bz []byte, indent string, isRoot bool) (s string, n int, err error) {
	var (
		_s  string
		_n  int
		typ amino.Typ3
	)
	for {
		if isRoot && len(bz) == 0 {
			return
		}
		_s, typ, _n, err = scanFieldKey(bz, indent+"  ")
		if slide(&bz, &n, _n) && concat(&s, _s) && err != nil {
			return
		}
		_s, _n, err = scanAny(typ, bz, indent+"  ")
		if slide(&bz, &n, _n) && concat(&s, _s) && err != nil {
			return
		}
	}
}

func scanFieldKey(bz []byte, indent string) (s string, typ amino.Typ3, n int, err error) {
	var u64 uint64
	u64, n = binary.Uvarint(bz)
	if n < 0 {
		n = 0
		err = errors.New("error decoding uvarint")
		return
	}
	typ = amino.Typ3(u64 & 0x07)
	number := uint32(u64 >> 3)
	s = fmt.Sprintf("%X", bz[:n])
	fmt.Printf("%s%s @%v %v\n", indent, s, number, typ)
	return
}

func scan4Byte(bz []byte, indent string) (s string, n int, err error) {
	if len(bz) < 4 {
		err = errors.New("while reading 8byte field, EOF was encountered")
		return
	}
	n = 4
	s = Blue(fmt.Sprintf("%X", bz[:4]))
	fmt.Printf("%s%s\n", indent, s)
	return
}

/*
func scanList(bz []byte, indent string) (s string, n int, err error) {
	// Read element Typ4.
	if len(bz) < 1 {
		err = errors.New("EOF while reading list element typ4.")
		return
	}
	var typ = amino.Typ4(bz[0])
	if typ&0xF0 > 0 {
		err = errors.New("Invalid list element typ4 byte")
	}
	s = fmt.Sprintf("%X", bz[:1])
	if slide(&bz, &n, 1) && err != nil {
		return
	}
	// Read number of elements.
	var num, _n = uint64(0), int(0)
	num, _n = binary.Uvarint(bz)
	if _n < 0 {
		_n = 0
		err = errors.New("error decoding list length (uvarint)")
	}
	s += Cyan(fmt.Sprintf("%X", bz[:_n]))
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	fmt.Printf("%s%s of %v with %v items\n", indent, s, typ, num)
	// Read elements.
	var _s string
	for i := 0; i < int(num); i++ {
		// Maybe read nil byte.
		if typ&0x08 != 0 {
			if len(bz) == 0 {
				err = errors.New("EOF while reading list nil byte")
				return
			}
			var nb = bz[0]
			slide(&bz, &n, 1)
			switch nb {
			case 0x00:
				s += "00"
				fmt.Printf("%s00 (not nil)\n", indent)
			case 0x01:
				s += "01" // Is nil (NOTE: reverse logic)
				fmt.Printf("%s01 (is nil)\n", indent)
				continue
			default:
				err = fmt.Errorf("Unexpected nil pointer byte %X", nb)
				return
			}
		}
		// Read element.
		_s, _n, err = scanAny(typ.Typ3(), bz, indent+"  ")
		if slide(&bz, &n, _n) && concat(&s, _s) && err != nil {
			return
		}
	}
	return
}
*/

/*
func scanInterface(bz []byte, indent string) (s string, n int, err error) {
	db, hasDb, pb, typ, _, isNil, _n, err := amino.DecodeDisambPrefixBytes(bz)
	if slide(&bz, &n, _n) && err != nil {
		return
	}
	pb3 := pb
	if isNil {
		s = Magenta("0000")
	} else if hasDb {
		s = Magenta(fmt.Sprintf("%X%X", db.Bytes(), pb3.Bytes()))
	} else {
		s = Magenta(fmt.Sprintf("%X", pb3.Bytes()))
	}
	if isNil {
		fmt.Printf("%s%s (nil interface)\n", indent, s)
	} else if hasDb {
		fmt.Printf("%s%s (disamb: %X, prefix: %X, typ: %v)\n",
			indent, s, db.Bytes(), pb.Bytes(), typ)
	} else {
		fmt.Printf("%s%s (prefix: %X, typ: %v)\n",
			indent, s, pb.Bytes(), typ)
	}
	_s, _n, err := scanAny(typ, bz, indent)
	if slide(&bz, &n, _n) && concat(&s, _s) && err != nil {
		return
	}
	return
}
*/

//----------------------------------------
// Misc.

func slide(bzPtr *[]byte, n *int, _n int) bool {
	if len(*bzPtr) < _n {
		panic("eof")
	}
	*bzPtr = (*bzPtr)[_n:]
	*n += _n
	return true
}

func concat(sPtr *string, _s string) bool {
	*sPtr += _s
	return true
}

func hexDecode(s string) []byte {
	bz, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return bz
}
