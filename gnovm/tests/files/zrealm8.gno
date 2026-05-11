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
	ms.Field++
	println(ms)
}

// Output:
// (struct{(1 int)} gno.land/r/test.MyStruct)
// (struct{(2 int)} gno.land/r/test.MyStruct)

// Realm:
// finalizerealm["gno.land/r/test"]
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](0)=
//     @@ -1,7 +1,7 @@
//      {
//          "Fields": [
//              {
//     -            "N": "AQAAAAAAAAA=",
//     +            "N": "AgAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "c664203dc898067434f72b8bc41892d443fe61c9",
//     +            "Hash": "bcc5900554f446403bf1b3970c8ed9d5f65b3147",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "80ca4df359937a9e285bff6980fb0f8eaacce37e",
//     +                "Hash": "ca3f923877f5a0d1e5641086e17112a93d2b1fc2",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
