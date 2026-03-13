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
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4](0)=
//     @@ -1,7 +1,7 @@
//      {
//          "Fields": [
//              {
//     -            "N": "AQAAAAAAAAA=",
//     +            "N": "AwAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "1e2bb51e04e8a93c9d6354134d6bd3dda394aa91",
//     +            "Hash": "9b46018b460926de117279798c385a6aa3310402",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2](0)=
//     @@ -78,7 +78,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "e525f7017dd65bfd915dcf24d2795bdf00c5558f",
//     +                "Hash": "fe00d79b2b28d0cb2509a562b97c8fb03c518aaf",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
