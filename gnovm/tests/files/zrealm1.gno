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

func main(cur realm,) {
	key := "somekey"
	root = InnerNode{
		Key:   key,
		Left:  nil,
		Right: nil,
	}
}

// Realm:
// finalizerealm["gno.land/r/test"]
// c[a8ada09dee16d791fd406d629fe29bb0ed084a30:6:0](273)={
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
//         "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6:0",
//         "LastObjectSize": "273",
//         "ModTime": "0",
//         "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//         "RefCount": "1"
//     }
// }
// u[a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0](158)=
//     @@ -2,9 +2,19 @@
//          "ObjectInfo": {
//              "ID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:3:0",
//              "LastObjectSize": "194",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "OwnerID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:2:0",
//              "RefCount": "1"
//          },
//     -    "Value": {}
//     +    "Value": {
//     +        "T": {
//     +            "@type": "/gno.RefType",
//     +            "ID": "gno.land/r/test.InnerNode"
//     +        },
//     +        "V": {
//     +            "@type": "/gno.RefValue",
//     +            "Hash": "4c5dd933e52485e4eec50197b43e6f4ade57bcad",
//     +            "ObjectID": "a8ada09dee16d791fd406d629fe29bb0ed084a30:6:0"
//     +        }
//     +    }
//      }
