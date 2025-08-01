// PKGPATH: gno.land/r/test
package test

type (
	word uint
	nat  []word
)

var zero *Int

// structLit
type Int struct {
	abs nat
}

func main() {
	zero = &Int{
		abs: []word{0},
	}
	a := zero.abs
	println(a)
}

// Output:
// (slice[(0 gno.land/r/test.word)] gno.land/r/test.nat)

// Realm:
// finalizerealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0](231)={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/test.word"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0",
//         "LastObjectSize": "231",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0](365)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/test.nat"
//             },
//             "V": {
//                 "@type": "/gno.SliceValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "88c0cba9197500e01ba3989ea2aae6ef2aaa1605",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0"
//                 },
//                 "Length": "1",
//                 "Maxcap": "1",
//                 "Offset": "0"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0",
//         "LastObjectSize": "365",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6:0",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:6:0](341)={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6:0",
//         "LastObjectSize": "341",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test.Int"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "0c90b3442e9e34fe751eb9f7b84286b6eec80183",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0](137)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//              "LastObjectSize": "255",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2:0",
//              "RefCount": "1"
//          },
//     @@ -13,6 +13,16 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/test.Int"
//                  }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "3962c15ecbadfbf1c5c47b70209ea69bc0193ace",
//     +                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6:0"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
