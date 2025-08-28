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
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:8](227)={
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "LastObjectSize": "227",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:7](359)={
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
//                     "Hash": "053ebe7d3e2087ff390f1c09b3f36cf0763f0967",
//                     "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//                 },
//                 "Length": "1",
//                 "Maxcap": "1",
//                 "Offset": "0"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "LastObjectSize": "359",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:6](335)={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6",
//         "LastObjectSize": "335",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test.Int"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "e757ea3d88983d3fc397e089882a1e31ee2c5e10",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3](134)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "251",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
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
//     +                "Hash": "afc8a8a4c127ea7b6713ec59220d7c6cdd6e842e",
//     +                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
