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
// c[r:f5a516808f8976c33939133293d598ce3bca4e8d:5]={
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
//         "ID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:5",
//         "ModTime": "0",
//         "OwnerID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:4",
//         "RefCount": "1"
//     }
// }
// c[r:f5a516808f8976c33939133293d598ce3bca4e8d:4]={
//     "ObjectInfo": {
//         "ID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:4",
//         "ModTime": "0",
//         "OwnerID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/p/demo/tests/p_crossrealm.Container"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "b7ba32ba8f6bd39fe6f6fbc73e7668d5e27ac3fe",
//             "ObjectID": "r:f5a516808f8976c33939133293d598ce3bca4e8d:5"
//         }
//     }
// }
