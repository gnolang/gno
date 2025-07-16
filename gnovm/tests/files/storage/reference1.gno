// PKGPATH: gno.land/r/storage
package storage

type foo struct {
	Field string
}

var somevalue *foo

func main() {
	somevalue = &foo{}
	somevalue.Field = "hello"
	print(somevalue)
}

// Output:
// &(struct{("hello" string)} gno.land/r/storage.foo)

// Storage:
// gno.land/r/storage: 728

// Realm:
// finalizerealm["gno.land/r/storage"]
// c[34f997a9ca158338c03cfc00686d77220a6cf62f:7:0](247)={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "hello"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:7:0",
//         "LastObjectSize": "247",
//         "ModTime": "0",
//         "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:6:0",
//         "RefCount": "1"
//     }
// }
// c[34f997a9ca158338c03cfc00686d77220a6cf62f:6:0](344)={
//     "ObjectInfo": {
//         "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:6:0",
//         "LastObjectSize": "344",
//         "ModTime": "0",
//         "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3:0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/storage.foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "f76e53dc38ec03eef4dc96726cb99bf1d0ea0b7c",
//             "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:7:0"
//         }
//     }
// }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:3:0](137)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3:0",
//              "LastObjectSize": "258",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:2:0",
//              "RefCount": "1"
//          },
//     @@ -13,6 +13,16 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/storage.foo"
//                  }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "60758d53ea2e600c67a5fdeeed4a07da574aa2a5",
//     +                "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:6:0"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
