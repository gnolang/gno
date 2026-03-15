package genproto2

import (
	"os"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/amino/tests"
)

func TestGenerateProtobuf3(t *testing.T) {
	cdc := amino.NewCodec()
	cdc.RegisterPackage(tests.Package)
	cdc.Seal()

	ctx := NewP3Context2(cdc)

	// Use all reflect types from the package (includes repr types).
	rtz := tests.Package.ReflectTypes()

	src, err := ctx.GenerateProtobuf3ForTypes("tests", rtz...)
	if err != nil {
		t.Fatalf("GenerateProtobuf3ForTypes: %v", err)
	}

	// Write to stdout for inspection.
	t.Logf("Generated source:\n%s", src)

	// Write to file for compilation check.
	outFile := "../tests/pb3_gen.go"
	if err := os.WriteFile(outFile, []byte(src), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	t.Logf("Wrote %s", outFile)
}
