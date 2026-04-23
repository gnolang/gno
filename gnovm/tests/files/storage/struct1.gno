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
// c[04f997a9ca158338c03cfc00686d77220a6cf62f:7](240)={
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
//         "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:7",
//         "LastObjectSize": "240",
//         "ModTime": "0",
//         "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:3",
//         "RefCount": "1"
//     }
// }
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:3",
//              "LastObjectSize": "338",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "RefCount": "1"
//          },
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "fd590cdbcf4922e4763aa37b29bf43887cabf06f",
//     -            "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:4"
//     +            "Hash": "bbca14191d9ff8f32abaf30b17c4d70d3258a5a0",
//     +            "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:7"
//              }
//          }
//      }
// d[04f997a9ca158338c03cfc00686d77220a6cf62f:4](-214)
