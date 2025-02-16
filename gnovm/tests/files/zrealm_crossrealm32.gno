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
// c[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:30]={
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
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:30",
//         "ModTime": "0",
//         "OwnerID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29",
//         "RefCount": "1"
//     }
// }
// c[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:31]={
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
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:31",
//         "ModTime": "0",
//         "OwnerID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29",
//         "RefCount": "1"
//     }
// }
// c[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "de4d4b9346995296fbd8210badf8bab1c111e60c",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:30"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "1570e95bdb3313b5a0656b56b9bba96c9c7704aa",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:31"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29",
//         "ModTime": "0",
//         "OwnerID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:2",
//         "RefCount": "1"
//     }
// }
// d[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:18]
// d[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:19]
// switchrealm["gno.land/r/crossrealm_test"]
// u[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "de4d4b9346995296fbd8210badf8bab1c111e60c",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:30"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "1570e95bdb3313b5a0656b56b9bba96c9c7704aa",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:31"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29",
//         "IsEscaped": true,
//         "ModTime": "3",
//         "RefCount": "2"
//     }
// }
