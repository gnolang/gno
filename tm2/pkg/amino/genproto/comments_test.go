package genproto

import (
	"path"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/assert"
)

// message comment
type TestMessageName struct {
	// field comment 1
	FieldName1 string
	// field comment 2
	FieldName2 []uint64
}

// message comment 2
type TestMessageName2 struct {
	// another field comment
	FieldName string
}

func TestComments(t *testing.T) {
	pkg := amino.RegisterPackage(
		amino.NewPackage(
			"github.com/gnolang/gno/tm2/pkg/amino/genproto",
			"amino_test",
			amino.GetCallersDirname(),
		).WithTypes(
			&TestMessageName{},
			&TestMessageName2{},
		// Add comments from this same source file.
		).WithComments(path.Join(amino.GetCallersDirname(), "comments_test.go")))

	p3c := NewP3Context()
	p3c.RegisterPackage(pkg)
	p3c.ValidateBasic()
	p3doc := p3c.GenerateProto3SchemaForTypes(pkg, pkg.ReflectTypes()...)
	proto3Schema := p3doc.Print()
	assert.Equal(t, `syntax = "proto3";
package amino_test;

option go_package = "github.com/gnolang/gno/tm2/pkg/amino/genproto/pb";

// messages
// message comment
message TestMessageName {
	// field comment 1
	string field_name1 = 1 [json_name = "FieldName1"];
	// field comment 2
	repeated uint64 field_name2 = 2 [json_name = "FieldName2"];
}

// message comment 2
message TestMessageName2 {
	// another field comment
	string field_name = 1 [json_name = "FieldName"];
}`, proto3Schema)
}
