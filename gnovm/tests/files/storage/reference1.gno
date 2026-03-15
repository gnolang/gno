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
// gno.land/r/storage: 720

// Realm:
// finalizerealm["gno.land/r/storage"]
// c[34f997a9ca158338c03cfc00686d77220a6cf62f:7](243)={
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
//         "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:7",
//         "LastObjectSize": "243",
//         "ModTime": "0",
//         "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:6",
//         "RefCount": "1"
//     }
// }
// c[34f997a9ca158338c03cfc00686d77220a6cf62f:6](338)={
//     "ObjectInfo": {
//         "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:6",
//         "LastObjectSize": "338",
//         "ModTime": "0",
//         "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/storage.foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "0ed9f1cbe4144466c3fe41bdd44ed7b36bdbd795",
//             "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:7"
//         }
//     }
// }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:3](134)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3",
//              "LastObjectSize": "254",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "OwnerID": "34f997a9ca158338c03cfc00686d77220a6cf62f:2",
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
//     +                "Hash": "52407cedb3387fbe6f0b5a4de2ca198cb409b523",
//     +                "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:6"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
// u[34f997a9ca158338c03cfc00686d77220a6cf62f:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "IsEscaped": true,
//              "LastObjectSize": "673",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -78,7 +78,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "ad42aecaecf4300655f56b5546337cc192679c01",
//     +                "Hash": "7a15a7242cf439fa0308e3d3c7530a6358d20419",
//                      "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3"
//                  }
//              },
