// PKGPATH: gno.land/r/storage
package storage

type foo struct {
	field string
}

var somevalue foo

func init() {
	somevalue.field = "hello"
}
func main() {
	somevalue = foo{"h2"}
	print(somevalue)
}

// Output:
// (struct{("h2" string)} gno.land/r/storage.foo)

// Storage:
// gno.land/r/storage: -8

// Realm:
// finalizerealm["gno.land/r/storage"]
// c[34f997a9ca158338c03cfc00686d77220a6cf62f:8:0](244)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "h2"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:8:0",
//         "LastObjectSize": "244",
//         "ModTime": "0",
//         "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3:0",
//         "RefCount": "1"
//     }
// }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:3:0](0)=
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "4a1c813a492f89565afeed70044e9f9a0a130a96",
//     -            "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:4:0"
//     +            "Hash": "2de0dd8f89cf4b940f1a07efe2a2a64a7603189c",
//     +            "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:8:0"
//              }
//          }
//      }
// d[34f997a9ca158338c03cfc00686d77220a6cf62f:4:0](-252)
