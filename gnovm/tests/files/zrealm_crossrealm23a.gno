// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import "gno.land/p/demo/tests/p_crossrealm"

var b0 *p_crossrealm.Container

func main() {
	b0 = &p_crossrealm.Container{
		A: 1,
	}
	print(".")
}

// Output:
// .

// Realm:
// switchrealm["gno.land/r/crossrealm_test"]
// c[f5a516808f8976c33939133293d598ce3bca4e8d:5]={
//     "Fields": [
//         {
//             "N": "AQAAAAAAAAA=",
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "32"
//             }
//         },
//         {}
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:5",
//         "ModTime": "0",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:4",
//         "RefCount": "1"
//     }
// }
// c[f5a516808f8976c33939133293d598ce3bca4e8d:4]={
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:4",
//         "ModTime": "0",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/tests/p_crossrealm.Container"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "01799eb680e2f34495c1050457f9ccd79795a561",
//             "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:5"
//         }
//     }
// }
