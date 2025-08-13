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
	tree, _ = tree.Put(types.String("key2"), "value2")
}

func main(cur realm) {
	var updated bool
	tree, updated = tree.Put(types.String("key3"), "value3")
	println(updated, tree.Size())
}

// Output:
// false 4

// Realm:
// finalizerealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:8](617)={
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "LastObjectSize": "617",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:7](374)={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "LastObjectSize": "374",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "698dea5e21f51fba85a5d915d52be1548d9a3cfe",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:12](134)=
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
//     +                    "Hash": "bab20b85fe31a602308d2c17efda653a17707ce1",
//     +                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12",
//              "LastObjectSize": "619",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:8](-124)=
//     @@ -7,7 +7,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "key3"
//     +                "value": "key0"
//                  }
//              },
//              {
//     @@ -17,7 +17,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "value3"
//     +                "value": "value0"
//                  }
//              },
//              {
//     @@ -48,8 +48,8 @@
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//     -        "LastObjectSize": "617",
//     -        "ModTime": "0",
//     +        "LastObjectSize": "746",
//     +        "ModTime": "6",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7](5)=
//     @@ -1,9 +1,9 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//     -        "LastObjectSize": "374",
//     -        "ModTime": "0",
//     -        "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12",
//     +        "LastObjectSize": "373",
//     +        "ModTime": "6",
//     +        "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          },
//          "Value": {
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "698dea5e21f51fba85a5d915d52be1548d9a3cfe",
//     +            "Hash": "752b3063c9f667045099a7908edd5da38cc50d5e",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:11](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11",
//              "LastObjectSize": "376",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "4a17484cc93c655365dd3aca8dd7b71f98efefc6",
//     +            "Hash": "cf890b5dcb3fcb9e94c5f07a807d9c6de32c4d0d",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:10](134)=
//     @@ -21,19 +21,10 @@
//                  }
//              },
//              {
//     -            "N": "AgAAAAAAAAA=",
//     +            "N": "AwAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     -            }
//     -        },
//     -        {
//     -            "T": {
//     -                "@type": "/gno.PointerType",
//     -                "Elt": {
//     -                    "@type": "/gno.RefType",
//     -                    "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//     -                }
//                  }
//              },
//              {
//     @@ -48,7 +39,26 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "db040c58097327c7918fdc57b6cec8f49230814b",
//     +                    "Hash": "22396e8e5f3e9bcb19ebf9370ebcebd11fd73cb9",
//     +                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//     +            }
//     +        },
//     +        {
//     +            "T": {
//     +                "@type": "/gno.PointerType",
//     +                "Elt": {
//     +                    "@type": "/gno.RefType",
//     +                    "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//     +                }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "6a869ff561705e0f8e1ef67b8f86e444c039c751",
//                          "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11"
//                      },
//                      "Index": "0",
//     @@ -59,7 +69,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "LastObjectSize": "748",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:9](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9",
//              "LastObjectSize": "374",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "b65adb52a839f07bf1db12b5889d4bff183f5134",
//     +            "Hash": "20f35f84225e8d6370f61fd377d6cfee6dbb537f",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -18,8 +18,8 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "b171dea0cd981edc859f2b1dea69ad0137f1c227",
//     -                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//     +                "Hash": "0bb9c8b4a7a1e8e57578d5719b5ec9342c8f6230",
//     +                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9"
//                  },
//                  "Index": "0",
//                  "TV": null
