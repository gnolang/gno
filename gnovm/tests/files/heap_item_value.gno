// PKGPATH: gno.land/r/test
package test

type S struct {
	A int
}

var a, b *S

func main() {
	a = new(S)
	a.A = 4
	b = a
}

// Realm:
// switchrealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:5]={
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:4]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "IsEscaped": true,
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
//             "Hash": "8c001dde13b1f4dc01fc6d3a5bb4bc0cdfe2a50b",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:5"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2]=
//     --- old
//     +++ new
//     @@ -54,6 +54,16 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/r/test.S"
//                      }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Escaped": true,
//     +                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              },
//              {
//     @@ -63,6 +73,16 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/r/test.S"
//                      }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Escaped": true,
//     +                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              },
//              {
