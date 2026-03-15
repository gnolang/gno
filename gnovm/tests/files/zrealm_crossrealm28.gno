// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/vm/crossrealm_b"
)

type Struct struct {
	A int
}

var s *Struct

func init() {
	s = &Struct{A: 100}
	crossrealm_b.SetObject(cross, func() { println(&s.A) })
}

func main(cur realm) {
	//s := &Struct{A: 100}
	crossrealm_b.SetObject(cross, 123)
	s.A = 123
	println(s)
}

// Output:
// &(struct{(123 int)} gno.land/r/crossrealm.Struct)

// Preprocessed:
// file{ package crossrealm; import crossrealm_b gno.land/r/tests/vm/crossrealm_b; type Struct (const-type gno.land/r/crossrealm.Struct); var s<!~VPBlock(2,1)> *(typeval{gno.land/r/crossrealm.Struct}); func init.2() { s<~VPBlock(3,1)> = &((const-type gno.land/r/crossrealm.Struct){A<VPField(0,0,A)>: (const (100 int))}); (const (ref(gno.land/r/tests/vm/crossrealm_b) package{})).SetObject((const (undefined)), func func(){ (const (println func(...interface {})))(&((const (ref(gno.land/r/crossrealm) package{})).s.A)) }) }; func main(cur (const-type .uverse.realm)) { (const (ref(gno.land/r/tests/vm/crossrealm_b) package{})).SetObject((const (undefined)), (const (123 int))); s<~VPBlock(3,1)>.A = (const (123 int)); (const (println func(...interface {})))(s<~VPBlock(3,1)>) } }

// Realm:
// finalizerealm["gno.land/r/tests/vm/crossrealm_b"]
// u[46279d1e03ecb38b84822f9caf584bd16913470d:13](-98)=
//     @@ -2,20 +2,15 @@
//          "ObjectInfo": {
//              "ID": "46279d1e03ecb38b84822f9caf584bd16913470d:13",
//              "LastObjectSize": "320",
//     -        "ModTime": "18",
//     +        "ModTime": "19",
//              "OwnerID": "46279d1e03ecb38b84822f9caf584bd16913470d:2",
//              "RefCount": "1"
//          },
//          "Value": {
//     +        "N": "ewAAAAAAAAA=",
//              "T": {
//     -            "@type": "/gno.FuncType",
//     -            "Params": null,
//     -            "Results": null
//     -        },
//     -        "V": {
//     -            "@type": "/gno.RefValue",
//     -            "Hash": "3fa519e4001c6f5839834b18d33f97c7f94139c5",
//     -            "ObjectID": "46279d1e03ecb38b84822f9caf584bd16913470d:19"
//     +            "@type": "/gno.PrimitiveType",
//     +            "value": "32"
//              }
//          }
//      }
// u[46279d1e03ecb38b84822f9caf584bd16913470d:2](0)=
//     @@ -412,7 +412,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "9162b92ff86a750a143bab0c753ded6c21fdac04",
//     +                "Hash": "a443f74b6b43fa241851c4b3805435dd0925fe14",
//                      "ObjectID": "46279d1e03ecb38b84822f9caf584bd16913470d:13"
//                  }
//              },
// d[46279d1e03ecb38b84822f9caf584bd16913470d:19](-298)
// finalizerealm["gno.land/r/crossrealm"]
// u[dc8307f7bbe97fdb07b2d0fa6907b8d67517c02a:8](5)=
//     @@ -1,7 +1,7 @@
//      {
//          "Fields": [
//              {
//     -            "N": "ZAAAAAAAAAA=",
//     +            "N": "ewAAAAAAAAA=",
//                  "T": {
//                      "@type": "/gno.PrimitiveType",
//                      "value": "32"
//     @@ -11,7 +11,7 @@
//          "ObjectInfo": {
//              "ID": "dc8307f7bbe97fdb07b2d0fa6907b8d67517c02a:8",
//              "LastObjectSize": "214",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "dc8307f7bbe97fdb07b2d0fa6907b8d67517c02a:7",
//              "RefCount": "1"
//          }
// u[dc8307f7bbe97fdb07b2d0fa6907b8d67517c02a:7](5)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "dc8307f7bbe97fdb07b2d0fa6907b8d67517c02a:7",
//              "LastObjectSize": "344",
//     -        "ModTime": "0",
//     +        "ModTime": "8",
//              "OwnerID": "dc8307f7bbe97fdb07b2d0fa6907b8d67517c02a:3",
//              "RefCount": "1"
//          },
//     @@ -13,7 +13,7 @@
//              },
//              "V": {
//                  "@type": "/gno.RefValue",
//     -            "Hash": "770a64d97b1129dcf25db6ea700cc34d027f3046",
//     +            "Hash": "4bd0ab551a62aa1e024d6bd92fc2e68080a56820",
//                  "ObjectID": "dc8307f7bbe97fdb07b2d0fa6907b8d67517c02a:8"
//              }
//          }
// u[dc8307f7bbe97fdb07b2d0fa6907b8d67517c02a:3](0)=
//     @@ -2,7 +2,7 @@
//          "ObjectInfo": {
//              "ID": "dc8307f7bbe97fdb07b2d0fa6907b8d67517c02a:3",
//              "LastObjectSize": "394",
//     -        "ModTime": "6",
//     +        "ModTime": "8",
//              "OwnerID": "dc8307f7bbe97fdb07b2d0fa6907b8d67517c02a:2",
//              "RefCount": "1"
//          },
//     @@ -18,7 +18,7 @@
//                  "@type": "/gno.PointerValue",
//                  "Base": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "64260dffa338d52040d93c48724d1d1d6e6d13cc",
//     +                "Hash": "cf4b6d2cfd5fd71d159a3bfd36ab486ebdb6dab3",
//                      "ObjectID": "dc8307f7bbe97fdb07b2d0fa6907b8d67517c02a:7"
//                  },
//                  "Index": "0",
// u[dc8307f7bbe97fdb07b2d0fa6907b8d67517c02a:2](0)=
//     @@ -78,7 +78,7 @@
//                  },
//                  "V": {
//                      "@type": "/gno.RefValue",
//     -                "Hash": "ba5242d61b2c3424f5fa634d9e2a9c388785dfe2",
//     +                "Hash": "485bf560e328fd580e62cf45f02f910c46f8f06a",
//                      "ObjectID": "dc8307f7bbe97fdb07b2d0fa6907b8d67517c02a:3"
//                  }
//              },
