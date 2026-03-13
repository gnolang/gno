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
// c[34f997a9ca158338c03cfc00686d77220a6cf62f:8](240)={
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
//         "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:8",
//         "LastObjectSize": "240",
//         "ModTime": "0",
//         "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3",
//         "RefCount": "1"
//     }
// }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:3](0)=
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "cb71b821cefa1bb05605966f33f2d53001f59237",
//     -            "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:4"
//     +            "Hash": "fb5d0d403090106fdcdd5778d01d657e038c00c8",
//     +            "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:8"
//              }
//          }
//      }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "IsEscaped": true,
//              "LastObjectSize": "813",
//     -        "ModTime": "7",
//     +        "ModTime": "8",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -78,7 +78,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "5122c40762af4d008a4f359648d1613fe38cd2d6",
//     +                "Hash": "18031f5617b07d749e3b1a719ac912fa72e0f5f9",
//                      "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3"
//                  }
//              },
// d[34f997a9ca158338c03cfc00686d77220a6cf62f:4](-248)
