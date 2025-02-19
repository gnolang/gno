// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

type foo struct {
	name string
}

func (*foo) String() string { return "hey" }

var f = &foo{"foo"}

var arr interface{}

func main() {
	arr = crossrealm.GetSlice8(f) // real array value from external realm
	println(".")
}

// Output:
// .

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm/slice"]
// u[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:14]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/crossrealm_test.foo"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:14",
//         "ModTime": "28",
//         "OwnerID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:2",
//         "RefCount": "1"
//     }
// }
// u[f5a516808f8976c33939133293d598ce3bca4e8d:3]={
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "IsEscaped": true,
//         "ModTime": "28",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "de578df5354e812a94f64e23eed06e9705d2b4ef",
//             "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:4"
//         }
//     }
// }
// switchrealm["gno.land/r/crossrealm_test"]
// c[f5a516808f8976c33939133293d598ce3bca4e8d:6]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/crossrealm_test.foo"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Escaped": true,
//                     "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:3"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:6",
//         "ModTime": "0",
//         "OwnerID": "f5a516808f8976c33939133293d598ce3bca4e8d:2",
//         "RefCount": "1"
//     }
// }
// u[f5a516808f8976c33939133293d598ce3bca4e8d:3]={
//     "ObjectInfo": {
//         "ID": "f5a516808f8976c33939133293d598ce3bca4e8d:3",
//         "IsEscaped": true,
//         "ModTime": "6",
//         "RefCount": "3"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/crossrealm_test.foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "de578df5354e812a94f64e23eed06e9705d2b4ef",
//             "ObjectID": "f5a516808f8976c33939133293d598ce3bca4e8d:4"
//         }
//     }
// }
