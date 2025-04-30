// PKGPATH: gno.land/r/storage_test
package storage_test

type foo struct {
	Field string
}

var somevalue *foo

func main() {
	crossing()
	somevalue = &foo{}
	somevalue.Field = "hello"
	print(somevalue)
}

// Output:
// &(struct{("hello" string)} gno.land/r/storage_test.foo)



// Storage:
// gno.land/r/storage_test: 720



// Realm:
// finalizerealm["gno.land/r/storage_test"]
// c[334570344df8e77213d2886cb407c37de6da36c1:7](243)={
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
//         "ID": "334570344df8e77213d2886cb407c37de6da36c1:7",
//         "LastObjectSize": "243",
//         "ModTime": "0",
//         "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:6",
//         "RefCount": "1"
//     }
// }
// c[334570344df8e77213d2886cb407c37de6da36c1:6](343)={
//     "ObjectInfo": {
//         "ID": "334570344df8e77213d2886cb407c37de6da36c1:6",
//         "LastObjectSize": "343",
//         "ModTime": "0",
//         "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:3",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/storage_test.foo"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "5cbe3dd2cb1a51f75adf8c948a6a19cf6d5d8117",
//             "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:7"
//         }
//     }
// }
// u[334570344df8e77213d2886cb407c37de6da36c1:3](134)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "334570344df8e77213d2886cb407c37de6da36c1:3",
//              "LastObjectSize": "259",
//     -        "ModTime": "0",
//     +        "ModTime": "5",
//              "OwnerID": "334570344df8e77213d2886cb407c37de6da36c1:2",
//              "RefCount": "1"
//          },
//     @@ -13,6 +13,16 @@
//                      "@type": "/gno.RefType",
//                      "ID": "gno.land/r/storage_test.foo"
//                  }
//     +        },
//     +        "V": {
//     +            "@type": "/gno.PointerValue",
//     +            "Base": {
//     +                "@type": "/gno.RefValue",
//     +                "Hash": "a56dc44469c2076358385ec4b936d85d5768be0e",
//     +                "ObjectID": "334570344df8e77213d2886cb407c37de6da36c1:6"
//     +            },
//     +            "Index": "0",
//     +            "TV": null
//              }
//          }
//      }
