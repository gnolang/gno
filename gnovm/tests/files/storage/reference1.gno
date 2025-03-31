// PKGPATH: gno.land/r/storage_test
package storage_test

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
// &(struct{("hello" string)} gno.land/r/storage_test.foo)

// Storage:
// gno.land/r/storage_test: 715

// Realm:
// switchrealm["gno.land/r/storage_test"]
// c[334570344df8e77213d2886cb407c37de6da36c1:5](243)={
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
//         "ID": "334570344df8e77213d2886cb407c37de6da36c1:5",
//         "ModTime": "0",
//         "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:4",
//         "RefCount": "1"
//     }
// }
// c[334570344df8e77213d2886cb407c37de6da36c1:4](343)={
//     "ObjectInfo": {
//         "ID": "334570344df8e77213d2886cb407c37de6da36c1:4",
//         "ModTime": "0",
//         "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:2",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/storage_test.foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "9b1f745c041648964ed2f5c223e8a55861d96b77",
//             "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:5"
//         }
//     }
// }
// u[334570344df8e77213d2886cb407c37de6da36c1:2](129)=
//     @@ -54,6 +54,16 @@
//                          "@type": "/gno.RefType",
//                          "ID": "gno.land/r/storage_test.foo"
//                      }
//     +            },
//     +            "V": {
//     +                "@type": "/gno.PointerValue",
//     +                "Base": {
//     +                    "@type": "/gno.RefValue",
//     +                    "Hash": "e9f9d33424e77a6b09211b54c3e2dbb6c4e74ecb",
//     +                    "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:4"
//     +                },
//     +                "Index": "0",
//     +                "TV": null
//                  }
//              },
//              {
