package params

import (
	"fmt"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKeeper_GetSetValues(t *testing.T) {
	t.Run("empty keeper", func(t *testing.T) {
		var (
			env         = setupTestEnv(t, "params_test")
			ctx, keeper = env.ctx, env.keeper
		)

		suffixes := []string{
			stringSuffix,
			boolSuffix,
			uint64Suffix,
			int64Suffix,
			bytesSuffix,
		}

		for _, suffix := range suffixes {
			assert.False(t, keeper.Has(ctx, fmt.Sprintf("param%s", suffix)))
		}
	})

	t.Run("valid key-value pairs", func(t *testing.T) {
		var (
			env         = setupTestEnv(t, "params_test")
			ctx, keeper = env.ctx, env.keeper
		)

		testTable := []struct {
			suffix string
			value  any
		}{
			{
				stringSuffix,
				"foo",
			},
			{
				boolSuffix,
				true,
			},
			{
				uint64Suffix,
				uint64(42),
			},
			{
				int64Suffix,
				int64(-1337),
			},
			{
				bytesSuffix,
				[]byte("hello world!"),
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.suffix, func(t *testing.T) {
				key := fmt.Sprintf("param%s", testCase.suffix)

				switch v := testCase.value.(type) {
				case string:
					require.NoError(t, keeper.SetString(ctx, key, v))
				case bool:
					require.NoError(t, keeper.SetBool(ctx, key, v))
				case uint64:
					require.NoError(t, keeper.SetUint64(ctx, key, v))
				case int64:
					require.NoError(t, keeper.SetInt64(ctx, key, v))
				case []byte:
					require.NoError(t, keeper.SetBytes(ctx, key, v))
				default:
					t.Fatalf("unsupported type for key %s", key)
				}

				// Check if the key was set successfully
				assert.True(t, keeper.Has(ctx, key))

				switch v := testCase.value.(type) {
				case string:
					readValue, err := keeper.GetString(ctx, key)
					require.NoError(t, err)

					assert.Equal(t, v, readValue)
				case bool:
					readValue, err := keeper.GetBool(ctx, key)
					require.NoError(t, err)

					assert.Equal(t, v, readValue)
				case uint64:
					readValue, err := keeper.GetUint64(ctx, key)
					require.NoError(t, err)

					assert.Equal(t, v, readValue)
				case int64:
					readValue, err := keeper.GetInt64(ctx, key)
					require.NoError(t, err)

					assert.Equal(t, v, readValue)
				case []byte:
					readValue, err := keeper.GetBytes(ctx, key)
					require.NoError(t, err)

					assert.Equal(t, v, readValue)
				default:
					t.Fatalf("unsupported type for key %s", key)
				}
			})
		}
	})

	t.Run("invalid keys, set", func(t *testing.T) {
		var (
			env         = setupTestEnv(t, "params_test")
			ctx, keeper = env.ctx, env.keeper
		)

		testTable := []struct {
			name  string
			key   string
			value any
		}{
			{
				"value mismatch, string",
				fmt.Sprintf("param%s", stringSuffix),
				uint64(42), // value mismatch
			},
			{
				"value mismatch, uint64",
				fmt.Sprintf("param%s", uint64Suffix),
				"foo", // value mismatch
			},
			{
				"value mismatch, bool",
				fmt.Sprintf("param%s", boolSuffix),
				int64(42), // value mismatch
			},
			{
				"value mismatch, int64",
				fmt.Sprintf("param%s", int64Suffix),
				[]byte{}, // value mismatch
			},
			{
				"value mismatch, bytes",
				fmt.Sprintf("param%s", bytesSuffix),
				uint64(42), // value mismatch
			},
			{
				"no suffix",
				"param",
				false,
			},
			{
				"empty key",
				uint64Suffix, // just the suffix (empty key)
				uint64(42),
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				var err error

				switch v := testCase.value.(type) {
				case string:
					err = keeper.SetString(ctx, testCase.key, v)
				case bool:
					err = keeper.SetBool(ctx, testCase.key, v)
				case uint64:
					err = keeper.SetUint64(ctx, testCase.key, v)
				case int64:
					err = keeper.SetInt64(ctx, testCase.key, v)
				case []byte:
					err = keeper.SetBytes(ctx, testCase.key, v)
				default:
					t.Fatalf("unsupported type for key %s", testCase.key)
				}

				assert.ErrorContains(t, err, "key should be like")
			})
		}
	})

	t.Run("invalid keys, get", func(t *testing.T) {
		var (
			env         = setupTestEnv(t, "params_test")
			ctx, keeper = env.ctx, env.keeper
		)

		testTable := []struct {
			name           string
			expectedSuffix string
			key            string
		}{
			{
				"key mismatch, string",
				stringSuffix,
				fmt.Sprintf("param%s", int64Suffix),
			},
			{
				"key mismatch, uint64",
				uint64Suffix,
				fmt.Sprintf("param%s", stringSuffix),
			},
			{
				"key mismatch, bool",
				boolSuffix,
				fmt.Sprintf("param%s", uint64Suffix),
			},
			{
				"key mismatch, int64",
				int64Suffix,
				fmt.Sprintf("param%s", uint64Suffix),
			},
			{
				"key mismatch, bytes",
				bytesSuffix,
				fmt.Sprintf("param%s", stringSuffix),
			},
		}

		for _, testCase := range testTable {
			t.Run(testCase.name, func(t *testing.T) {
				var err error

				switch testCase.expectedSuffix {
				case stringSuffix:
					_, err = keeper.GetString(ctx, testCase.key)
				case boolSuffix:
					_, err = keeper.GetBool(ctx, testCase.key)
				case uint64Suffix:
					_, err = keeper.GetUint64(ctx, testCase.key)
				case int64Suffix:
					_, err = keeper.GetInt64(ctx, testCase.key)
				case bytesSuffix:
					_, err = keeper.GetBytes(ctx, testCase.key)
				}

				assert.ErrorContains(t, err, "key should be like")
			})
		}
	})
}

func TestGetAndSetParams(t *testing.T) {
	t.Parallel()

	type params struct {
		p1 int
		p2 string
	}

	env := setupTestEnv(t, "get_set_params_test")
	ctx := env.ctx
	keeper := env.keeper
	// SetParams
	a := params{p1: 1, p2: "a"}
	err := keeper.SetParams(ctx, ModuleName, a)
	require.NoError(t, err)

	// GetParams
	a1 := params{}
	_, err1 := keeper.GetParams(ctx, ModuleName, &a1)
	require.NoError(t, err1)
	require.True(t, amino.DeepEqual(a, a1), "a and a1 should equal")
}
