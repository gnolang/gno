// PKGPATH: gno.land/r/test
package test

var root *Node

type Key any

type Node struct {
	Key   Key
	Left  *Node
	Right *Node
}

func init() {
	root = &Node{
		Key: "old",
	}
}

func main(cur realm,) {
	root = &Node{
		Key: "new",
	}
}

// Realm:
// finalizerealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:8](389)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "new"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/test.Node"
//                 }
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/test.Node"
//                 }
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "LastObjectSize": "389",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "RefCount": "1"
//     }
// }
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:7](336)={
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7",
//         "LastObjectSize": "336",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/test.Node"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "c7d7bfccb4288546e693f18963ec7c3e67efa8de",
//             "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//         }
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3](0)=
//     @@ -18,7 +18,7 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "477e1546db4d3a70091c73a992039811bc2666fe",
//     +                "Hash": "23c20ff754ccaa3772c628a7b2a912cf3175541f",
//                      "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//                  },
//                  "Index": "0",
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:7](-336)
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:8](-389)
