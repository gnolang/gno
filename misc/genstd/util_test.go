package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_pkgNameFromPath(t *testing.T) {
	tt := []struct {
		input, result string
	}{
		{"math", "go_math"},
		{"crypto/sha256", "go_crypto_sha256"},
		{"github.com/import/path", "ext_github_com_import_path"},
		// consecutive unsupported characters => _
		{"kebab----------case", "go_kebab_case"},

		{"github.com/gnolang/gno/misc/test", "repo_misc_test"},
		{"github.com/gnolang/gno/tm2/pkg/crypto", "tm2_crypto"},
		{"github.com/gnolang/gno/gnovm/test", "vm_test"},
		{"github.com/gnolang/gno/gnovm/stdlibs/std", "libs_std"},
		{"github.com/gnolang/gno/gnovm/tests/stdlibs/std", "testlibs_std"},
	}
	for i, tv := range tt {
		t.Run(fmt.Sprintf("n%d", i+1), func(t *testing.T) {
			assert.Equal(t, tv.result, pkgNameFromPath(tv.input))
		})
	}
}
