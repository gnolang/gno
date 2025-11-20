// PKGPATH: gno.land/r/tests
package tests

type myStruct struct {
	A int
	B string
}

var (
	// Native types
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
	astringslice []string      = []string{"A"}
	astruct      myStruct      = myStruct{16, "A"}
	auint        uint          = 16
	auint8       uint8         = 16
	auint16      uint16        = 16
	auint32      uint32        = 16
	auint64      uint64        = 16
	ainterface   any           = struct{ a float32 }{16.0}
	afunc        func() string = func() string { return "A" }

	// TODO:
	// amap         map[string]string = map[string]string{"a": "1"}
	// more slices
	// slice/map containing unsupported types

	// Not supported:
	// error
)

func init() {
	printVars("preinit ")
	mutateVars("B")
	printVars("postinit")
}

func main(cur realm) {
	printVars("premain ")
	mutateVars("C")
	printVars("postmain")
}

func mutateVars(stringModifier string) {
	abool = !abool
	abyte *= 2
	afloat32 *= 2
	afloat64 *= 2
	aint *= 2
	aint8 *= 2
	aint16 *= 2
	aint32 *= 2
	aint64 *= 2
	arune = rune(stringModifier[0])
	astring += stringModifier
	astringslice = append(astringslice, stringModifier)
	auint *= 2
	auint8 *= 2
	auint16 *= 2
	auint32 *= 2
	auint64 *= 2
	astruct.A *= 2
	astruct.B = stringModifier
	ainterface = struct{ a string }{stringModifier}
	afunc = func() string { return stringModifier }

	// TODO:
	// amap["c"] = "3"
}

func printVars(phase string) {
	println(phase,
		// variables
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
		astringslice[0],
		astruct,
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
// preinit  true 22 16.16 16.16 16 16 16 16 16 97 hello slice[("A" string)] A (struct{(16 int),("A" string)} gno.land/r/tests.myStruct) 16 16 16 16 16 struct{(16 float32)} A
// postinit false 44 32.32 32.32 32 32 32 32 32 66 helloB slice[("A" string),("B" string)] A (struct{(32 int),("B" string)} gno.land/r/tests.myStruct) 32 32 32 32 32 struct{("B" string)} B
// premain  false 44 32.32 32.32 32 32 32 32 32 66 helloB slice[("A" string),("B" string)] A (struct{(32 int),("B" string)} gno.land/r/tests.myStruct) 32 32 32 32 32 struct{("B" string)} B
// postmain true 88 64.64 64.64 64 64 64 64 64 67 helloBC slice[("A" string),("B" string),("C" string)] A (struct{(64 int),("C" string)} gno.land/r/tests.myStruct) 64 64 64 64 64 struct{("C" string)} C
