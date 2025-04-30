// PKGPATH: gno.land/r/storage_test
package storage_test

type foo struct {
	field string
}

var somevalue foo

func main() {
	crossing()
	somevalue.field = "hello"
	somevalue = foo{"h2"}
	print(somevalue)
}

// Output:
// (struct{("h2" string)} gno.land/r/storage_test.foo)



// Storage:
// gno.land/r/storage_test: 31



// Realm:
// finalizerealm["gno.land/r/storage_test"]
// c[334570344df8e77213d2886cb407c37de6da36c1:7](240)={
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
//         "ID": "334570344df8e77213d2886cb407c37de6da36c1:7",
//         "LastObjectSize": "240",
//         "ModTime": "0",
//         "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:3",
//         "RefCount": "1"
//     }
// }
// u[334570344df8e77213d2886cb407c37de6da36c1:4](32)=
//     @@ -4,14 +4,18 @@
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "16"
//     +            },
//     +            "V": {
//     +                "@type": "/gno.StringValue",
//     +                "value": "hello"
//                  }
//              }
//          ],
//          "ObjectInfo": {
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:4",
//              "LastObjectSize": "214",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:3",
//     -        "RefCount": "1"
//     +        "RefCount": "0"
//          }
//      }
// u[334570344df8e77213d2886cb407c37de6da36c1:3](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:3",
//              "LastObjectSize": "343",
//     -        "ModTime": "0",
//     +        "ModTime": "6",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:2",
//              "RefCount": "1"
//          },
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "836c831a8ece251b218a8df34b0b77a3ef820fed",
//     -            "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:4"
//     +            "Hash": "46662456cae62b0448c34590c7cc63d535be0d02",
//     +            "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:7"
//              }
//          }
//      }
// d[334570344df8e77213d2886cb407c37de6da36c1:4](-246)
