// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm/struct"

var s *crossrealm.C

func cb(c *crossrealm.C) {
	s = c
}
func main() {
	crossrealm.GetStruct(cb)
	println(s)
}

// Output:
// &(struct{(ref(5965ae851c02ab677bc8394b408535c1db9b2635:11) gno.land/r/demo/tests/crossrealm/struct.A),(ref(5965ae851c02ab677bc8394b408535c1db9b2635:12) gno.land/r/demo/tests/crossrealm/struct.B),(struct{("d" string)} gno.land/r/demo/tests/crossrealm/struct.D)} gno.land/r/demo/tests/crossrealm/struct.C)

// Realm:
// switchrealm["gno.land/r/crossrealm_test"]
// u[5965ae851c02ab677bc8394b408535c1db9b2635:9]={
//     "ObjectInfo": {
//         "ID": "5965ae851c02ab677bc8394b408535c1db9b2635:9",
//         "IsEscaped": true,
//         "ModTime": "3",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests/crossrealm/struct.C"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "ObjectID": "5965ae851c02ab677bc8394b408535c1db9b2635:10"
//         }
//     }
// }
// switchrealm["gno.land/r/demo/tests/crossrealm/struct"]
// c[5965ae851c02ab677bc8394b408535c1db9b2635:14]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "d"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "5965ae851c02ab677bc8394b408535c1db9b2635:14",
//         "ModTime": "0",
//         "OwnerID": "5965ae851c02ab677bc8394b408535c1db9b2635:10",
//         "RefCount": "1"
//     }
// }
// u[5965ae851c02ab677bc8394b408535c1db9b2635:10]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/struct.A"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "77afd6587e56836f288520b275bebc4911b7ef57",
//                 "ObjectID": "5965ae851c02ab677bc8394b408535c1db9b2635:11"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/struct.B"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "c4e78b61af7cd5557452e3d22d20d1c3b1de4b41",
//                 "ObjectID": "5965ae851c02ab677bc8394b408535c1db9b2635:12"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/struct.D"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "70a58641e15bf62bf3d1d95ce2e8b2048286f42f",
//                 "ObjectID": "5965ae851c02ab677bc8394b408535c1db9b2635:14"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "5965ae851c02ab677bc8394b408535c1db9b2635:10",
//         "ModTime": "13",
//         "OwnerID": "5965ae851c02ab677bc8394b408535c1db9b2635:9",
//         "RefCount": "1"
//     }
// }
// u[5965ae851c02ab677bc8394b408535c1db9b2635:9]={
//     "ObjectInfo": {
//         "ID": "5965ae851c02ab677bc8394b408535c1db9b2635:9",
//         "IsEscaped": true,
//         "ModTime": "14",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests/crossrealm/struct.C"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "4cf8da80d5dec249a7b325ad5608c4e8187f60ac",
//             "ObjectID": "5965ae851c02ab677bc8394b408535c1db9b2635:10"
//         }
//     }
// }
// d[5965ae851c02ab677bc8394b408535c1db9b2635:13]
// switchrealm["gno.land/r/crossrealm_test"]
