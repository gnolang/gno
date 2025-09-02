// PKGPATH: gno.land/r/crossrealm
package crossrealm

import (
	"gno.land/r/tests/crossrealm_b"
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
// file{ package crossrealm; import crossrealm_b gno.land/r/tests/crossrealm_b; type Struct (const-type gno.land/r/crossrealm.Struct); var s<!~VPBlock(2,1)> *(typeval{gno.land/r/crossrealm.Struct}); func init.2() { s<~VPBlock(3,1)> = &((const-type gno.land/r/crossrealm.Struct){A<VPField(0,0,A)>: (const (100 int))}); (const (ref(gno.land/r/tests/crossrealm_b) package{})).SetObject((const (undefined)), func func(){ (const (println func(...interface {})))(&((const (ref(gno.land/r/crossrealm) package{})).s.A)) }) }; func main(cur (const-type .uverse.realm)) { (const (ref(gno.land/r/tests/crossrealm_b) package{})).SetObject((const (undefined)), (const (123 int))); s<~VPBlock(3,1)>.A = (const (123 int)); (const (println func(...interface {})))(s<~VPBlock(3,1)>) } }

// Realm:
// finalizerealm["gno.land/r/tests/crossrealm_b"]
// u[3003d3e51bf2f146cfba386aa7161085b561ec2d:13](-98)=
//     @@ -2,20 +2,15 @@
//          "ObjectInfo": {
//              "ID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:13",
//              "LastObjectSize": "320",
//     -        "ModTime": "15",
//     +        "ModTime": "16",
//              "OwnerID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:2",
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
//     -            "Hash": "c11818db8b64a40be20dbb0a2eda8270f61a0bfc",
//     -            "ObjectID": "3003d3e51bf2f146cfba386aa7161085b561ec2d:16"
//     +            "@type": "/gno.PrimitiveType",
//     +            "value": "32"
//              }
//          }
//      }
// d[3003d3e51bf2f146cfba386aa7161085b561ec2d:16](-298)
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
