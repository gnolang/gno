// PKGPATH: gno.land/r/test
package test

type S struct {
	A int
}

var a, b *S

func main(cur realm,) {
	a = new(S)
	a.A = 4
	b = a
}

// Realm:
// finalizerealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0](218)={
//     "Fields": [
//         {
//             "N": "BAAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0",
//         "LastObjectSize": "218",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0](341)={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0",
//         "IsEscaped": true,
//         "LastObjectSize": "341",
//         "ModTime": "0",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test.S"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "f2c9e51f8699c22c208df322c315d75b2ea40f9f",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0](139)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//              "LastObjectSize": "253",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2:0",
//              "RefCount": "1"
//          },
//     @@ -13,6 +13,16 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/test.S"
//                  }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Escaped": true,
//     +                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:4:0](139)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4:0",
//              "LastObjectSize": "253",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2:0",
//              "RefCount": "1"
//          },
//     @@ -13,6 +13,16 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/test.S"
//                  }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Escaped": true,
//     +                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
