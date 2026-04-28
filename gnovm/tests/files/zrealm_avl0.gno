// PKGPATH: gno.land/r/test
package test

import (
	"gno.land/p/nt/avl/v0"
)

var node *avl.Node

func init() {
	node = avl.NewNode("key0", "value0")
	// node, _ = node.Set("key0", "value0")
}

func main(cur realm) {
	var updated bool
	node, updated = node.Set("key1", "value1")
	// println(node, updated)
	println(updated, node.Size())
}

// Output:
// false 2

// Realm:
// finalizerealm["gno.land/r/test"]
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:7](6)=
//     @@ -2,8 +2,8 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "LastObjectSize": "341",
//     -        "ModTime": "0",
//     -        "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//     +        "ModTime": "10",
//     +        "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "RefCount": "1"
//          },
//          "Value": {
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:12](547)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "key1"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "value1"
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
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12",
//         "LastObjectSize": "547",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:11](344)={
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11",
//         "LastObjectSize": "344",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl/v0.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "d89a4f41f3ed19bc57ce93b4a99a1afb308b8d56",
//             "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12"
//         }
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:10](749)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "key1"
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
//                     "Hash": "44b12d4d6d199e7891ef2f767d2feedb3582de96",
//                     "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
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
//                     "Hash": "b66d1b196b2bc5a98fb8fc4eed58e095d8f49471",
//                     "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "LastObjectSize": "749",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:9](342)={
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "LastObjectSize": "342",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/nt/avl/v0.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "15172b3190ceaef48ccddb466a59dc0abc3d3dd2",
//             "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10"
//         }
//     }
// }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "391",
//     -        "ModTime": "6",
//     +        "ModTime": "8",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -18,8 +18,8 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "f6f6fde72f1d8a2c89a1aca5c06faa5e023a4222",
//     -                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
//     +                "Hash": "475807f87052e23bbaaed5cfc5d25edda93dd80f",
//     +                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9"
//                  },
//                  "Index": "0",
//                  "TV": null
