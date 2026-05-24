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
	ms.Field += 2
	println(ms)
}

// Output:
// (struct{(1 int)} gno.land/r/test.MyStruct)
// (struct{(3 int)} gno.land/r/test.MyStruct)

// Realm:
// finalizerealm["gno.land/r/test"]
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](0)=
//     @@ -1,7 +1,7 @@
//      {
//          "Fields": [
//              {
//     -            "N": "AQAAAAAAAAA=",
//     +            "N": "AwAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "c664203dc898067434f72b8bc41892d443fe61c9",
//     +            "Hash": "c432c83ed71fe74b88a6fea9ccdb3457b3a2fecc",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "80ca4df359937a9e285bff6980fb0f8eaacce37e",
//     +                "Hash": "b982f0e71d2ac16df7b64342fb491d0992ea84dc",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
