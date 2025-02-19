// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import (
	crossrealm "gno.land/r/demo/tests/crossrealm/slice"
)

var SS []crossrealm.XYZ

var f = func(s []crossrealm.XYZ) {
	SS = s
}

func main() {
	// XXX, even elem of the array is not real,
	// but the array is already attached to
	// external realm, so the elem has no
	// chance to attach to this realm.
	// so it is valid.
	crossrealm.GetSlice4(f)
	println(".")
}

// Output:
// .

// Realm:
// switchrealm["gno.land/r/crossrealm_test"]
// switchrealm["gno.land/r/demo/tests/crossrealm/slice"]
// c[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29]={
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
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29",
//         "ModTime": "0",
//         "OwnerID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:5",
//         "RefCount": "1"
//     }
// }
// u[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:5]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "9aac0359e72ff36096f642fc4c7d2e134f28993d",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:20"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "ada47c453e82855b67936f208c6cc6a5c8698dad",
//                 "ObjectID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "ff34c6500d0daee58e4ac9899c0e7ecc131db63e:5",
//         "IsEscaped": true,
//         "ModTime": "28",
//         "RefCount": "2"
//     }
// }
// d[ff34c6500d0daee58e4ac9899c0e7ecc131db63e:7]
// switchrealm["gno.land/r/crossrealm_test"]
