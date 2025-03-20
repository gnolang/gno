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
// &(struct{(ref(r:5965ae851c02ab677bc8394b408535c1db9b2635:11) gno.land/r/demo/tests/crossrealm/struct.A),(ref(r:5965ae851c02ab677bc8394b408535c1db9b2635:12) gno.land/r/demo/tests/crossrealm/struct.B),(struct{("d" string)} gno.land/r/demo/tests/crossrealm/struct.D)} gno.land/r/demo/tests/crossrealm/struct.C)

// Realm:
// switchrealm["gno.land/r/crossrealm_test"]
// u[r:5965ae851c02ab677bc8394b408535c1db9b2635:9]={
//     "ObjectInfo": {
//         "ID": "r:5965ae851c02ab677bc8394b408535c1db9b2635:9",
//         "IsEscaped": true,
//         "ModTime": "3",
//         "OwnerID": "r:5965ae851c02ab677bc8394b408535c1db9b2635:2",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests/crossrealm/struct.C"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "ObjectID": "r:5965ae851c02ab677bc8394b408535c1db9b2635:10"
//         }
//     }
// }
// switchrealm["gno.land/r/demo/tests/crossrealm/struct"]
// c[r:5965ae851c02ab677bc8394b408535c1db9b2635:14]={
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
//         "ID": "r:5965ae851c02ab677bc8394b408535c1db9b2635:14",
//         "ModTime": "0",
//         "OwnerID": "r:5965ae851c02ab677bc8394b408535c1db9b2635:10",
//         "RefCount": "1"
//     }
// }
// u[r:5965ae851c02ab677bc8394b408535c1db9b2635:10]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/struct.A"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "59054b6f0596aa85e80e67908d5746a45e675d01",
//                 "ObjectID": "r:5965ae851c02ab677bc8394b408535c1db9b2635:11"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/struct.B"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "0219929092adc485d1e42ed61ed28e79b5c004e4",
//                 "ObjectID": "r:5965ae851c02ab677bc8394b408535c1db9b2635:12"
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.RefType",
//                 "ID": "gno.land/r/demo/tests/crossrealm/struct.D"
//             },
//             "V": {
//                 "@type": "/gno.RefValue",
//                 "Hash": "3e39a5556887b3e762792a17e0a39be93feca14f",
//                 "ObjectID": "r:5965ae851c02ab677bc8394b408535c1db9b2635:14"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "r:5965ae851c02ab677bc8394b408535c1db9b2635:10",
//         "ModTime": "13",
//         "OwnerID": "r:5965ae851c02ab677bc8394b408535c1db9b2635:9",
//         "RefCount": "1"
//     }
// }
// d[r:5965ae851c02ab677bc8394b408535c1db9b2635:13]
// switchrealm["gno.land/r/crossrealm_test"]
