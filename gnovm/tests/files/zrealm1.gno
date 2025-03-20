// PKGPATH: gno.land/r/test
package test

var root Node

type (
	Node any
	Key  any
)

type InnerNode struct {
	Key   Key
	Left  Node
	Right Node
}

func main() {
	key := "somekey"
	root = InnerNode{
		Key:   key,
		Left:  nil,
		Right: nil,
	}
}

// Realm:
// switchrealm["gno.land/r/test"]
// c[r:a8ada09dee16d791fd406d629fe29bb0ed084a30:4]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "somekey"
//             }
//         },
//         {},
//         {}
//     ],
//     "ObjectInfo": {
//         "ID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "ModTime": "0",
//         "OwnerID": "r:a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "RefCount": "1"
//     }
// }
