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

func init() {
	root = InnerNode{
		Key: "old",
	}
}

func main() {
	crossing()

	root = InnerNode{
		Key: "new",
	}
}

// Realm:
// switchrealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:8]={
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
//         {},
//         {}
//     ],
//     "ObjectInfo": {
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//         "RefCount": "1"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3",
//     -        "ModTime": "6",
//     +        "ModTime": "7",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2",
//              "RefCount": "1"
//          },
//     @@ -12,8 +12,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "c1bfc6b3d7043721364563a780c15c757c10a49f",
//     -            "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:7"
//     +            "Hash": "17c8ef092e704bbac8183328f9295793011c8364",
//     +            "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:8"
//              }
//          }
//      }
// d[a8ada09dee16d791fd406d629fe29bb0ed084a30:7]
