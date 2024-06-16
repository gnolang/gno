package doctest

import (
	"bytes"
	"fmt"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// executeCodeBLock executes a code block using gnoVM and caching the result.
func executeCodeBlock(c CodeBlock) (string, error) {
	if c.T != "go" {
		return "", fmt.Errorf("unsupported language: %s", c.T)
	}

	m := gno.NewMachine("main", nil)

	// capture output
	var output bytes.Buffer
	m.Output = &output

	pkgContent := c.Content

	// throw panic when parsing fails
	parsedCode := gno.MustParseFile(fmt.Sprintf("%d.%s", c.Index, c.T), pkgContent)

	m.RunFiles(parsedCode)
	m.RunMain()

	result := output.String()
	return result, nil
}
