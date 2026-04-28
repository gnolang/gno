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
// c[04f997a9ca158338c03cfc00686d77220a6cf62f:8](240)={
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
//         "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:8",
//         "LastObjectSize": "240",
//         "ModTime": "0",
//         "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:3",
//         "RefCount": "1"
//     }
// }
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:3](0)=
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "d0f1ba7ee33fbbbf66a1c14967ed5c48b66e1bca",
//     -            "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:4"
//     +            "Hash": "8973bec246829f2dd6dde20fa21f72b45af4b956",
//     +            "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:8"
//              }
//          }
//      }
// d[04f997a9ca158338c03cfc00686d77220a6cf62f:4](-248)
