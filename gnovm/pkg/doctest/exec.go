package doctest

import (
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

var execResult map[string]string

func executeCodeBlock(c CodeBlock) error {
	if c.T != "go" || c.T != "gno" {
		return fmt.Errorf("unsupported language: %s", c.T)
	}

	m := gno.NewMachine("runMD", nil)

	// TODO: need to static analysis the code block
	pkgContent := c.Content
	parsedCode := gno.MustParseFile(fmt.Sprintf("%d.%s", c.Index, c.T), pkgContent)

	m.RunFiles(parsedCode)
	m.RunMain()
	res := m.PopValue().V.String()

	execResult[fmt.Sprintf("%d.%s", c.Index, c.T)] = res

	return nil
}
