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
	somevalue = foo{"h2"}
	print(somevalue)
}

// Output:
// (struct{("h2" string)} gno.land/r/storage_test.foo)

// Storage:
// gno.land/r/storage_test: -5

// Realm:
// switchrealm["gno.land/r/storage_test"]
// c[334570344df8e77213d2886cb407c37de6da36c1:5](240)={
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
//         "ID": "334570344df8e77213d2886cb407c37de6da36c1:5",
//         "ModTime": "0",
//         "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:2",
//         "RefCount": "1"
//     }
// }
// u[334570344df8e77213d2886cb407c37de6da36c1:2](0)=
//     @@ -54,8 +54,8 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "38c5985f800f03536705a87106ec69de0ef17653",
//     -                "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:3"
//     +                "Hash": "8aec9f0c1cca83fe81c8cd719c5b3edec450a945",
//     +                "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:5"
//                  }
//              },
//              {
// d[334570344df8e77213d2886cb407c37de6da36c1:3](-245)
