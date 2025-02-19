// PKGPATH: gno.land/r/crossrealm_test
package crossrealm_test

import crossrealm "gno.land/r/demo/tests/crossrealm/map"

var m1 map[string]int

func main() {
	m1 = crossrealm.GetMap()
	println(m1)
	crossrealm.ModifyMap()
	println(m1)
}

// Output:
// map{("a" string):(1 int)}
// map{("a" string):(2 int)}

// Realm:
// switchrealm["gno.land/r/demo/tests/crossrealm/map"]
// switchrealm["gno.land/r/demo/tests/crossrealm/map"]
// switchrealm["gno.land/r/crossrealm_test"]
// u[5a06306c27263d4b9ae8f2f30e82f8ef014042cd:6]={
//     "List": {
//         "List": [
//             {
//                 "Key": {
//                     "T": {
//                         "@type": "/gno.PrimitiveType",
//                         "value": "16"
//                     },
//                     "V": {
//                         "@type": "/gno.StringValue",
//                         "value": "a"
//                     }
//                 },
//                 "Value": {
//                     "N": "AgAAAAAAAAA=",
//                     "T": {
//                         "@type": "/gno.PrimitiveType",
//                         "value": "32"
//                     }
//                 }
//             }
//         ]
//     },
//     "ObjectInfo": {
//         "ID": "5a06306c27263d4b9ae8f2f30e82f8ef014042cd:6",
//         "IsEscaped": true,
//         "ModTime": "3",
//         "RefCount": "2"
//     }
// }
