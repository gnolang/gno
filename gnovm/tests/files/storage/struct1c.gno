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
// c[34f997a9ca158338c03cfc00686d77220a6cf62f:7:0](248)={
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
//         "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:7:0",
//         "LastObjectSize": "248",
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
//     -            "Hash": "2b93a533a682793571373703d9da1693b9a6d3e3",
//     -            "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:4:0"
//     +            "Hash": "e3ca7a65a9ac7d2dbb8c729a91ba5d6f572fc5f6",
//     +            "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:7:0"
//              }
//          }
//      }
// d[34f997a9ca158338c03cfc00686d77220a6cf62f:4:0](-247)
