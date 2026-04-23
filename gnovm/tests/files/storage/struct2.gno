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
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:2](0)=
//     @@ -4,7 +4,7 @@
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "IsEscaped": true,
//              "LastObjectSize": "729",
//     -        "ModTime": "7",
//     +        "ModTime": "8",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "5bda02c49358fcee26ac0216e11955619a59e04c",
//     +                "Hash": "17553827dbc01fa82f5ffba59adc0241397c373e",
//                      "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:3"
//                  }
//              },
// d[04f997a9ca158338c03cfc00686d77220a6cf62f:4](-248)
