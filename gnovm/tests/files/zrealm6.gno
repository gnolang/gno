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
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:7](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//              "LastObjectSize": "373",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//              "RefCount": "1"
//          },
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -18,7 +18,7 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "4831ee01b39c1810a4c0047498b1d8628ea9e3e1",
//     +                "Hash": "e66e0fe0de1a5410fe5e34187607591888af3047",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                  },
//                  "Index": "0",
