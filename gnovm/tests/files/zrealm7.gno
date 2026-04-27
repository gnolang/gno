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
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:14](619)={
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
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:14",
//         "LastObjectSize": "619",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13",
//         "RefCount": "1"
//     }
// }
// c[08ada09dee16d791fd406d629fe29bb0ed084a30:13](376)={
//     "ObjectInfo": {
//         "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13",
//         "LastObjectSize": "376",
//         "ModTime": "0",
//         "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "filetests/extern/timtadh/data_structures/tree/avl.AvlNode"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "709eba3056a061a866f833d88fdc06c8dcfdfd82",
//             "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:14"
//         }
//     }
// }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:12](135)=
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
//     +                    "Hash": "edf9c5a1ce4d6fcea83b33fbbb62ad0711f3940c",
//     +                    "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:13"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12",
//              "LastObjectSize": "619",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:10](134)=
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
//     @@ -48,7 +39,24 @@
//                      "@type": "/gno.PointerValue",
//                      "Base": {
//                          "@type": "/gno.RefValue",
//     -                    "Hash": "9640a0414ff5d70dc029395e44ffdccc3c59ab23",
//     +                    "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
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
//                          "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11"
//                      },
//                      "Index": "0",
//     @@ -59,7 +67,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "LastObjectSize": "748",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:11](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:11",
//              "LastObjectSize": "376",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "9da2b11c167d64c5e217e53f3579ee8d972594be",
//     +            "Hash": "ee406ec9441b8feeb4fc548c1df1f7ec190cb60c",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:12"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:8](-124)=
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
//     -                    "Hash": "fe19367377a1146980353a3dcd0c3de3c2de21ed",
//     -                    "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9"
//     -                },
//     -                "Index": "0",
//     -                "TV": null
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8",
//              "LastObjectSize": "746",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "RefCount": "1"
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:9](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9",
//              "LastObjectSize": "374",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "2e4e53cb297a58323ae4a6dc98facbbc1f07e765",
//     +            "Hash": "773ba6efa3609387bbf7a956b462b397523453bf",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:10"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:7](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "LastObjectSize": "373",
//     -        "ModTime": "0",
//     +        "ModTime": "12",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "e8a406fa3e95dfa7986ce2b867319aa193decd80",
//     +            "Hash": "29144c4c505d5d0ba099ae9ce0158beb875f1ff9",
//                  "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:8"
//              }
//          }
// u[08ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "08ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "423",
//     -        "ModTime": "6",
//     +        "ModTime": "12",
//              "OwnerID": "08ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -18,8 +18,8 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "2644d74bd7bdd3f9dec5bed61244ffdde5f2ac2b",
//     -                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:7"
//     +                "Hash": "346c2acf9ad4fefa5cfffb3ff29185809329e5e7",
//     +                "ObjectID": "08ada09dee16d791fd406d629fe29bb0ed084a30:9"
//                  },
//                  "Index": "0",
//                  "TV": null
