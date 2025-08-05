// PKGPATH: gno.land/r/test
package test

type MyStruct struct {
	Field int
}

var ms MyStruct

func init() {
	ms.Field = 1
}

func main(cur realm,) {
	println(ms)
	ms.Field--
	println(ms)
}

// Output:
// (struct{(1 int)} gno.land/r/test.MyStruct)
// (struct{(0 int)} gno.land/r/test.MyStruct)

// Realm:
// finalizerealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4](0)=
//     @@ -1,7 +1,6 @@
//      {
//          "Fields": [
//              {
//     -            "N": "AQAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
