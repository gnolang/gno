// PKGPATH: gno.land/r/test
package test

import (
	"filetests/extern/timtadh/data_structures/tree/avl"
	"filetests/extern/timtadh/data_structures/types"
)

var tree *avl.AvlNode

func init() {
	tree, _ = tree.Put(types.String("key0"), "value0")
	tree, _ = tree.Put(types.String("key1"), "value1")
}

func main(cur realm) {
	var updated bool
	tree, updated = tree.Put(types.String("key3"), "value3")
	println(updated, tree.Size())
}

// Output:
// false 3

// Realm:
// finalizerealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "filetests/extern/timtadh/data_structures/types.String"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "key3"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "value3"
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0]={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "5e7b9845910a25188582e6bd989d50241bd4f39f",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0]=
//     @@ -21,7 +21,7 @@
//                  }
//              },
//              {
//     -            "N": "AQAAAAAAAAA=",
//     +            "N": "AgAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -43,12 +43,22 @@
//                          "@type": "/gno.RefType",
//                          "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//                      }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "3c8e7c05c6c819870255dc02852145ff2d480468",
//     +                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0",
//              "RefCount": "1"
//          },
//     @@ -12,7 +12,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "392bb3dc28984a542210435ce0873cb79a51e4a8",
//     +            "Hash": "f77441ce4a46a191400a7508042967a249b1f9ab",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0]=
//     @@ -21,7 +21,7 @@
//                  }
//              },
//              {
//     -            "N": "AgAAAAAAAAA=",
//     +            "N": "AwAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -48,7 +48,7 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "28cca1ddf37936deb6fef6ee14256644bcac654a",
//     +                    "Hash": "92ff5274c38bd835e3075c15017be47630bdc191",
//                          "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0"
//                      },
//                      "Index": "0",
//     @@ -58,7 +58,7 @@
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0",
//     -        "ModTime": "0",
//     +        "ModTime": "10",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//              "RefCount": "1"
//          },
//     @@ -12,7 +12,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "ccca9aff0db98fc3b4a89ae81c9102342b863686",
//     +            "Hash": "6ee464e298739ab5e1510185066f83250586bb6d",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//     -        "ModTime": "6",
//     +        "ModTime": "10",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2:0",
//              "RefCount": "1"
//          },
//     @@ -17,7 +17,7 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "2cbe6b5982d49dfbbc01ac43fa4e10edc681c013",
//     +                "Hash": "906751d347dd9e7c5627cd7fe65eafc80fc9aaf7",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0"
//                  },
//                  "Index": "0",
