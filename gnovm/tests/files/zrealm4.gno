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
	tree, updated = tree.Put(types.String("key0"), "value0-new")
	println(updated, tree.Size())
}

// Output:
// true 1

// Realm:
// finalizerealm["gno.land/r/test"]
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:8](9)=
//     @@ -17,7 +17,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.StringValue",
//     -                "value": "value0"
//     +                "value": "value0-new"
//                  }
//              },
//              {
//     @@ -49,7 +49,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//              "LastObjectSize": "617",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "RefCount": "1"
//          }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "LastObjectSize": "423",
//     -        "ModTime": "6",
//     +        "ModTime": "8",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,6 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "9ca98fdf0502999046966e61b84ad729e39753e7",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                  },
//                  "Index": "0",
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "LastObjectSize": "373",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "ff7d2f7d23827a0af50e33994bc896d1b1ab3ea8",
//     +            "Hash": "6e91c2c1ddacaac0d8a29bbbfa6816187d7a491c",
//                  "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//              }
//          }
