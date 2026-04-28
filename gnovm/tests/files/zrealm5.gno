// PKGPATH: gno.land/r/test
package test

import (
	"filetests/extern/timtadh/data_structures/tree/avl"
	"filetests/extern/timtadh/data_structures/types"
)

var tree *avl.AvlNode

func init() {
	tree, _ = tree.Put(types.String("key0"), "value0")
}

func main(cur realm) {
	var updated bool
	tree, updated = tree.Put(types.String("key1"), "value1")
	println(updated, tree.Size())
}

// Output:
// false 2

// Realm:
// finalizerealm["gno.land/r/test"]
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:10](618)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "filetests/extern/timtadh/data_structures/types.String"
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
//                     "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//         "LastObjectSize": "618",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:9](374)={
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9",
//         "LastObjectSize": "374",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "3a52a8ad68d61b5eac06aca00f259bdfd6ca3037",
//             "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10"
//         }
//     }
// }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:8](134)=
//     @@ -21,7 +21,7 @@
//                  }
//              },
//              {
//     -            "N": "AQAAAAAAAAA=",
//     +            "N": "AgAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -43,13 +43,23 @@
//                          "@type": "/gno.RefType",
//                          "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//                      }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "0e553559c1e64c66cc5aed8c3ecfbefed2dc7cc6",
//     +                    "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8",
//              "LastObjectSize": "617",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "423",
//     -        "ModTime": "6",
//     +        "ModTime": "8",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,6 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "14081ad8498d444f42f2ec4fa8b1c6b46d043817",
//                      "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                  },
//                  "Index": "0",
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:7](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "LastObjectSize": "373",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "3b944e0ac2a76fef67c2fd3383dfb971447bdc42",
//     +            "Hash": "d85ee2e77889a2ec8ebdad2cb37c46320bdcb3d6",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8"
//              }
//          }
