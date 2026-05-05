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
// gno.land/r/storage: 715

// Realm:
// finalizerealm["gno.land/r/storage"]
// c[04f997a9ca158338c03cfc00686d77220a6cf62f:7](243)={
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
//         "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:7",
//         "LastObjectSize": "243",
//         "ModTime": "0",
//         "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:6",
//         "RefCount": "1"
//     }
// }
// c[04f997a9ca158338c03cfc00686d77220a6cf62f:6](338)={
//     "ObjectInfo": {
//         "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:6",
//         "LastObjectSize": "338",
//         "ModTime": "0",
//         "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:3",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/storage.foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "3a413fbf36d1191bf438c588d9f345b8ecf129b2",
//             "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:7"
//         }
//     }
// }
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:3](134)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:3",
//              "LastObjectSize": "254",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "OwnerID": "04f997a9ca158338c03cfc00686d77220a6cf62f:2",
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
//     +                "Hash": "4875f42d817632f78fe3ca47bfabeafbfbf22ee5",
//     +                "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:6"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
