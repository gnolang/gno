package genproto

import (
	"reflect"
	"testing"

	sm1 "github.com/gnolang/gno/tm2/pkg/amino/genproto/example/submodule"
	"github.com/jaekwon/testify/assert"
)

func TestBasic(t *testing.T) {
	t.Parallel()

	p3c := NewP3Context()
	p3c.RegisterPackage(sm1.Package)
	p3doc := P3Doc{PackageName: "test"}
	obj := sm1.StructSM{}
	p3message := p3c.GenerateProto3MessagePartial(&p3doc, reflect.TypeOf(obj))
	assert.Equal(t, p3message.Print(), `message StructSM {
	sint64 field_a = 1;
	string field_b = 2;
	submodule2.StructSM2 field_c = 3;
}
`)

	assert.Equal(t, p3doc.Print(), `syntax = "proto3";
package test;

// imports
import "github.com/gnolang/gno/tm2/pkg/amino/genproto/example/submodule2/submodule2.proto";`)

	p3doc = p3c.GenerateProto3SchemaForTypes(sm1.Package, reflect.TypeOf(obj))
	assert.Equal(t, p3doc.Print(), `syntax = "proto3";
package submodule;

option go_package = "github.com/gnolang/gno/tm2/pkg/amino/genproto/example/submodule/pb";

// imports
import "github.com/gnolang/gno/tm2/pkg/amino/genproto/example/submodule2/submodule2.proto";

// messages
message StructSM {
	sint64 field_a = 1;
	string field_b = 2;
	submodule2.StructSM2 field_c = 3;
}`)
}
