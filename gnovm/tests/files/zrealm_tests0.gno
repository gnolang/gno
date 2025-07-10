// PKGPATH: gno.land/r/demo/tests0
package tests0

import (
	"gno.land/r/demo/tests"
	"gno.land/r/demo/tests_foo"
)

func init() {
	tests_foo.AddFooStringer("one")
	tests_foo.AddFooStringer("two")
}

func main(cur realm) {
	tests_foo.AddFooStringer("three")
	println(tests.Render(""))
	println("end")
}

// Output:
// 0: &FooStringer{one}
// 1: &FooStringer{two}
// 2: &FooStringer{three}
//
// end

// Realm:
// finalizerealm["gno.land/r/demo/tests"]
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:59:0]={
//     "Fields": [
//         {
//             "T": {
//                 "@type": "/gno.PrimitiveType",
//                 "value": "16"
//             },
//             "V": {
//                 "@type": "/gno.StringValue",
//                 "value": "three"
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:59:0",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:58:0",
//         "RefCount": "1"
//     }
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:58:0]={
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:58:0",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:57:0",
//         "RefCount": "1"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests_foo.FooStringer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "119f670aea32410f7ede937b0e52d803a17a244d",
//             "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:59:0"
//         }
//     }
// }
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:57:0]={
//     "Data": null,
//     "List": [
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "925f5ec55589710eaf57eac5873436074369da41",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:52:0"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "4458b93fd33392c7c4398b66398325e83deb503d",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:55:0"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         },
//         {
//             "T": {
//                 "@type": "/gno.PointerType",
//                 "Elt": {
//                     "@type": "/gno.RefType",
//                     "ID": "gno.land/r/demo/tests_foo.FooStringer"
//                 }
//             },
//             "V": {
//                 "@type": "/gno.PointerValue",
//                 "Base": {
//                     "@type": "/gno.RefValue",
//                     "Hash": "5f909b68cbb4f28cdd866493f20198b7fc3eef3f",
//                     "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:58:0"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:57:0",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:6:0",
//         "RefCount": "1"
//     }
// }
// u[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:6:0]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:6:0",
//     -        "ModTime": "53",
//     +        "ModTime": "56",
//              "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2:0",
//              "RefCount": "1"
//          },
//     @@ -18,11 +18,11 @@
//                  "@type": "/gno.SliceValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "491fe4b1d3ca109210a46c2c517420f63de64b1b",
//     -                "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:54:0"
//     +                "Hash": "98f8f1d257b6f6d9688d6797d5dd9cad69c7d066",
//     +                "ObjectID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:57:0"
//                  },
//     -            "Length": "2",
//     -            "Maxcap": "2",
//     +            "Length": "3",
//     +            "Maxcap": "3",
//                  "Offset": "0"
//              }
//          }
// u[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:52:0]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:52:0",
//     -        "ModTime": "54",
//     +        "ModTime": "57",
//              "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:51:0",
//              "RefCount": "1"
//          },
// u[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:55:0]=
//     @@ -1,7 +1,7 @@
//      {
//          "ObjectInfo": {
//              "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:55:0",
//     -        "ModTime": "0",
//     +        "ModTime": "57",
//              "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:54:0",
//              "RefCount": "1"
//          },
// d[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:54:0]
// finalizerealm["gno.land/r/demo/tests"]
// finalizerealm["gno.land/r/demo/tests"]
// finalizerealm["gno.land/r/demo/tests"]
// finalizerealm["gno.land/r/demo/tests0"]
