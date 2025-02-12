// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	"std"

	"gno.land/r/demo/tests/crossrealm"
	"gno.land/r/demo/tests/crossrealm_b"
)

func main() {
	f := crossrealm_b.Fooer
	crossrealm.SetFooer(f)
	crossrealm.CallFooerFoo()
	f.SetS("B")
	crossrealm.CallFooerFoo()
	print(".")
}

// Output:
// hello A cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// hello B cur=gno.land/r/demo/tests/crossrealm_b prev=gno.land/r/demo/tests/crossrealm
// .

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// u[0edc46caf30c00efd87b6c272673239eafbd051e:4]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "B"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0edc46caf30c00efd87b6c272673239eafbd051e:4",
//         "ModTime": "5",
//         "OwnerID": "0edc46caf30c00efd87b6c272673239eafbd051e:3",
//         "RefCount": "1"
//     }
// }
// switchrealm["gno.land/r/demo/tests/crossrealm_b"]
// switchrealm["gno.land/r/demo/tests/crossrealm"]
// switchrealm["gno.land/r/crossrealm_test"]

// Error:
//
