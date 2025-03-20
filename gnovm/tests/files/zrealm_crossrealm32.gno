// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

var SS []crossrealm.XYZ

func main() {
	SS = crossrealm.GetSlice3()
	println(".")
}

// Output:
// .

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm/slice"]
// c[r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:30]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "3"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:30",
//         "ModTime": "0",
//         "OwnerID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29",
//         "RefCount": "1"
//     }
// }
// c[r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:31]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "0"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:31",
//         "ModTime": "0",
//         "OwnerID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29",
//         "RefCount": "1"
//     }
// }
// c[r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "295b6935876d54ed5a9f807f90e006d29dda2394",
//                 "ObjectID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:30"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "bc851a49bc21d4dfe80cbb9d59d72adc24dfaa1e",
//                 "ObjectID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:31"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29",
//         "ModTime": "0",
//         "OwnerID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:2",
//         "RefCount": "1"
//     }
// }
// d[r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:18]
// d[r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:19]
// switchrealm["gno.land/r/crossrealm_test"]
// u[r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "295b6935876d54ed5a9f807f90e006d29dda2394",
//                 "ObjectID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:30"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "bc851a49bc21d4dfe80cbb9d59d72adc24dfaa1e",
//                 "ObjectID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:31"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29",
//         "IsEscaped": true,
//         "ModTime": "3",
//         "RefCount": "2"
//     }
// }
