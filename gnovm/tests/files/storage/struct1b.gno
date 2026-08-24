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
//     -            "Hash": "fd590cdbcf4922e4763aa37b29bf43887cabf06f",
//     -            "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:4"
//     +            "Hash": "959ac8265241c987e9b584a312c37a620809bcef",
//     +            "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:7"
//              }
//          }
//      }
// u[04f997a9ca158338c03cfc00686d77220a6cf62f:2](5)=
//     @@ -4,7 +4,7 @@
//              "ID": "04f997a9ca158338c03cfc00686d77220a6cf62f:2",
//              "IsEscaped": true,
//              "LastObjectSize": "589",
//     -        "ModTime": "0",
//     +        "ModTime": "7",
//              "RefCount": "2"
//          },
//          "Parent": null,
//     @@ -46,7 +46,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "805531c45161b2d90e00d128c2b8b82755bd4e15",
//     +                "Hash": "40ca4e03b583eeae2c211143febbd041bc6e0fed",
//                      "ObjectID": "04f997a9ca158338c03cfc00686d77220a6cf62f:3"
//                  }
//              },
// d[04f997a9ca158338c03cfc00686d77220a6cf62f:4](-214)
