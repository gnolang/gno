// PKGPATH: gno.land/r/storage
package storage

type foo struct {
	field string
}

var somevalue foo = foo{"hello"}

func main() {
	somevalue = foo{"hello2"}
	println(somevalue)

}

// Output:
// (struct{("hello2" string)} gno.land/r/storage.foo)

// Storage:
// gno.land/r/storage: 6

// Realm:
// finalizerealm["gno.land/r/storage"]
// c[04f997a9ca158338c03cfc00686d77220a6cf62f:7](244)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "hello2"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:7",
//         "LastObjectSize": "244",
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
//     -            "Hash": "8c902f00b69bdbfd67032c4db79852ededc606f0",
//     -            "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:4"
//     +            "Hash": "959ac8265241c987e9b584a312c37a620809bcef",
//     +            "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:7"
//              }
//          }
//      }
// d[04f997a9ca158338c03cfc00686d77220a6cf62f:4](-243)
