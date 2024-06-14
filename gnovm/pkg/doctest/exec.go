package doctest

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
)

// Global variable to store execution results
var codeResults map[string]string

func init() {
	codeResults = make(map[string]string)
}

// executeCodeBLock executes a code block using gnoVM and caching the result.
func executeCodeBlock(c CodeBlock) error {
	if c.T != "go" && c.T != "gno" {
		return fmt.Errorf("unsupported language: %s", c.T)
	} else {
		c.T = "gno"
	}

	m := gno.NewMachine("runMD", nil)

	// TODO: need to static analysis the code block
	pkgContent := c.Content
	parsedCode := gno.MustParseFile(fmt.Sprintf("%d.%s", c.Index, c.T), pkgContent)

	m.RunFiles(parsedCode)
	m.RunMain()
	res := m.PopValue().V.String()


	// ignore the whitespace in the source code
	key := generateCacheKey([]byte(strings.ReplaceAll(c.Content, " ", "")))
	codeResults[key] = res

	return nil
}

// generateCacheKey creates a SHA-256 hah of the source code to be used as a cache key
// to avoid re-executing the same code block.
func generateCacheKey(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}