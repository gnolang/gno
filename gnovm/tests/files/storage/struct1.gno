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
// c[34f997a9ca158338c03cfc00686d77220a6cf62f:7](240)={
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
//         "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:7",
//         "LastObjectSize": "240",
//         "ModTime": "0",
//         "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3",
//         "RefCount": "1"
//     }
// }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3",
//              "LastObjectSize": "338",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "RefCount": "1"
//          },
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "2054070e999df8e7cb3d1f1b92c7019aa996206d",
//     -            "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:4"
//     +            "Hash": "38c8d7d3b202f3ff95c28b39fcfcf9f04114c5ec",
//     +            "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:7"
//              }
//          }
//      }
// d[34f997a9ca158338c03cfc00686d77220a6cf62f:4](-214)
