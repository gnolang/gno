// PKGPATH: gno.land/r/test
package test

// Pins the persisted store representation of a lazy interface bind: the
// BoundMethodValue is stored with Func null and the dispatch context
// (Method + MethodPkg) as data, resolved on each call.

type Impl struct{ x int }

func (i Impl) Get() int { return i.x }

type IG interface{ Get() int }

var G func() int

func main(cur realm) {
	var o IG = Impl{x: 7}
	G = o.Get
	println(G())
}

// Output:
// 7

// Realm:
// finalizerealm["gno.land/r/test"]
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:7](214)={
//     "Fields": [
//         {
//             "N": "BwAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "LastObjectSize": "214",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:6](361)={
//     "Func": null,
//     "Method": "Get",
//     "MethodPkg": "gno.land/r/test",
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6",
//         "LastObjectSize": "361",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "RefCount": "1"
//     },
//     "Receiver": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test.Impl"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "cf9fac1f1c39731a1e061643ebc8ceed19ab5402",
//             "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
//         }
//     }
// }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](111)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "245",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -22,6 +22,11 @@
//                          }
//                      }
//                  ]
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "970035c55b546a7567b5cfb8a8022ac9b67b8c4b",
//     +            "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:6"
//              }
//          }
//      }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "IsEscaped": true,
//              "LastObjectSize": "713",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -58,7 +58,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "92eaa028fb084aaaf6e5056f61419f9704f272d0",
//     +                "Hash": "ecaa39fc6a8701586fdc80dc77a093e201b87f68",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3"
//                  }
//              },
