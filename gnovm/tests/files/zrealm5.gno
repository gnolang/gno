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
//     -                "Hash": "9ca98fdf0502999046966e61b84ad729e39753e7",
//     +                "Hash": "d3937c5ccc0342c231c56fa310e2b541f2fcd38e",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                  },
//                  "Index": "0",
