// PKGPATH: gno.land/r/storage_test
package storage_test

type foo struct {
	field string
}

var somevalue foo

func main() {
	somevalue = foo{"hello2"}
	println(somevalue)

}

// Output:
// (struct{("hello2" string)} gno.land/r/storage_test.foo)

// Storage:
// gno.land/r/storage_test: 35

// Realm:
// finalizerealm["gno.land/r/storage_test"]
// c[334570344df8e77213d2886cb407c37de6da36c1:7](244)={
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
//         "ID": "334570344df8e77213d2886cb407c37de6da36c1:7",
//         "LastObjectSize": "244",
//         "ModTime": "0",
//         "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:3",
//         "RefCount": "1"
//     }
// }
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
//     +            "Hash": "5e1a696207973bdfd7aae0d2651d1ed803a324f5",
//     +            "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:7"
//              }
//          }
//      }
// d[334570344df8e77213d2886cb407c37de6da36c1:4](-214)
