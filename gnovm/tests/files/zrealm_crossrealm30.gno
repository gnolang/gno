// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

var SS []crossrealm.XYZ

func main() {
	SS = crossrealm.GetSlice()
	println(".")
}

// Output:
// .

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm/slice"]
// switchrealm["gno.land/r/crossrealm_test"]
// u[r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:15]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "ee0b0db2758a60075732af94600fce1f3f0b7b45",
//                 "ObjectID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:16"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "b5783602afd1983ea7e420869bda47f048f4293b",
//                 "ObjectID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:17"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:15",
//         "IsEscaped": true,
//         "ModTime": "3",
//         "OwnerID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:2",
//         "RefCount": "2"
//     }
// }
