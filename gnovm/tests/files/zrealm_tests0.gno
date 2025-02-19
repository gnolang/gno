// PKGPATH: gno.land/r/demo/tests_test
package tests_test

import (
	"gno.land/r/demo/tests"
	"gno.land/r/demo/tests_foo"
)

func init() {
	tests_foo.AddFooStringer("one")
	tests_foo.AddFooStringer("two")
}

func main() {
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
// switchrealm["gno.land/r/demo/tests"]
// c[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:13]={
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
//                     "Escaped": true,
//                     "ObjectID": "f643af0964f4fc0bfa4d9d204506cc75a00fac2d:3"
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
//                     "Escaped": true,
//                     "ObjectID": "f643af0964f4fc0bfa4d9d204506cc75a00fac2d:5"
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
//                     "Escaped": true,
//                     "ObjectID": "f643af0964f4fc0bfa4d9d204506cc75a00fac2d:7"
//                 },
//                 "Index": "0",
//                 "TV": null
//             }
//         }
//     ],
//     "ObjectInfo": {
//         "ID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:13",
//         "ModTime": "0",
//         "OwnerID": "0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:2",
//         "RefCount": "1"
//     }
// }
// u[f643af0964f4fc0bfa4d9d204506cc75a00fac2d:3]={
//     "ObjectInfo": {
//         "ID": "f643af0964f4fc0bfa4d9d204506cc75a00fac2d:3",
//         "IsEscaped": true,
//         "ModTime": "13",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests_foo.FooStringer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "45f556b68be863359e2bce915ca9bea54cb433ee",
//             "ObjectID": "f643af0964f4fc0bfa4d9d204506cc75a00fac2d:4"
//         }
//     }
// }
// u[f643af0964f4fc0bfa4d9d204506cc75a00fac2d:5]={
//     "ObjectInfo": {
//         "ID": "f643af0964f4fc0bfa4d9d204506cc75a00fac2d:5",
//         "IsEscaped": true,
//         "ModTime": "13",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests_foo.FooStringer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "852145bd1f1b33fa43aa9e6c1c41ccfc31e952c2",
//             "ObjectID": "f643af0964f4fc0bfa4d9d204506cc75a00fac2d:6"
//         }
//     }
// }
// u[f643af0964f4fc0bfa4d9d204506cc75a00fac2d:7]={
//     "ObjectInfo": {
//         "ID": "f643af0964f4fc0bfa4d9d204506cc75a00fac2d:7",
//         "IsEscaped": true,
//         "ModTime": "13",
//         "RefCount": "2"
//     },
//     "Value": {
//         "T": {
//             "@type": "/gno.RefType",
//             "ID": "gno.land/r/demo/tests_foo.FooStringer"
//         },
//         "V": {
//             "@type": "/gno.RefValue",
//             "Hash": "db60709913153489a5694aa8139398ca99c75f9d",
//             "ObjectID": "f643af0964f4fc0bfa4d9d204506cc75a00fac2d:8"
//         }
//     }
// }
// d[0ffe7732b4d549b4cf9ec18bd68641cd2c75ad0a:12]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests_foo"]
// switchrealm["gno.land/r/demo/tests"]
// switchrealm["gno.land/r/demo/tests_test"]
