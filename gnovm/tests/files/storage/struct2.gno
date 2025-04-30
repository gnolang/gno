// PKGPATH: gno.land/r/storage_test
package storage_test

type foo struct {
	field string
}

var somevalue foo

func init() {
	somevalue.field = "hello"
}
func main() {
	crossing()
	somevalue = foo{"h2"}
	print(somevalue)
}

// Output:
// (struct{("h2" string)} gno.land/r/storage_test.foo)



// Storage:
// gno.land/r/storage_test: -8



// Realm:
// finalizerealm["gno.land/r/storage_test"]
// c[334570344df8e77213d2886cb407c37de6da36c1:8](240)={
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
//         "ID": "334570344df8e77213d2886cb407c37de6da36c1:8",
//         "LastObjectSize": "240",
//         "ModTime": "0",
//         "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:3",
//         "RefCount": "1"
//     }
// }
// u[334570344df8e77213d2886cb407c37de6da36c1:3](0)=
//     @@ -13,8 +13,8 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "6500a8358940dce4df7ba94c49720676eddc10d1",
//     -            "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:4"
//     +            "Hash": "b30184135e8063eb1e79ec3ed6f1be623913bb37",
//     +            "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:8"
//              }
//          }
//      }
// d[334570344df8e77213d2886cb407c37de6da36c1:4](-248)
