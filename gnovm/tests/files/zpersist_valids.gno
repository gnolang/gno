// PKGPATH: gno.land/r/demo/tests_test
package tests_test

var (
	// Native types
	abigint      bigint        = 16
	abool        bool          = true
	abyte        byte          = 0x16
	afloat32     float32       = 16.16
	afloat64     float64       = 16.16
	aint         int           = 16
	aint8        int8          = 16
	aint16       int16         = 16
	aint32       int32         = 16
	aint64       int64         = 16
	arune        rune          = 'a'
	astring      string        = "hello"
	astringslice []string      = []string{"a"}
	auint        uint          = 16
	auint8       uint8         = 16
	auint16      uint16        = 16
	auint32      uint32        = 16
	auint64      uint64        = 16
	ainterface   interface{}   = interface{}
	afunc        func() string = func() string { return "a" }

	// TODO:
	// amap         map[string]string = map[string]string{"a": "1"}

	// Not supported:
	// error
)

func init() {
	printVars("preinit ")
	aint *= 2
	afloat64 *= 2
	astring += "."
	astringslice = append(astringslice, "b")
	abool = false
	arune = 'b'
	afunc = func() string { return "b" }

	// TODO:
	// amap["b"] = "2"
	printVars("postinit")
}

func main() {
	printVars("premain ")
	aint *= 2
	afloat64 *= 2
	astring += "."
	astringslice = append(astringslice, "c")
	abool = true
	arune = 'c'
	afunc = func() string { return "c" }

	// TODO:
	// amap["c"] = "3"
	printVars("postmain")
}

func printVars(phase string) {
	println(phase,
		// variables
		abigint,
		abool,
		abyte,
		afloat32,
		afloat64,
		aint,
		aint8,
		aint16,
		aint32,
		aint64,
		arune,
		astring,
		astringslice,
		auint,
		auint8,
		auint16,
		auint32,
		auint64,
		ainterface,
		afunc(),
	)
}

// Output:
// preinit  16 true 22 16.16 16.16 16 16 16 16 16 97 hello slice[("a" string)] 16 16 16 16 16 typeval{interface{} (0xc000370cd0)} a
// postinit 16 false 22 16.16 32.32 32 16 16 16 16 98 hello. slice[("a" string),("b" string)] 16 16 16 16 16 typeval{interface{} (0xc000370cd0)} b
// premain  16 false 22 16.16 32.32 32 16 16 16 16 98 hello. slice[ref(0e4830f6b0cb49c615cab8df855d8dd0f7ebc58c:4)] 16 16 16 16 16 typeval{interface{} (0xc00007e870)} b
// postmain 16 true 22 16.16 64.64 64 16 16 16 16 99 hello.. slice[("a" string),("b" string),("c" string)] 16 16 16 16 16 typeval{interface{} (0xc00007e870)} c
