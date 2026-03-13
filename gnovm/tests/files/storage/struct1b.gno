// PKGPATH: gno.land/r/storage
package storage

type foo struct {
	field string
}

var somevalue foo

func main() {
	somevalue = foo{"hello2"}
	println(somevalue)

}

// Output:
// (struct{("hello2" string)} gno.land/r/storage.foo)

// Storage:
// gno.land/r/storage: 40

// Realm:
// finalizerealm["gno.land/r/storage"]
// c[34f997a9ca158338c03cfc00686d77220a6cf62f:7](244)={
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
//         "ID": "34f997a9ca158338c03cfc00686d77220a6cf62f:7",
//         "LastObjectSize": "244",
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
//     +            "Hash": "828319b9f2bc74534d09894634deb113649ba6b1",
//     +            "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:7"
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
//     -                "Hash": "fc43278e097d5020fccbc0b4b8f05afa73707e2d",
//     +                "Hash": "9c3cf4875a2a4741ff184f097b36514a6bacdff9",
//                      "ObjectID": "34f997a9ca158338c03cfc00686d77220a6cf62f:3"
//                  }
//              },
// d[34f997a9ca158338c03cfc00686d77220a6cf62f:4](-214)
