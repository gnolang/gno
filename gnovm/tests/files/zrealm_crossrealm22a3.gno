// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/func"
)

var b = 1

type Local_S struct {
	name string
}

var ls = &Local_S{name: "Local_S"}

func main() {
	f := func() bool {
		c := b
		return true
	}
	crossrealm.SetCallback(f)

	ls.name = "modified"

	println(ls)
}

// Output:
// &(struct{("modified" string)} gno.land/r/crossrealm_test.Local_S)

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm/func"]
// switchrealm["gno.land/r/crossrealm_test"]
// u[r:f5a516808f8976c33939133293d598ce3bca4e8d:4]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "modified"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:4",
//         "ModTime": "5",
//         "OwnerID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "RefCount": "1"
//     }
// }
