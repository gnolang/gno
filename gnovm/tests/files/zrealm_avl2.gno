// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/nt/avl/v0"
)

var tree avl.Tree

func init() {
	tree.Set("key1", "value1")
}

func main(cur realm) {
	var updated bool
	updated = tree.Set("key2", "value2")
	// println(tree, updated)
	println(updated, tree.Size())
}

// Output:
// false 2

// Realm:
// finalizerealm["gno.land/r/test"]
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:8](6)=
//     @@ -2,8 +2,8 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8",
//              "LastObjectSize": "341",
//     -        "ModTime": "0",
//     -        "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//     +        "ModTime": "11",
//     +        "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11",
//              "RefCount": "1"
//          },
//          "Value": {
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:13](547)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "key2"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "value2"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl/v0.Node"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl/v0.Node"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13",
//         "LastObjectSize": "547",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:12](344)={
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12",
//         "LastObjectSize": "344",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl/v0.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "f12372559c7846f4b035c6970b7cd62d4b1740c6",
//             "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13"
//         }
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:11](750)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "key2"
//             }
//         },
//         {},
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "64"
//             }
//         },
//         {
//             "N": "AgAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl/v0.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "81a22abf9449036700a2307934da1ebd4b0bea90",
//                     "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/p/nt/avl/v0.Node"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "508dd833fa0beceda7b1781ad8ad2d89820a31c7",
//                     "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11",
//         "LastObjectSize": "750",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:10](343)={
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "LastObjectSize": "343",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl/v0.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "d6f122b735bb8a199491f9b645e9ce8708d2c952",
//             "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11"
//         }
//     }
// }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:4](1)=
//     @@ -12,8 +12,8 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "3e9df728b22ef4b60e7abb81593f72e936a948e9",
//     -                    "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8"
//     +                    "Hash": "353a1ea31d40e8baab45cbf5104e066f22424ba5",
//     +                    "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10"
//                      },
//                      "Index": "0",
//                      "TV": null
//     @@ -23,7 +23,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:4",
//              "LastObjectSize": "389",
//     -        "ModTime": "7",
//     +        "ModTime": "9",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          }
