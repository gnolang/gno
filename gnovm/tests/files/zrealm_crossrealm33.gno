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
// c[r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29]={
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
//         "ID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29",
//         "ModTime": "0",
//         "OwnerID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:5",
//         "RefCount": "1"
//     }
// }
// u[r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:5]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "c0a8ddd6cc93886c615d83da96acf19574ac4ad5",
//                 "ObjectID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:20"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/slice.XYZ"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "3eba3f8937283fd19372768c68fa1cfffa65e546",
//                 "ObjectID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:29"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:5",
//         "IsEscaped": true,
//         "ModTime": "28",
//         "OwnerID": "r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:2",
//         "RefCount": "2"
//     }
// }
// d[r:ff34c6500d0daee58e4ac9899c0e7ecc131db63e:7]
// switchrealm["gno.land/r/crossrealm_test"]
