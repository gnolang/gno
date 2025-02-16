// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"gno.land/p/demo/tests"
)

var root interface{}

func main() {
	root = tests.F
	println(root)
}

// Output:
// (struct{("p" string)} gno.land/p/demo/tests.Foo)

// Realm:
// switchrealm["gno.land/r/crossrealm_test"]
// c[f5a516808f8976c33939133293d598ce3bca4e8d:4]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "p"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:4",
//         "ModTime": "0",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "1"
//     }
// }
