package privval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateBasic(t *testing.T) {
	t.Parallel()

	t.Run("sign state file path is not set", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultPrivValidatorConfig()
		cfg.SignState = ""

		assert.ErrorIs(t, cfg.ValidateBasic(), errInvalidSignStatePath)
	})

	t.Run("local signer file path is not set", func(t *testing.T) {
		t.Parallel()

		cfg := DefaultPrivValidatorConfig()
		cfg.LocalSigner = ""

		assert.ErrorIs(t, cfg.ValidateBasic(), errInvalidLocalSignerPath)
	})

	// TODO: Test authorizedKeys
	// TODO: Test root dir
}

// func TestConfigFilePath(t *testing.T) {
// 	t.Parallel()
//
// 	const testPath = "test_path"
//
// 	t.Run("set DefaultPrivValidatorConfig path", func(t *testing.T) {
// 		t.Parallel()
//
// 		cfg := DefaultPrivValidatorConfig(testPath)
//
// 		assert.Equal(t, cfg.SignState, filepath.Join(testPath, defaultSignStateName))
// 		assert.Equal(t, cfg.LocalSigner, filepath.Join(testPath, defaultLocalSignerName))
// 	})
//
// 	t.Run("set TestPrivValidatorConfig path", func(t *testing.T) {
// 		t.Parallel()
//
// 		cfg := TestPrivValidatorConfig(testPath)
//
// 		assert.Equal(t, cfg.SignState, filepath.Join(testPath, defaultSignStateName))
// 		assert.Equal(t, cfg.LocalSigner, filepath.Join(testPath, defaultLocalSignerName))
// 	})
// }
