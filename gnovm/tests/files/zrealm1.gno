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
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:4]={
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//         "RefCount": "1"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:2]=
//     @@ -18,7 +18,17 @@
//              }
//          },
//          "Values": [
//     -        {},
//     +        {
//     +            "T": {
//     +                "@type": "/gno.RefType",
//     +                "ID": "gno.land/r/test.InnerNode"
//     +            },
//     +            "V": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "ae4e9e2d205cc0081d4ee249e1d188ebe270b220",
//     +                "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:4"
//     +            }
//     +        },
//              {
//                  "T": {
//                      "@type": "/gno.TypeType"
