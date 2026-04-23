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
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:8](214)={
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
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "LastObjectSize": "214",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:7](335)={
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "IsEscaped": true,
//         "LastObjectSize": "335",
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
//             "Hash": "a280d74394f8a508ba2efa2b960833f6ad7db04a",
//             "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8"
//         }
//     }
// }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](134)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "249",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -13,6 +13,15 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/test.S"
//                  }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](134)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "249",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -13,6 +13,15 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/test.S"
//                  }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//              "LastObjectSize": "762",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "f7e2c3928a8380d31ebef5b75d4e4d314d58b499",
//     +                "Hash": "176bd8a5f2b6f553ae22a77f5e98a40934a26735",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
//     @@ -56,7 +56,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "81e0c6ca227ff564c90d2b560c954fe5543c9769",
//     +                "Hash": "3900e5c2ef5f3655096096f14eac85f84d118bb8",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4"
//                  }
//              },
