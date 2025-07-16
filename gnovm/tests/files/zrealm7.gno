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
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:14:0](623)={
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14:0",
//         "LastObjectSize": "623",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:13:0",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:13:0](382)={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:13:0",
//         "LastObjectSize": "382",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "36f5d19ee3845b0886a21c67ec83c6d1d07e0d09",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:14:0"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0](138)=
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
//     +                    "Hash": "6afb0b585ba30d119f5bc459011715094952c5a6",
//     +                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:13:0"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0",
//              "LastObjectSize": "623",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0](-127)=
//     @@ -21,7 +21,7 @@
//                  }
//              },
//              {
//     -            "N": "AwAAAAAAAAA=",
//     +            "N": "AQAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -43,23 +43,13 @@
//                          "@type": "/gno.RefType",
//                          "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//                      }
//     -            },
//     -            "V": {
//     -                "@type": "/gno.PointerValue",
//     -                "Base": {
//     -                    "@type": "/gno.RefValue",
//     -                    "Hash": "0ddb364eb4d98f0e18d7f7e0b9efc3cd02bdda64",
//     -                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0",
//              "LastObjectSize": "753",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0",
//              "LastObjectSize": "379",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "b2a17e1cad945d3e331ce58d3737d036202ff9d1",
//     +            "Hash": "83afa039d38297249375eb5840f46c0998a39df7",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0",
//              "LastObjectSize": "382",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "03fd580f124e54193e01cb1e2715698851759c98",
//     +            "Hash": "739be84d014d3ed0acc6208530798b9d02c391cc",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:12:0"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0](137)=
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
//     -                    "Hash": "4d3b43fa652fdf75ba644c4757134eb79c05bcd1",
//     +                    "Hash": "d2ce70a0dcd8fd0aa8d74c494989ec1025d0a40a",
//     +                    "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0"
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
//     +                    "Hash": "584a7646b24ecc4c2a4c4634be95180c0cd8f265",
//                          "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:11:0"
//                      },
//                      "Index": "0",
//     @@ -59,7 +69,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0",
//              "LastObjectSize": "755",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0",
//              "LastObjectSize": "380",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8:0",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "643a5ff59f9a9703fef13311ba3785492f498542",
//     +            "Hash": "2f3f6f8e4ef29fdf0b9244b7b31cdb9024c6dd69",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:10:0"
//              }
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//              "LastObjectSize": "430",
//     -        "ModTime": "6",
//     +        "ModTime": "12",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2:0",
//              "RefCount": "1"
//          },
//     @@ -18,8 +18,8 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "776d05d6e61327b4ce18639c5db788289dc32e43",
//     -                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7:0"
//     +                "Hash": "b989c3b6812e4a3be71ed8e1d600479aff16b842",
//     +                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:9:0"
//                  },
//                  "Index": "0",
//                  "TV": null
