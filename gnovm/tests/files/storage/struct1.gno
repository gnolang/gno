// PKGPATH: gno.land/r/storage
package storage

type foo struct {
	field string
}

var somevalue foo

func main() {
	somevalue.field = "hello"
	somevalue = foo{"h2"}
	print(somevalue)
}

// Output:
// (struct{("h2" string)} gno.land/r/storage.foo)

// Storage:
// gno.land/r/storage: 31

// Realm:
// finalizerealm["gno.land/r/storage"]
// c[34f997a9ca158338c03cfc00686d77220a6cf62f:7:0](244)={
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
//         "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:7:0",
//         "LastObjectSize": "244",
//         "ModTime": "0",
//         "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3:0",
//         "RefCount": "1"
//     }
// }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:3:0](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3:0",
//              "LastObjectSize": "344",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:2:0",
//              "RefCount": "1"
//          },
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "6c3dd1fae4e3ae890fe3a9c261caaca5bcaa4799",
//     -            "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:4:0"
//     +            "Hash": "d9322513a5eb031f58c5135c1c67e7a6a0bb5aad",
//     +            "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:7:0"
//              }
//          }
//      }
// d[34f997a9ca158338c03cfc00686d77220a6cf62f:4:0](-218)
