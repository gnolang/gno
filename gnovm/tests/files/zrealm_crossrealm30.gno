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
// u[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:15]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "c341f36102ff3f8e2f54df678a394e385459dadd",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:16"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "dd8ac5d5b483b2e2ae7fb7fff494674c82d99367",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:17"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:15",
//         "IsEscaped": true,
//         "ModTime": "3",
//         "OwnerID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:2",
//         "RefCount": "2"
//     }
// }
