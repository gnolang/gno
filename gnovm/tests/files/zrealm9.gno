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
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "1e2bb51e04e8a93c9d6354134d6bd3dda394aa91",
//     +            "Hash": "6580fe9f3ec745192bc1ee8e1c84c0e591ed8b36",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -78,7 +78,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "e525f7017dd65bfd915dcf24d2795bdf00c5558f",
//     +                "Hash": "6b6e372a0a1a59ad77045a2213c51a2d3fd60c20",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
